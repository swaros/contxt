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
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/systools"
)

// lineExecuter is the main function to execute a script line
// it returns the exit code of the executed command
// and a boolean value if the execution was successful
func (t *targetExecuter) lineExecuter(codeLine string, currentTask configure.Task) (int, bool) {
	replacedLine := codeLine
	if t.phHandler != nil {
		replacedLine = t.phHandler.HandlePlaceHolderWithScope(codeLine, t.arguments) // placeholders
	}
	t.out(MsgTarget{Target: currentTask.ID, Context: "command", Info: replacedLine}) // output the command
	t.setPh("RUN."+currentTask.ID+".CMD.LAST", replacedLine)                         // set or overwrite the last script command for the target
	t.setPh("RUN.SCRIPT_LINE", replacedLine)                                         // set or overwrite the last script command for the target

	runCmd, runArgs := t.commandFallback.GetMainCmd(currentTask.Options) // get the main command and arguments
	t.SetMainCmd(runCmd, runArgs...)                                     // set the main command and arguments

	// here we execute the current script line
	execCode, realExitCode, execErr := t.ExecuteScriptLine(runCmd, runArgs, replacedLine,
		func(logLine string, err error) bool { // callback for any logline
			t.setPh("RUN."+currentTask.ID+".LOG.LAST", logLine) // set or overwrite the last script output for the target
			if currentTask.Listener != nil {                    // do we have listener?
				t.listenerWatch(logLine, err, &currentTask) // listener handler
			}

			// The whole output can be ignored by configuration
			// if this is not enabled then we handle all these here
			if !currentTask.Options.Hideout {

				outStr := logLine                    // hardcoded format for the logoutput iteself
				if currentTask.Options.Stickcursor { // optional set back the cursor to the beginning
					t.out(MsgStickCursor(true)) // trigger the stick cursor
				}

				t.out(MsgExecOutput(outStr))         // prints the codeline
				if currentTask.Options.Stickcursor { // cursor stick handling
					t.out(MsgStickCursor(false)) // trigger the stick cursor after output
				}
			}

			stopReasonFound, message := t.checkReason(currentTask.Stopreasons, logLine, err) // do we found a defined reason to stop execution
			if stopReasonFound {
				if currentTask.Options.Displaycmd {
					t.out(MsgType("stopreason"), MsgReason(message), MsgProcess("aborted"))
				}
				return false
			}
			return true
		}, func(process *os.Process) { // callback if the process started and we got the process id
			pidStr := fmt.Sprintf("%d", process.Pid) // we use them as info for the user only
			t.setPh("RUN.PID", pidStr)
			t.setPh("RUN."+t.target+".PID", pidStr)
			if currentTask.Options.Displaycmd {
				t.out(MsgPid(process.Pid), MsgProcess("started"))
			}
		})

	// check execution codes from the executer
	if execErr != nil {
		if currentTask.Options.Displaycmd {
			t.out("exec error", MsgError(execErr))
		}

	}
	// check execution codes
	switch execCode {
	case systools.ExitByStopReason:
		return systools.ExitByStopReason, true
	case systools.ExitCmdError:
		if currentTask.Options.IgnoreCmdError {
			if currentTask.Stopreasons.Onerror {
				t.out(MsgTarget{Target: t.target, Context: "error-catch-by-onerror", Info: "" + execErr.Error()}, MsgError(execErr), MsgCommand(codeLine), MsgNumber(realExitCode))
				return systools.ExitByStopReason, true
			}
			t.out(MsgTarget{Target: t.target, Context: "execution-error-ignored", Info: execErr.Error()}, MsgError(execErr), MsgCommand(codeLine), MsgNumber(realExitCode))

		} else {
			t.getLogger().WithFields(logrus.Fields{"processCode": realExitCode, "error": execErr}).Error("task exection error")
			ErrorMsg := errors.New(codeLine + " fails with error: " + execErr.Error())
			t.out(MsgTarget{Target: t.target, Context: "execution-error", Info: ErrorMsg.Error()}, MsgError(ErrorMsg), MsgCommand(codeLine), MsgNumber(realExitCode))
			//systools.Exit(realExitCode) // origin behavior

			// returns the error code
			return systools.ExitCmdError, true
		}
	case systools.ExitOk:
		return systools.ExitOk, false
	}
	return systools.ExitNoCode, true
}

// ExecuteScriptLine executes a script line and returns the exit code
// the callback function is called for each line of the output
// the startInfo function is called if the process started
func (t *targetExecuter) ExecuteScriptLine(dCmd string, dCmdArgs []string, command string, callback func(string, error) bool, startInfo func(*os.Process)) (int, int, error) {
	return Execute(dCmd, dCmdArgs, command, callback, startInfo)
}

func Execute(dCmd string, dCmdArgs []string, command string, callback func(string, error) bool, startInfo func(*os.Process)) (int, int, error) {
	cmdArg := append(dCmdArgs, command)
	cmd := exec.Command(dCmd, cmdArg...)
	stdoutPipe, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	err := cmd.Start()
	if err != nil {
		return systools.ExitCmdError, 0, err
	}

	startInfo(cmd.Process)
	scanner := bufio.NewScanner(stdoutPipe)

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		keepRunning := callback(m, nil)
		if !keepRunning {
			cmd.Process.Kill()
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

				if len(actionDef.Script) > 0 { // script are directs executes without any async or other executes out of scope
					someReactionTriggered = true
					var dummyArgs map[string]string = make(map[string]string) // create empty arguments as scoped values
					for _, triggerScript := range actionDef.Script {          // run any line of script
						t.getLogger().WithFields(logrus.Fields{
							"cmd": triggerScript,
						}).Debug("TRIGGER SCRIPT ACTION")
						subRun := t.CopyToTarget(t.target)
						subRun.SetArgs(dummyArgs)
						subRun.lineExecuter(triggerScript, *currentTask)
					}

				}

				if actionDef.Target != "" { // here we have a target defined thats needs to be started
					someReactionTriggered = true
					t.getLogger().WithFields(logrus.Fields{
						"target":  actionDef.Target,
						"trigger": triggerMessage,
					}).Debug("TRIGGER ACTION")

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
					t.out(MsgError(errors.New("trigger-defined-without-action")), MsgInfo(triggerMessage))
					t.getLogger().WithFields(logrus.Fields{
						"trigger": triggerMessage,
						"output":  logLine,
					}).Warn("trigger defined without any action")
				}
			} else {
				t.getLogger().WithFields(logrus.Fields{
					"output": logLine,
				}).Debug("no trigger found")
			}
		}
	}
}
