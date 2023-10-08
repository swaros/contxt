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
	"runtime"
	"strings"
	"sync"
	"time"

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

	// some of the commands are not working well async, because they
	// are switching between workspaces. so we have to disable async
	// for them
	shell.SetNeverAsyncCmd("workspace")

	// add native exit command
	exitCmd := ctxshell.NewNativeCmd("exit", "exit the shell", func(args []string) error {
		ctxout.PrintLn(ctxout.ForeBlue, "bye bye", ctxout.CleanTag)
		systools.Exit(0)
		return nil
	})
	exitCmd.SetCompleterFunc(func(line string) []string {
		return []string{"exit"}
	})
	shell.AddNativeCmd(exitCmd)

	/* disable this for now
	// add native commands to menu
	// this one is for testing only
	demoCmd := ctxshell.NewNativeCmd("demo", "demo command", func(args []string) error {
		c.Println("demo command executed:", strings.Join(args, " "))
		for i := 0; i < 5000; i++ {
			time.Sleep(10 * time.Millisecond)
			c.Println("i do something .. we are in round ", i)
		}
		return nil
	})


	// while developing, you can use this to test the completer
	// and the command itself
	demoCmd.SetCompleterFunc(func(line string) []string {
		return []string{"demo"}
	})

	shell.AddNativeCmd(demoCmd)
	*/

	// set the prompt handler
	shell.SetPromptFunc(func(reason int) string {

		label := ""
		if reason == ctxshell.UpdateByInput {
			if dir, err := dirhandle.Current(); err == nil {
				label = dir
			}
		}
		/*
			switch reason {
			case ctxshell.UpdateByInput:
				label = "input"
			case ctxshell.UpdateBySignal:
				label = "signal"
			case ctxshell.UpdateByPeriod:
				label = "period"

			}
		*/
		label = shellHandler.taskLabel(label)
		// depends runtime.GOOS
		if runtime.GOOS == "windows" {
			return shellHandler.windowsPrompt(reason, label)
		} else {
			return shellHandler.linuxPrompt(reason, label)
		}
	})
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
func (cs *CtxShell) taskLabel(label string) string {
	watchers := tasks.ListWatcherInstances()
	if len(watchers) > 0 {
		cs.shell.UpdatePromptPeriod(100 * time.Millisecond)
		taskCount := 0
		for _, watcher := range watchers {
			watchMan := tasks.GetWatcherInstance(watcher)
			if watchMan != nil {
				allRunnungs := watchMan.GetAllRunningTasks()
				if len(allRunnungs) > 0 {
					taskCount += len(allRunnungs)
				}
			}
		}
		if taskCount > 0 {
			if cs.MaxTasks < taskCount {
				cs.MaxTasks = taskCount
			}
			bChar := cs.getABraillCharByTime()
			taskCountAsString := ctxout.ForeWhite + strings.Repeat(bChar, taskCount)
			taskDoneAsString := ctxout.ForeDarkGrey + strings.Repeat("⠿", cs.MaxTasks-taskCount)
			label += ctxout.BackBlack + taskCountAsString + taskDoneAsString + ctxout.BackWhite
			cs.LabelForeColor = ctxout.ForeBlue
			cs.LabelBackColor = ctxout.BackDarkGrey
		}

	} else {
		cs.shell.UpdatePromptPeriod(1 * time.Second)
		cs.LabelForeColor = ctxout.ForeBlue
		cs.LabelBackColor = ctxout.BackWhite
		cs.MaxTasks = 0
	}
	return ctxout.ToString(label)
}

func (cs *CtxShell) getABraillCharByTime() string {
	// we need to return a braille char
	// depending on the milliseconds of the current time
	braillTableString := "⠄⠆⠇⠋⠙⠸⠰⠠⠐⠈"
	braillTable := []rune(braillTableString)
	millis := time.Now().UnixNano() / int64(time.Millisecond)
	index := int(millis % int64(len(braillTable)))
	return string(braillTable[index])
}

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

func (cs *CtxShell) linuxPrompt(reason int, label string) string {

	// display the current time in the prompt
	// this is just for testing

	timeNowAsString := time.Now().Format("15:04:05")
	// the maximum labe size is half of the terminal width
	w, _, err := systools.GetStdOutTermSize()
	if err != nil {
		w = 80
	}
	maxLen := w / 2
	if len(label) > maxLen {
		label = label[len(label)-maxLen:] + "..."
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
		"ctx<f:yellow>shell:</><f:blue></> ",
	)
}
