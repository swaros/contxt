// Copyright (c) 2023 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// # Licensed under the MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package runner

import (
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/abiosoft/readline"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/ctxshell"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
)

const (
	ModusInit = 1
	ModusRun  = 2
	ModusTask = 3
	ModusIdle = 4
)

type CtxShell struct {
	cmdSession     *CmdExecutorImpl
	shell          *ctxshell.Cshell
	Modus          int
	MaxTasks       int
	CollectedTasks []string
	SynMutex       sync.Mutex
	LabelForeColor string
	LabelBackColor string
}

func shellRunner(c *CmdExecutorImpl) {
	// run the context shell
	shell := ctxshell.NewCshell()
	shellHandler := &CtxShell{
		cmdSession:     c,
		shell:          shell,
		Modus:          ModusInit,
		MaxTasks:       0,
		LabelForeColor: ctxout.ForeBlue,
		LabelBackColor: ctxout.BackWhite,
	}

	// add cobra commands to the shell, so they can be used there too.
	// first we need to define the exceptions
	// we do not want to have in the menu
	shell.SetIgnoreCobraCmd("completion", "interactive")
	// afterwards we can add the commands by injecting the root command
	shell.SetCobraRootCommand(c.session.Cobra.RootCmd)

	// rename the exit command to quit
	shell.SetExitCmdStr("quit")

	// some of the commands are not working well async, because they
	// are switching between workspaces. so we have to disable async
	// for them
	shell.SetNeverAsyncCmd("workspace")

	// capture ctrl+z and do nothing, so we will not send to the background
	shell.AddKeyBinding(readline.CharCtrlZ, func() bool { return false })

	// add native exit command
	exitCmd := ctxshell.NewNativeCmd("exit", "exit the shell", func(args []string) error {
		systools.Exit(0)
		return nil
	})
	// completer function for the exit command
	exitCmd.SetCompleterFunc(func(line string) []string {
		return []string{"exit"}
	})
	shell.AddNativeCmd(exitCmd)

	// add task clean command
	cleanTasksCmd := ctxshell.NewNativeCmd("taskreset", "resets all tasks", func(args []string) error {
		return tasks.NewGlobalWatchman().ResetAllTasksIfPossible()
	})
	cleanTasksCmd.SetCompleterFunc(func(line string) []string {
		return []string{"taskreset"}
	})
	shell.AddNativeCmd(cleanTasksCmd)

	// add task stop command
	stoppAllCmd := ctxshell.NewNativeCmd("stoptasks", "stop all the running processes", func(args []string) error {
		ctxshell.NewCshell().SetStopOutput(true)
		tasks.NewGlobalWatchman().StopAllTasks(func(target string, time int, succeed bool) {
			if succeed {
				ctxout.PrintLn(ctxout.NewMOWrap(), "stopped process: ", target)
			} else {
				ctxout.PrintLn(ctxout.NewMOWrap(), "failed to stop processes: ", target)
			}
		})
		ctxshell.NewCshell().SetStopOutput(false)
		tasks.HandleAllMyPid(func(pid int) error {
			ctxout.PrintLn(ctxout.NewMOWrap(), "killing process: ", pid)
			if proc, err := os.FindProcess(pid); err == nil {
				if err := proc.Kill(); err != nil {
					return err
				} else {
					ctxout.PrintLn(ctxout.NewMOWrap(), "killed process: ", pid)
				}
			} else {
				ctxout.PrintLn(ctxout.NewMOWrap(), "failed to kill process: ", pid)
				return err
			}
			return nil
		})

		return nil
	})
	stoppAllCmd.SetCompleterFunc(func(line string) []string {
		return []string{"stoptasks"}
	})
	shell.AddNativeCmd(stoppAllCmd)

	// set the prompt handler
	shell.SetPromptFunc(func(reason int) string {

		label := ""
		// in idle or init mode we display the current directory
		if shellHandler.Modus == ModusIdle || shellHandler.Modus == ModusInit {
			if dir, err := dirhandle.Current(); err == nil {
				label += dir
			} else {
				label += err.Error()
			}
		}

		label = shellHandler.autoSetLabel(label)
		// depends runtime.GOOS we have oure own prompt handler
		// becaue on windows we have not all the features we have on linux
		if runtime.GOOS == "windows" {
			return shellHandler.windowsPrompt(reason, label)
		} else {
			return shellHandler.linuxPrompt(reason, label)
		}
	})
	// rebind the the session output handler
	// so any output will be handled by the shell
	c.session.OutPutHdnl = shell
	// start the shell
	shell.SetAsyncCobraExec(true).
		SetAsyncNativeCmd(true).
		UpdatePromptEnabled(true).
		UpdatePromptPeriod(1 * time.Second).
		Run()
}

// adds an additonial task label to the prompt and increases the prompt update period
// if there are running tasks.
// if no tasks are running, the prompt update period will be set to 1 second.
// also it sets the mode to ModusTask if any tasks are running.
func (cs *CtxShell) autoSetLabel(label string) string {
	watchers := tasks.ListWatcherInstances()
	taskCount := 0
	// this is only saying, we have some watchers found. it is not saying, that there are any tasks running
	// for this we have to check the watchers one by one
	if len(watchers) > 0 {
		taskBar := ""
		for _, watcher := range watchers {
			watchMan := tasks.GetWatcherInstance(watcher)
			if watchMan != nil {
				allRunnungs := watchMan.GetAllRunningTasks()
				if len(allRunnungs) > 0 {
					taskCount += len(allRunnungs)
					// add the tasks to the collected tasks they are not already in
					for _, task := range allRunnungs {
						if !systools.StringInSlice(task, cs.CollectedTasks) {
							cs.CollectedTasks = append(cs.CollectedTasks, task)
						}
					}
				}
				// build the taskbar
				runningChar := cs.getABraillCharByTime()
				doneChar := "✓"
				for _, task := range cs.CollectedTasks {
					if watchMan.TaskRunning(task) {
						taskBar += ctxout.ForeWhite + runningChar
					} else {
						taskBar += ctxout.ForeBlack + doneChar
					}
				}
			}
		}
		// do we have any tasks running?
		if taskCount > 0 {
			cs.shell.UpdatePromptPeriod(100 * time.Millisecond)
			label += taskBar
			cs.LabelForeColor = ctxout.ForeWhite
			cs.LabelBackColor = ctxout.BackDarkGrey
			cs.Modus = ModusTask
			label = ctxout.ToString(ctxout.NewMOWrap(), label)
			return cs.fitStringLen(label, ctxout.ToString("t", taskCount))
		} else {
			// no tasks running, so reset the all the task related stuff
			cs.shell.UpdatePromptPeriod(1 * time.Second)
			cs.LabelForeColor = ctxout.ForeBlue
			cs.LabelBackColor = ctxout.BackWhite
			cs.MaxTasks = 0
			cs.Modus = ModusIdle
			cs.CollectedTasks = []string{}
		}
	}
	return cs.fitStringLen(label, "")

}

// fit the string length to the half of the terminal width, if an fallback is set, it will be returned
func (cs *CtxShell) fitStringLen(label string, fallBack string) string {
	w, _, err := systools.GetStdOutTermSize()
	if err != nil {
		w = 80
	}
	maxLen := w / 2
	if systools.StrLen(systools.NoEscapeSequences(label)) > maxLen {
		// if fallback is set, we return it
		if fallBack != "" {
			return fallBack
		}
		// if no fallback is set, we reduce the label
		label = systools.StringSubLeft(label, maxLen)

	}
	return label
}

// a braille char
// depending on the milliseconds of the current time
func (cs *CtxShell) getABraillCharByTime() string {
	braillTableString := "⠄⠆⠇⠋⠙⠸⠰⠠⠐⠈"
	braillTable := []rune(braillTableString)
	millis := time.Now().UnixNano() / int64(time.Millisecond)
	index := int(millis % int64(len(braillTable)))
	return string(braillTable[index])
}

// returns the prompt for windows.
// here we are limited to the ascii chars we can use.
func (cs *CtxShell) windowsPrompt(reason int, label string) string {

	return ctxout.ToString(
		ctxout.NewMOWrap(),
		ctxout.ForeBlue,
		label,
		" ",
		ctxout.ForeCyan,
		"› ",
		ctxout.CleanTag,
	)
}

// returns the prompt for linux.
func (cs *CtxShell) linuxPrompt(reason int, label string) string {

	// display the current time in the prompt
	// this is just for testing

	timeNowAsString := time.Now().Format("15:04:05")

	if reason == ctxshell.UpdateByNotify {
		if found, msg := cs.shell.GetCurrentMessage(); found {
			timeNowAsString = msg.GetMsg()
		}

	}

	return ctxout.ToString(
		ctxout.NewMOWrap(),
		ctxout.BackBlue,
		ctxout.ForeWhite,
		"",
		timeNowAsString,
		" ",
		cs.LabelForeColor,
		cs.LabelBackColor,
		label,
		ctxout.BackBlue,
		ctxout.ForeWhite,
		"ctx<f:black>:</><f:blue></> ",
	)
}
