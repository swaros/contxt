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
package tasks

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/process"
	"github.com/swaros/contxt/module/systools"
)

func (t *targetExecuter) runAnkCmd(task *configure.Task) (int, error) {
	// nothing to do, get out
	if len(task.Cmd) < 1 {
		return systools.ExitNoCode, nil
	}
	// handle directory change
	curDir, dirError := t.directoryCheckPrep(task)
	if dirError != nil {
		return systools.ExitCmdError, dirError
	}
	ankRunner := NewAnkoRunner()
	defer ankRunner.ClearBuffer()

	// we do not want to print directly to the console
	ankRunner.SetOutputSupression(true)

	cmdFull := t.fullFillVars(strings.Join(task.Cmd, "\n"))
	// set the buffer hook for the anko runner
	// so we get any output from the anko script
	ankRunner.SetBufferHook(func(msg string) {
		t.outPut(task, nil, msg)
		t.setPh("CMD."+task.ID+".LOG.LAST", msg)
		if task.Listener != nil { // do we have listener?
			t.listenerWatch(msg, nil, task) // listener handler
		}
	})

	t.setPh("CMD."+task.ID+".SOURCE", cmdFull) // set or overwrite the last script output for the target
	if task.Listener != nil {                  // do we have listener?
		t.listenerWatch(cmdFull, nil, task) // listener handler
	}

	_, err := ankRunner.RunAnko(cmdFull)
	if err != nil {
		return systools.ExitCmdError, err
	}

	curDir.Popd()
	return systools.ExitOk, nil

}

// targetTaskExecuter is the main function to execute a script line
// it returns the exit code of the executed command
// and a boolean value if the execution was successful
func (t *targetExecuter) targetTaskExecuter(codeLine string, currentTask configure.Task, watchman *Watchman) (int, bool) {
	replacedLine := t.fullFillVars(codeLine) // replace placeholders in the script line
	if currentTask.Options.Displaycmd {
		t.out(MsgTarget{Target: currentTask.ID, Context: "command", Info: replacedLine}) // output the command
	}
	t.setPh("RUN."+currentTask.ID+".CMD.LAST", replacedLine) // set or overwrite the last script command for the target
	t.setPh("RUN.SCRIPT_LINE", replacedLine)                 // set or overwrite the last script command for the target

	runCmd, runArgs := t.commandFallback.GetMainCmd(currentTask.Options) // get the main command and arguments
	t.SetMainCmd(runCmd, runArgs...)                                     // set the main command and arguments

	// keep the current directory
	curDir, dirError := t.directoryCheckPrep(&currentTask)
	if dirError != nil {
		return systools.ExitCmdError, true
	}

	// here we execute the current script line
	execCode, realExitCode, execErr := t.ExecuteScriptLine(
		runCmd,
		runArgs,
		replacedLine,
		func(logLine string, err error) bool { // callback for any logline
			t.setPh("RUN."+currentTask.ID+".LOG.LAST", logLine) // set or overwrite the last script output for the target
			if currentTask.Listener != nil {                    // do we have listener?
				t.listenerWatch(logLine, err, &currentTask) // listener handler
			}

			// The whole output can be ignored by configuration
			// if this is not enabled then we handle all these here
			t.outPut(&currentTask, err, logLine)

			stopReasonFound, message := t.checkReason(currentTask.Stopreasons, logLine, err) // do we found a defined reason to stop execution
			if stopReasonFound {
				if currentTask.Options.Displaycmd {
					t.out(MsgProcess{Target: currentTask.ID, StatusChange: "aborted", Comment: message})
					//t.out(MsgType("stopreason"), MsgReason(message), MsgProcess("aborted"))
				}
				return false
			}
			return true
		}, func(process *os.Process) { // callback if the process started and we got the process id
			pidStr := fmt.Sprintf("%d", process.Pid) // we use them as info for the user only
			t.setPh("RUN.PID", pidStr)
			t.setPh("RUN."+t.target+".PID", pidStr)
			// update watchman with the process infos, if there is an task for this target
			// this should be the case always by any watchman target update, but we check it anyway
			if wtask, found := watchman.GetTask(t.target); found {
				wtask.StartTrackProcess(process)
				wtask.LogCmd(runCmd, runArgs, replacedLine)
				if err := watchman.UpdateTask(t.target, wtask); err != nil {
					t.getLogger().Error("can not update task", err)
					t.out(MsgError(MsgError{Err: err, Reference: codeLine, Target: currentTask.ID}))
				}
			}
			if currentTask.Options.Displaycmd {
				t.out(MsgPid{Pid: process.Pid, Target: currentTask.ID}, MsgProcess{Target: currentTask.ID, StatusChange: "started", Comment: replacedLine})
			}
		})

	curDir.Popd() // restore the current directory
	if currentTask.Options.Displaycmd {
		t.out(MsgProcess{
			Target:       currentTask.ID,
			StatusChange: "done",
			Comment:      fmt.Sprintf("command code: %d, internal code %d", realExitCode, execCode),
		})
	}

	// check execution codes from the executer
	if execErr != nil {
		if currentTask.Options.Displaycmd {
			t.out(MsgError(MsgError{Err: execErr, Reference: codeLine, Target: currentTask.ID}))
		}

	}
	// check execution codes
	switch execCode {
	case systools.ExitByStopReason:
		return systools.ExitByStopReason, true
	case systools.ExitCmdError:
		if currentTask.Options.IgnoreCmdError {
			if currentTask.Stopreasons.Onerror {
				t.out(
					MsgError(MsgError{Err: execErr, Reference: codeLine, Target: currentTask.ID}),
					MsgCommand(codeLine),
					MsgNumber(realExitCode),
				)
				return systools.ExitByStopReason, true
			}

		} else {
			logFields := mimiclog.Fields{
				"processCode": realExitCode,
				"error":       execErr,
			}
			t.getLogger().Error("task exection error", logFields)
			ErrorMsg := errors.New(codeLine + " fails with error: " + execErr.Error())
			t.out(
				MsgError(MsgError{Err: ErrorMsg, Reference: codeLine, Target: currentTask.ID}),
				MsgCommand(codeLine),
				MsgNumber(realExitCode),
			)
			// if we have a hard exit on error we exit the whole process,
			// if the flag 'hardExitOnError' is not set we return the error code
			if t.hardExitOnError {
				systools.Exit(realExitCode) // origin behavior
			}

			// returns the error code
			return systools.ExitCmdError, true
		}
	case systools.ExitOk:
		return systools.ExitOk, false
	}
	return systools.ExitNoCode, true
}

func (t *targetExecuter) fullFillVars(codeLine string) string {
	replacedLine := codeLine
	if t.phHandler != nil {
		replacedLine = t.phHandler.HandlePlaceHolderWithScope(codeLine, t.arguments) // placeholders
	}
	return replacedLine
}

func (t *targetExecuter) outPut(task *configure.Task, err error, output string) {
	if !task.Options.Hideout {
		outStr := output              // hardcoded format for the logoutput iteself
		if task.Options.Stickcursor { // optional set back the cursor to the beginning
			t.out(MsgStickCursor(true)) // trigger the stick cursor
		}

		if err != nil { // if we have an error we print it
			t.out(MsgError(MsgError{Err: err, Reference: outStr, Target: task.ID}))
		}

		t.out(MsgExecOutput(MsgExecOutput{Target: task.ID, Output: outStr})) // prints the output from the running process
		if task.Options.Stickcursor {                                        // cursor stick handling
			t.out(MsgStickCursor(false)) // trigger the stick cursor after output
		}
	}

}

func (t *targetExecuter) directoryCheckPrep(currentTask *configure.Task) (*dirhandle.Popd, error) {
	curDir := dirhandle.Pushd()
	if currentTask.Options.WorkingDir != "" {
		cdErr := os.Chdir(t.phHandler.HandlePlaceHolder(currentTask.Options.WorkingDir)) // change the directory
		if cdErr != nil {
			t.getLogger().Error("can not change directory", cdErr)
			t.out(MsgError(MsgError{Err: cdErr, Reference: currentTask.Options.WorkingDir, Target: currentTask.ID}))
			return nil, cdErr
		}
	} else {
		// if the rootpath exists, we change to this path
		if t.rootPath != "" {
			if er := os.Chdir(t.rootPath); er != nil {
				t.getLogger().Error("can not change directory", er)
				t.out(MsgError(MsgError{Err: er, Reference: t.rootPath, Target: currentTask.ID}))
				return nil, er
			}
		} else {
			// if the rootpath does not exists, we look for tht BASEPATH placeholder
			// and change to this path
			if t.phHandler != nil {
				if basePath := t.phHandler.GetPH("BASEPATH"); basePath != "" {
					if er := os.Chdir(basePath); er != nil {
						t.getLogger().Error("can not change directory", er)
						t.out(MsgError(MsgError{Err: er, Reference: basePath, Target: currentTask.ID}))
						return nil, er
					}
				}
			}
		}

	}
	return curDir, nil
}

// ExecuteScriptLine executes a script line and returns the exit code
// the callback function is called for each line of the output
// the startInfo function is called if the process started
func (t *targetExecuter) ExecuteScriptLine(dCmd string, dCmdArgs []string, command string, callback func(string, error) bool, startInfo func(*os.Process)) (int, int, error) {
	return Execute(dCmd, dCmdArgs, command, callback, startInfo)
}

// Execute executes a command and returns the internal exit code, the command exit code and an error
// the callback function is called for each line of the output
// the startInfo function is called if the process started and the process id is available
func Execute(dCmd string, dCmdArgs []string, command string, callback func(string, error) bool, startInfo func(*os.Process)) (int, int, error) {
	cmdArg := append(dCmdArgs, command)
	cmd := exec.Command(dCmd, cmdArg...)

	stdoutPipe, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	err := cmd.Start()
	if err != nil {
		return systools.ExitCmdError, 0, err
	}
	process.TryPid2Pgid(cmd)

	startInfo(cmd.Process)
	scanner := bufio.NewScanner(stdoutPipe)

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		keepRunning := callback(m, nil)
		if !keepRunning {
			cmd.Process.Kill()
			process.KillProcessTree(cmd.Process.Pid)
			return systools.ExitByStopReason, 0, err
		}

	}
	err = cmd.Wait()
	if err != nil {
		callback(err.Error(), err)
		errRealCode := 0
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				errRealCode = status.ExitStatus()
			}

		}
		return systools.ExitCmdError, errRealCode, err
	}

	return systools.ExitOk, 0, err
}

// listenerWatch checks if a trigger is hit and executes the action
func (t *targetExecuter) listenerWatch(logLine string, e error, currentTask *configure.Task) {
	if currentTask.Listener != nil {

		for _, listener := range currentTask.Listener {
			triggerFound, triggerMessage := t.checkReason(listener.Trigger, logLine, e) // check if a trigger have a match
			if triggerFound {
				t.setPh("RUN."+t.target+".LOG.HIT", logLine)
				if currentTask.Options.Displaycmd {
					t.out(MsgType("run-trigger-sricpt-line"), MsgCommand(logLine))
				}

				someReactionTriggered := false                 // did this trigger something? used as flag
				actionDef := configure.Action(listener.Action) // extract action

				if keyname, ok := t.verifiedKeyname(actionDef.Target); !ok { // check if the target is a valid keyname
					t.out(MsgError(MsgError{Err: errors.New("invalid keyname for target reference: " + actionDef.Target), Reference: triggerMessage, Target: t.target}))
					return
				} else {
					actionDef.Target = keyname
				}

				if len(actionDef.Script) > 0 { // script are directs executes without any async or other executes out of scope
					someReactionTriggered = true
					var dummyArgs map[string]string = make(map[string]string) // create empty arguments as scoped values
					for _, triggerScript := range actionDef.Script {          // run any line of script
						t.getLogger().Debug("TRIGGER SCRIPT ACTION", triggerScript)
						subRun := t.CopyToTarget(t.target)
						subRun.SetArgs(dummyArgs)
						subRun.targetTaskExecuter(triggerScript, *currentTask, t.watch)
					}

				}

				if actionDef.Target != "" { // here we have a target defined thats needs to be started
					someReactionTriggered = true

					logFields := mimiclog.Fields{"target": actionDef.Target, "trigger": triggerMessage}
					t.getLogger().Debug("TRIGGER ACTION", logFields)

					if currentTask.Options.Displaycmd {

						t.out(MsgType("run-trigger-target-output"), MsgCommand(actionDef.Target), MsgTarget{Target: actionDef.Target, Context: "run-trigger-target-output", Info: "start triggered action"})
					}

					var scopeVars map[string]string = make(map[string]string) // create empty arguments as scoped values

					// because we are anyway in a async scope, we should no longer
					// try to run this target too async.
					// also the target is triggered by an specific log entriy, it makes
					// sence to stop the execution of the parent, til this target is executed
					t.out(MsgTarget{Target: actionDef.Target, Context: "execute-trigger-target", Info: "start triggered action"})
					t.executeTemplate(false, actionDef.Target, scopeVars)

				}
				if !someReactionTriggered {
					t.out(
						MsgError(MsgError{Err: errors.New("trigger defined without any action"), Reference: triggerMessage, Target: t.target}),
						MsgInfo(triggerMessage),
					)
					t.getLogger().Warn("trigger defined without any action", triggerMessage, logLine)
				}
			} else {
				t.getLogger().Debug("no trigger found in", logLine)
			}
		}
	}
}
