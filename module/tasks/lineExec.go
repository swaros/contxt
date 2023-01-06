package tasks

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/manout"
)

func (t *targetExecuter) lineExecuter(codeLine string, currentTask configure.Task) (int, bool) {
	replacedLine := codeLine
	if t.phHandler != nil {
		replacedLine = t.phHandler.HandlePlaceHolderWithScope(codeLine, t.arguments) // placeholders
	}
	t.out(MsgTarget(currentTask.ID), MsgCommand(replacedLine)) // output the command
	t.setPh("RUN."+currentTask.ID+".CMD.LAST", replacedLine)   // set or overwrite the last script command for the target
	t.setPh("RUN.SCRIPT_LINE", replacedLine)                   // set or overwrite the last script command for the target

	// here we execute the current script line
	execCode, realExitCode, execErr := t.ExecuteScriptLine(replacedLine,
		func(logLine string, err error) bool { // callback for any logline
			t.setPh("RUN."+currentTask.ID+".LOG.LAST", logLine) // set or overwrite the last script output for the target
			if currentTask.Listener != nil {                    // do we have listener?
				t.listenerWatch(logLine, err, &currentTask) // listener handler
			}

			// The whole output can be ignored by configuration
			// if this is not enabled then we handle all these here
			if !currentTask.Options.Hideout {

				//outStr := systools.LabelPrintWithArg(logLine, colorCode, "39", 2) // hardcoded format for the logoutput iteself
				outStr := manout.MessageCln(logLine)
				if currentTask.Options.Stickcursor { // optional set back the cursor to the beginning
					//fmt.Print("\033[G\033[K") // done by escape codes
					t.out(MsgStickCursor(true)) // trigger the stick cursor
				}

				t.out(MsgExecOutput(outStr))         // prints the codeline
				if currentTask.Options.Stickcursor { // cursor stick handling
					//fmt.Print("\033[A")
					t.out(MsgStickCursor(false)) // trigger the stick cursor after output
				}
			}

			stopReasonFound, message := t.checkReason(currentTask.Stopreasons, logLine, err) // do we found a defined reason to stop execution
			if stopReasonFound {
				if currentTask.Options.Displaycmd {
					t.out(MsgType("stopreason"), MsgReason(message))
				}
				return false
			}
			return true
		}, func(process *os.Process) { // callback if the process started and we got the process id
			pidStr := fmt.Sprintf("%d", process.Pid) // we use them as info for the user only
			t.setPh("RUN.PID", pidStr)
			t.setPh("RUN."+t.target+".PID", pidStr)
			if currentTask.Options.Displaycmd {
				t.out(MsgProcess("pid"), MsgPid(process.Pid))
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
				return systools.ExitByStopReason, true
			}
			t.out(manout.MessageCln(manout.ForeYellow, "NOTE!\t", manout.BackLightYellow, manout.ForeDarkGrey, " a script execution was failing. no stopreason is set so execution will continued "))
			t.out(manout.MessageCln("\t", manout.BackLightYellow, manout.ForeDarkGrey, " if this is expected you can ignore this message.                                 "))
			t.out(manout.MessageCln("\t", manout.BackLightYellow, manout.ForeDarkGrey, " but you should handle error cases                                                "))
			t.out("\ttarget :\t", manout.MessageCln(manout.ForeBlue, t.target))
			t.out("\tcommand:\t", manout.MessageCln(manout.ForeYellow, codeLine))

		} else {
			errMsg := " = exit code from command: "
			lastMessage := manout.MessageCln(manout.BackRed, manout.ForeYellow, realExitCode, manout.CleanTag, manout.ForeLightRed, errMsg, manout.ForeWhite, codeLine)
			t.out("\t Exit ", lastMessage)
			t.out()
			t.out("\t check the command. if this command can fail you may fit the execution rules. see options:")
			t.out("\t you may disable a hard exit on error by setting ignoreCmdError: true")
			t.out("\t if you do so, a Note will remind you, that a error is happend in this case.")
			t.out()
			t.getLogger().WithFields(logrus.Fields{"processCode": realExitCode, "error": execErr}).Error("task exection error")

			//systools.Exit(realExitCode) // origin behavior

			// returns the error code
			return systools.ExitCmdError, true
		}
	case systools.ExitOk:
		return systools.ExitOk, false
	}
	return systools.ExitNoCode, true
}

func (t *targetExecuter) getCmd() (string, []string) {
	defaultCmd, defaultArgs := t.commandFallback.GetMainCmd()
	if t.mainCmd != "" {
		defaultCmd = t.mainCmd
	}
	if len(t.mainCmdArgs) > 0 {
		defaultArgs = t.mainCmdArgs
	}
	return defaultCmd, defaultArgs
}

func (t *targetExecuter) ExecuteScriptLine(command string, callback func(string, error) bool, startInfo func(*os.Process)) (int, int, error) {
	dCmd, dCmdArgs := t.getCmd()
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

func (t *targetExecuter) listenerWatch(logLine string, e error, currentTask *configure.Task) {
	if currentTask.Listener != nil {

		for _, listener := range currentTask.Listener {
			triggerFound, triggerMessage := t.checkReason(listener.Trigger, logLine, e) // check if a trigger have a match
			if triggerFound {
				t.setPh("RUN."+t.target+".LOG.HIT", logLine)
				if currentTask.Options.Displaycmd {
					t.out(manout.MessageCln(manout.ForeCyan, "[trigger]\t", manout.ForeYellow, triggerMessage, manout.Dim, " ", logLine))
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
						"target": actionDef.Target,
					}).Debug("TRIGGER ACTION")

					if currentTask.Options.Displaycmd {
						t.out(manout.MessageCln(manout.ForeCyan, "[trigger]\t ", manout.ForeGreen, "target:", manout.ForeLightGreen, actionDef.Target))
					}

					// TODO: check if this is stille neded, or somehow usefull
					hitKeyCnt := "RUN.LISTENER." + actionDef.Target + ".HIT.CNT"
					lastCnt := t.getPh(hitKeyCnt)
					if lastCnt == "" {
						t.setPh(hitKeyCnt, "1")
					} else {
						iCnt, err := strconv.Atoi(lastCnt)
						if err != nil {
							t.getLogger().Fatal("fail converting trigger count")
						}
						iCnt++
						t.setPh(hitKeyCnt, strconv.Itoa(iCnt))
					}

					t.getLogger().WithFields(logrus.Fields{
						"trigger":   triggerMessage,
						"target":    actionDef.Target,
						"hitKeyCnt": hitKeyCnt,
					}).Info("TRIGGER Called")
					var scopeVars map[string]string = make(map[string]string)

					t.getLogger().WithFields(logrus.Fields{
						"target": actionDef.Target,
					}).Info("RUN Triggered target (not async)")

					// because we are anyway in a async scope, we should no longer
					// try to run this target too async.
					// also the target is triggered by an specific log entriy, it makes
					// sence to stop the execution of the parent, til this target is executed
					t.out("running target ", manout.ForeCyan, actionDef.Target, manout.ForeLightCyan, " trigger action")
					t.executeTemplate(false, actionDef.Target, scopeVars)

				}
				if !someReactionTriggered {
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
