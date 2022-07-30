// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Licensed under the MIT License
//
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
package cmdhandle

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/context/dirhandle"
	"github.com/swaros/manout"

	"github.com/swaros/contxt/context/systools"

	"github.com/swaros/contxt/context/configure"
)

const (
	// ExitOk the process was executed without errors
	ExitOk = 0
	// ExitByStopReason the process stopped because of a defined reason
	ExitByStopReason = 101
	// ExitNoCode means there was no code associated
	ExitNoCode = 102
	// ExitCmdError means the execution of the command fails. a error by the command itself
	ExitCmdError = 103
	// ExitByRequirement means a requirement was not fulfills
	ExitByRequirement = 104
	// ExitAlreadyRunning means the task is not started, because it is already created
	ExitAlreadyRunning = 105
)

// this flag is for the runner logic replacement that have still issues.
// this 'solution' is not nice but a different branch would be more difficult to handle
var Experimental = true

// SharedFolderExecuter runs shared .contxt.yml files directly without merging them into
// the current contxt file
func SharedFolderExecuter(template configure.RunConfig, locationHandle func(string, string)) {
	if len(template.Config.Use) > 0 {
		GetLogger().WithField("uses", template.Config.Use).Info("shared executer")
		for _, shared := range template.Config.Use {
			externalPath := HandleUsecase(shared)
			GetLogger().WithField("path", externalPath).Info("shared contxt location")
			currentDir, _ := dirhandle.Current()
			os.Chdir(externalPath)
			locationHandle(externalPath, currentDir)
			os.Chdir(currentDir)
		}
	}
}

func RunShared(targets string) {

	allTargets := strings.Split(targets, ",")
	template, templatePath, exists, terr := GetTemplate()
	if terr != nil {
		CtxOut(manout.MessageCln(manout.ForeRed, "Error ", manout.CleanTag, terr.Error()))
		return
	}
	if !exists {
		return
	}

	if template.Config.Loglevel != "" { // set logger level by template definition
		setLogLevelByString(template.Config.Loglevel)
	}

	GetLogger().WithField("targets", allTargets).Info("SHARED START run targets...")

	// handle all shared usages. these usages are set
	// in the template by the string map named Use in the config section
	// Config:
	//    Use:
	//      - shared_task_1
	//      - shared_task_2
	if len(template.Config.Use) > 0 {
		GetLogger().WithField("uses", template.Config.Use).Info("found external dependecy")
		CtxOut(manout.MessageCln(manout.ForeCyan, "[SHARED loop]"))
		for _, shared := range template.Config.Use {
			CtxOut(manout.MessageCln(manout.ForeCyan, "[SHARED CONTXT][", manout.ForeBlue, shared, manout.ForeCyan, "] "))
			externalPath := HandleUsecase(shared)
			GetLogger().WithField("path", externalPath).Info("shared contxt location")
			currentDir, _ := dirhandle.Current()
			os.Chdir(externalPath)
			for _, runTarget := range allTargets {
				CtxOut(manout.MessageCln(manout.ForeCyan, manout.ForeGreen, runTarget, manout.ForeYellow, " [ external:", manout.ForeWhite, externalPath, manout.ForeYellow, "] ", manout.ForeDarkGrey, templatePath))
				RunTargets(runTarget, false)
				CtxOut(manout.MessageCln(manout.ForeCyan, "["+manout.ForeBlue, shared+"] ", manout.ForeGreen, runTarget, " DONE"))
			}
			os.Chdir(currentDir)
		}
		CtxOut(manout.MessageCln(manout.ForeCyan, "[SHARED done]"))
	}
	GetLogger().WithField("targets", allTargets).Info("  SHARED DONE run targets...")
}

// RunTargets executes multiple targets
// the targets string can have multiple targets
// seperated by comma
func RunTargets(targets string, sharedRun bool) {

	// validate first
	if err := TestTemplate(); err != nil {
		CtxOut("found issues in the current template ", err)
		systools.Exit(32)
		return
	}

	SetPH("CTX_TARGETS", targets)

	// this flag should only true on the first execution
	if sharedRun {
		// do it here makes sure we are not in the shared scope
		currentDir, _ := dirhandle.Current()
		SetPH("CTX_PWD", currentDir)
		// run shared use
		RunShared(targets)
	}

	allTargets := strings.Split(targets, ",")
	template, templatePath, exists, terr := GetTemplate()
	if terr != nil {
		CtxOut(manout.MessageCln(manout.ForeRed, "Error ", manout.CleanTag, terr.Error()))
		systools.Exit(33)
		return
	}
	GetLogger().WithField("targets", allTargets).Info("run targets...")
	var runSequencially = false // default is async mode
	if exists {                 // TODO: the exists check just for this config reading seems wrong
		runSequencially = template.Config.Sequencially
		if template.Config.Coloroff {
			manout.ColorEnabled = false
		}
	}

	if template.Config.Loglevel != "" { // loglevel by config
		setLogLevelByString(template.Config.Loglevel)
	}

	var wg sync.WaitGroup // the main waitgroup

	// handle all imports.
	// these are yaml or json files, they can be accessed for reading by the gson doted format
	if len(template.Config.Imports) > 0 {
		GetLogger().WithField("Import", template.Config.Imports).Info("import second level vars")
		handleFileImportsToVars(template.Config.Imports)
	} else {
		GetLogger().Info("No second level Variables defined")
	}

	// experimental usage of taskrunner
	if Experimental {
		if runSequencially { // non async run
			for _, trgt := range allTargets {
				SetPH("CTX_TARGET", trgt)
				CtxOut(LabelFY("exec"), InfoMinor("execute target in sequence"), ValF(trgt), manout.ForeLightCyan, " ", templatePath)
				ExecPathFile(&wg, !runSequencially, template, trgt)
			}
		} else {
			var futuresExecs []FutureStack
			for _, trgt := range allTargets { // iterate all targets
				CtxOut(LabelFY("exec"), InfoMinor("execute target in Async"), ValF(trgt), manout.ForeLightCyan, " ", templatePath)
				futuresExecs = append(futuresExecs, FutureStack{
					AwaitFunc: func(ctx context.Context) interface{} {
						ctxTarget := ctx.Value(CtxKey{}).(string)                       // get the target from context
						SetPH("CTX_TARGET", ctxTarget)                                  // update global target. TODO: makes this any sense in async?
						return ExecPathFile(&wg, !runSequencially, template, ctxTarget) // execute target
					},
					Argument: trgt,
				})
			}
			futures := ExecFutureGroup(futuresExecs)                      // execute all async task
			CtxOut(LabelFY("exec"), "all targets started ", len(targets)) // just info
			WaitAtGroup(futures)                                          // wait until all task are done
			CtxOut(LabelFY("exec"), "all targets done ", len(targets))    // also just info for the user
		}

	} else {
		// NONE experimental usage of taskrunner
		if !runSequencially {
			// run in thread
			for _, runTarget := range allTargets {
				SetPH("CTX_TARGET", runTarget)
				wg.Add(1)
				CtxOut(manout.MessageCln(manout.ForeBlue, "[exec:async] ", manout.BoldTag, runTarget, " ", manout.ForeWhite, templatePath))
				go ExecuteTemplateWorker(&wg, true, runTarget, template)
			}
			wg.Wait()
		} else {
			// trun one by one
			for _, runTarget := range allTargets {
				SetPH("CTX_TARGET", runTarget)
				CtxOut(manout.MessageCln(manout.ForeBlue, "[exec:seq] ", manout.BoldTag, runTarget, " ", manout.ForeWhite, templatePath))
				exitCode := ExecPathFile(&wg, false, template, runTarget)
				GetLogger().WithField("exitcode", exitCode).Info("RunTarget [Sequencially runmode] done with exitcode")
			}
		}
	}

	CtxOut(manout.MessageCln(manout.ForeBlue, "[done] ", manout.BoldTag, targets))
	GetLogger().Info("target task execution done")
}

func setLogLevelByString(loglevel string) {
	level, err := logrus.ParseLevel(loglevel)
	if err != nil {
		GetLogger().Error("Invalid loglevel in task defined.", err)
	} else {
		GetLogger().SetLevel(level)
	}

}

func listenerWatch(script configure.Task, target, logLine string, waitGroup *sync.WaitGroup, useWaitGroup bool, runCfg configure.RunConfig) {
	if script.Listener != nil {

		GetLogger().WithFields(logrus.Fields{
			"count":    len(script.Listener),
			"listener": script.Listener,
		}).Debug("testing Listener")

		for _, listener := range script.Listener {
			triggerFound, triggerMessage := checkReason(listener.Trigger, logLine) // check if a trigger have a match
			if triggerFound {
				SetPH("RUN."+target+".LOG.HIT", logLine)
				if script.Options.Displaycmd {
					CtxOut(manout.MessageCln(manout.ForeCyan, "[trigger]\t", manout.ForeYellow, triggerMessage, manout.Dim, " ", logLine))
				}

				someReactionTriggered := false                 // did this trigger something? used as flag
				actionDef := configure.Action(listener.Action) // extract action

				if len(actionDef.Script) > 0 { // script are directs executes without any async or other executes out of scope
					someReactionTriggered = true
					var dummyArgs map[string]string = make(map[string]string) // create empty arguments as scoped values
					for _, triggerScript := range actionDef.Script {          // run any line of script
						GetLogger().WithFields(logrus.Fields{
							"cmd": triggerScript,
						}).Debug("TRIGGER SCRIPT ACTION")
						lineExecuter(waitGroup, useWaitGroup, script.Stopreasons, runCfg, "93", "46", triggerScript, target, dummyArgs, script)
					}

				}

				if actionDef.Target != "" { // here we have a target defined thats needs to be started
					someReactionTriggered = true
					GetLogger().WithFields(logrus.Fields{
						"target": actionDef.Target,
					}).Debug("TRIGGER ACTION")

					if script.Options.Displaycmd {
						CtxOut(manout.MessageCln(manout.ForeCyan, "[trigger]\t ", manout.ForeGreen, "target:", manout.ForeLightGreen, actionDef.Target))
					}

					// TODO: i can't remember why i am doing this placeholder thing
					hitKeyTargets := "RUN.LISTENER." + target + ".HIT.TARGETS" // compose the placeholder key
					lastHitTargets := GetPH(hitKeyTargets)                     // get the last stored value if exists
					if !strings.Contains(lastHitTargets, "("+actionDef.Target+")") {
						lastHitTargets = lastHitTargets + "(" + actionDef.Target + ")"
						SetPH(hitKeyTargets, lastHitTargets)
					}

					hitKeyCnt := "RUN.LISTENER." + actionDef.Target + ".HIT.CNT"
					lastCnt := GetPH(hitKeyCnt)
					if lastCnt == "" {
						SetPH(hitKeyCnt, "1")
					} else {
						iCnt, err := strconv.Atoi(lastCnt)
						if err != nil {
							GetLogger().Fatal("fail converting trigger count")
						}
						iCnt++
						SetPH(hitKeyCnt, strconv.Itoa(iCnt))
					}

					GetLogger().WithFields(logrus.Fields{
						"trigger":   triggerMessage,
						"target":    actionDef.Target,
						"waitgroup": useWaitGroup,
						"RUN.LISTENER." + target + ".HIT.TARGETS": lastHitTargets,
					}).Info("TRIGGER Called")
					var scopeVars map[string]string = make(map[string]string)

					if Experimental {
						GetLogger().WithFields(logrus.Fields{
							"target": actionDef.Target,
						}).Info("RUN Triggered target (not async)")

						// because we are anyway in a async scope, we should no longer
						// try to run this target too async.
						// also the target is triggered by an specific log entriy, it makes
						// sence to stop the execution of the parent, til this target is executed
						CtxOut("running target ", manout.ForeCyan, actionDef.Target, manout.ForeLightCyan, " trigger action")
						executeTemplate(waitGroup, useWaitGroup, runCfg, actionDef.Target, scopeVars)
					} else {

						if useWaitGroup {
							GetLogger().WithFields(logrus.Fields{
								"target": actionDef.Target,
							}).Info("RUN ASYNC")

							go executeTemplate(waitGroup, useWaitGroup, runCfg, actionDef.Target, scopeVars)

						} else {
							GetLogger().WithFields(logrus.Fields{
								"target": actionDef.Target,
							}).Info("RUN SEQUENCE")
							executeTemplate(waitGroup, useWaitGroup, runCfg, actionDef.Target, scopeVars)
						}
					}
				}
				if !someReactionTriggered {
					GetLogger().WithFields(logrus.Fields{
						"trigger": triggerMessage,
						"output":  logLine,
					}).Warn("trigger defined without any action")
				}
			} else {
				GetLogger().WithFields(logrus.Fields{
					"output": logLine,
				}).Debug("no trigger found")
			}
		}
	}
}

// the main script handler
func lineExecuter(
	waitGroup *sync.WaitGroup, // the main waitgoup
	useWaitGroup bool, // flag if we have to use the waitgroup. also means we run in async mode
	stopReason configure.Trigger, // configuration for the stop reasons
	runCfg configure.RunConfig, // the runtime configuration
	colorCode, bgCode, // colorcodes for the left panel
	codeLine, // the script that have to be processed
	target string, // the actual target
	arguments map[string]string, // the arguments for the current scope
	script configure.Task) (int, bool) {
	panelSize := 12                   // default panelsize
	if script.Options.Panelsize > 0 { // overwrite panel size if set
		panelSize = script.Options.Panelsize
	}
	var mainCommand = defaultString(script.Options.Maincmd, DefaultCommandFallBack) // get the maincommand by default first
	if configure.GetOs() == "windows" {                                             // handle windows behavior depending default commands
		mainCommand = defaultString(script.Options.Maincmd, DefaultCommandFallBackWindows)
	}
	replacedLine := HandlePlaceHolderWithScope(codeLine, arguments) // placeholders
	if script.Options.Displaycmd {                                  // do we show the argument?
		CtxOut(LabelFY("cmd"), ValF(target), InfoF(replacedLine))
	}

	SetPH("RUN.SCRIPT_LINE", replacedLine) // overwrite the current scriptline. this is only reliable if we not in async mode
	var targetLabel CtxTargetOut = CtxTargetOut{
		ForeCol:   defaultString(script.Options.Colorcode, colorCode),
		BackCol:   defaultString(script.Options.Bgcolorcode, bgCode),
		PanelSize: panelSize,
	}
	// here we execute the current script line
	execCode, realExitCode, execErr := ExecuteScriptLine(mainCommand, script.Options.Mainparams, replacedLine,
		func(logLine string) bool { // callback for any logline

			SetPH("RUN."+target+".LOG.LAST", logLine) // set or overwrite the last script output for the target

			if script.Listener != nil { // do we have listener?
				GetLogger().WithFields(logrus.Fields{
					"cnt":      len(script.Listener),
					"listener": script.Listener,
				}).Debug("CHECK Listener")
				listenerWatch(script, target, logLine, waitGroup, useWaitGroup, runCfg) // listener handler
			}
			targetLabel.Target = target
			// The whole output can be ignored by configuration
			// if this is not enabled then we handle all these here
			if !script.Options.Hideout {
				// the background color
				if script.Options.Format != "" { // do we have a specific format for the label, then we use them instead
					format := HandlePlaceHolderWithScope(script.Options.Format, script.Variables) // handle placeholder in the label
					fomatedOutStr := manout.Message(fmt.Sprintf(format, target))                  // also format the message depending format codes
					targetLabel.Alternative = fomatedOutStr
				}

				//outStr := systools.LabelPrintWithArg(logLine, colorCode, "39", 2) // hardcoded format for the logoutput iteself
				outStr := manout.MessageCln(logLine)
				if script.Options.Stickcursor { // optional set back the cursor to the beginning
					fmt.Print("\033[G\033[K") // done by escape codes
				}

				CtxOut(targetLabel, outStr)     // prints the codeline
				if script.Options.Stickcursor { // cursor stick handling
					fmt.Print("\033[A")
				}
			}

			stopReasonFound, message := checkReason(stopReason, logLine) // do we found a defined reason to stop execution
			if stopReasonFound {
				if script.Options.Displaycmd {
					CtxOut(LabelFY("stop-reason"), ValF(message))
				}
				return false
			}
			return true
		}, func(process *os.Process) { // callback if the process started and we got the process id
			pidStr := fmt.Sprintf("%d", process.Pid) // we use them as info for the user only
			SetPH("RUN.PID", pidStr)
			SetPH("RUN."+target+".PID", pidStr)
			if script.Options.Displaycmd {
				CtxOut(LabelFY("pid"), ValF(process.Pid))
			}
		})
	if execErr != nil {
		if script.Options.Displaycmd {
			CtxOut(LabelErrF("exec error"), ValF(execErr))
		}
	}
	// check execution codes
	switch execCode {
	case ExitByStopReason:
		return ExitByStopReason, true
	case ExitCmdError:
		if script.Options.IgnoreCmdError {
			if script.Stopreasons.Onerror {
				return ExitByStopReason, true
			}
			CtxOut(manout.MessageCln(manout.ForeYellow, "NOTE!\t", manout.BackLightYellow, manout.ForeDarkGrey, " a script execution was failing. no stopreason is set so execution will continued "))
			CtxOut(manout.MessageCln("\t", manout.BackLightYellow, manout.ForeDarkGrey, " if this is expected you can ignore this message.                                 "))
			CtxOut(manout.MessageCln("\t", manout.BackLightYellow, manout.ForeDarkGrey, " but you should handle error cases                                                "))
			CtxOut("\ttarget :\t", manout.MessageCln(manout.ForeYellow, target))
			CtxOut("\tcommand:\t", manout.MessageCln(manout.ForeYellow, codeLine))

		} else {
			errMsg := " = exit code from command: "
			lastMessage := manout.MessageCln(manout.BackRed, manout.ForeYellow, realExitCode, manout.CleanTag, manout.ForeLightRed, errMsg, manout.ForeWhite, codeLine)
			CtxOut("\t Exit ", lastMessage)
			CtxOut()
			CtxOut("\t check the command. if this command can fail you may fit the execution rules. see options:")
			CtxOut("\t you may disable a hard exit on error by setting ignoreCmdError: true")
			CtxOut("\t if you do so, a Note will remind you, that a error is happend in this case.")
			CtxOut()
			GetLogger().Error("runtime error:", execErr, "exit", realExitCode)
			systools.Exit(realExitCode)
			// returns the error code
			return ExitCmdError, true
		}
	case ExitOk:
		return ExitOk, false
	}
	return ExitNoCode, true
}

func generateFuturesByTargetListAndExec(RunTargets []string, waitGroup *sync.WaitGroup, runCfg configure.RunConfig) []Future {
	if len(RunTargets) < 1 {
		return []Future{}
	}
	var runTargetExecs []FutureStack
	for _, needTarget := range RunTargets {
		GetLogger().Debug("runTarget name should be added " + needTarget)
		runTargetExecs = append(runTargetExecs, FutureStack{
			AwaitFunc: func(ctx context.Context) interface{} {
				argTask := ctx.Value(CtxKey{}).(string)
				_, argmap := StringSplitArgs(argTask, "arg")
				GetLogger().Debug("add runTarget task " + argTask)
				return executeTemplate(waitGroup, true, runCfg, argTask, argmap)
			},
			Argument: needTarget})

	}
	CtxOut(LabelFY("async targets"), InfoF("count"), len(runTargetExecs), InfoF(" targets"))
	return ExecFutureGroup(runTargetExecs)
}

// merge a list of task to an single task.
func mergeTargets(target string, runCfg configure.RunConfig) configure.Task {
	var checkTasks configure.Task
	first := true
	if len(runCfg.Task) > 0 {
		for _, script := range runCfg.Task {
			if strings.EqualFold(target, script.ID) {
				canRun, failMessage := checkRequirements(script.Requires)
				if canRun {
					// update task variables
					for keyName, variable := range script.Variables {
						SetPH(keyName, HandlePlaceHolder(variable))
					}
					if first {
						checkTasks = script
						first = false
					} else {
						mergo.Merge(&checkTasks, script, mergo.WithOverride, mergo.WithAppendSlice)
					}
				} else {
					GetLogger().Debug(failMessage)
				}
			}
		}
	}
	return checkTasks
}

func executeTemplate(waitGroup *sync.WaitGroup, runAsync bool, runCfg configure.RunConfig, target string, scopeVars map[string]string) int {
	if runAsync {
		waitGroup.Add(1)
		defer waitGroup.Done()
	}
	// check if task is already running
	// this check depends on the target name.
	if !runCfg.Config.AllowMutliRun && TaskRunning(target) {
		GetLogger().WithField("task", target).Warning("task would be triggered again while is already running. IGNORED")
		return ExitAlreadyRunning
	}
	// increment task counter
	incTaskCount(target)
	defer incTaskDoneCount(target)

	GetLogger().WithFields(logrus.Fields{
		"target": target,
	}).Info("executeTemplate LOOKING for target")

	// Checking if the Tasklist have something
	// to handle
	if len(runCfg.Task) > 0 {
		returnCode := ExitOk

		// the main variables will be set at first
		// but only if the they not already exists
		// from other task or by start argument
		for keyName, variable := range runCfg.Config.Variables {
			SetIfNotExists(keyName, HandlePlaceHolder(variable))
		}
		// set the colorcodes for the labels on left side of screen
		colorCode, bgCode := systools.CreateColorCode()

		// updates global variables
		SetPH("RUN.TARGET", target)

		// this flag is only used
		// for a "target not found" message later
		targetFound := false

		// oure tasklist that will use later
		var taskList []configure.Task

		// depending on the config
		// we merge the tasks and handle them as one task,
		// or we keep them as a list of tasks what would
		// keep more flexibility.
		// by merging task we can loose runtime definitions
		if runCfg.Config.MergeTasks {
			mergedScript := mergeTargets(target, runCfg)
			taskList = append(taskList, mergedScript)
		} else {
			for _, script := range runCfg.Task {
				if strings.EqualFold(target, script.ID) {
					taskList = append(taskList, script)
				}
			}
		}

		// check if we have found the target
		for curTIndex, script := range taskList {
			if strings.EqualFold(target, script.ID) {
				GetLogger().WithFields(logrus.Fields{
					"target":    target,
					"scopeVars": scopeVars,
				}).Info("executeTemplate EXECUTE target")
				targetFound = true

				stopReason := script.Stopreasons

				var messageCmdCtrl CtxOutCtrl = CtxOutCtrl{ // define a controll hook, depending on the display comand option
					IgnoreCase: !script.Options.Displaycmd, // we ignore thie message, as long the display command is NOT set
				}

				// check requirements
				canRun, message := checkRequirements(script.Requires)
				if !canRun {
					GetLogger().WithFields(logrus.Fields{
						"target": target,
						"reason": message,
					}).Info("executeTemplate IGNORE because requirements not matching")
					if script.Options.Displaycmd {
						CtxOut(messageCmdCtrl, LabelFY("require"), ValF(message), InfoF("Task-Block "), curTIndex+1, " of ", len(taskList), " skipped")
					}
					// ---- return ExitByRequirement
					continue
				}

				// get the task related variables
				for keyName, variable := range script.Variables {
					SetPH(keyName, HandlePlaceHolder(variable))
					scopeVars[keyName] = variable
				}
				backToDir := ""
				// if working dir is set change to them
				if script.Options.WorkingDir != "" {
					backToDir, _ = dirhandle.Current()
					chDirError := os.Chdir(HandlePlaceHolderWithScope(script.Options.WorkingDir, scopeVars))
					if chDirError != nil {
						manout.Error("Workspace setting seems invalid ", chDirError)
						systools.Exit(10)
					}
				}

				// just the abort flag.
				abort := false

				// experimental usage of needs
				if Experimental {
					// -- NEEDS
					// needs are task, the have to be startet once
					// before we continue.
					// any need can have his own needs they needs to
					// be executed
					if len(script.Needs) > 0 {

						CtxOut(messageCmdCtrl, LabelFY("target"), ValF(target), InfoF("require"), ValF(len(script.Needs)), InfoF("needs. async?"), ValF(runAsync))
						GetLogger().WithField("needs", script.Needs).Debug("Needs for the script")
						if runAsync {
							var needExecs []FutureStack
							for _, needTarget := range script.Needs {
								if TaskRunsAtLeast(needTarget, 1) {
									CtxOut(messageCmdCtrl, LabelFY("need check"), ValF(target), InfoRed("already executed"), ValF(needTarget))
									GetLogger().Debug("need already handled " + needTarget)
								} else {
									GetLogger().Debug("need name should be added " + needTarget)
									CtxOut(messageCmdCtrl, LabelFY("need check"), ValF(target), InfoF("executing"), ValF(needTarget))
									needExecs = append(needExecs, FutureStack{
										AwaitFunc: func(ctx context.Context) interface{} {
											argNeed := ctx.Value(CtxKey{}).(string)
											_, argmap := StringSplitArgs(argNeed, "arg")
											GetLogger().Debug("add need task " + argNeed)
											return executeTemplate(waitGroup, true, runCfg, argNeed, argmap)
										},
										Argument: needTarget})
								}
							}
							futures := ExecFutureGroup(needExecs) // create the futures and start the tasks
							results := WaitAtGroup(futures)       // wait until any task is executed

							GetLogger().WithField("result", results).Debug("needs result")
						} else {
							for _, needTarget := range script.Needs {
								if TaskRunsAtLeast(needTarget, 1) { // do not run needs the already runs
									GetLogger().Debug("need already handled " + needTarget)
								} else {
									_, argmap := StringSplitArgs(needTarget, "arg")
									executeTemplate(waitGroup, false, runCfg, needTarget, argmap)
								}
							}
						}
						CtxOut(LabelFY("target"), ValF(target), InfoF("needs done"))
					}
				} else {
					// NONE experimental usage of needs
					// checking needs
					if len(script.Needs) > 0 {
						GetLogger().WithFields(logrus.Fields{
							"needs": script.Needs,
						}).Info("executeTemplate NEEDS found")
						if runAsync {
							waitHits := 0
							timeOut := script.Options.TimeoutNeeds
							if timeOut < 1 {
								GetLogger().Info("No timeoutNeeds value set. using default of 300000")
								timeOut = 300000 // 5 minutes in milliseconds as default
							} else {
								GetLogger().WithField("timeout", timeOut).Info("timeout for task " + target)
							}
							tickTime := script.Options.TickTimeNeeds
							if tickTime < 1 {
								tickTime = 1000 // 1 second as ticktime
							}
							WaitForTasksDone(script.Needs, time.Duration(timeOut)*time.Millisecond, time.Duration(tickTime)*time.Millisecond, func() bool {
								// still waiting
								waitHits++
								GetLogger().Debug("Waiting for Task be done")
								return true
							}, func() {
								// done

							}, func() {
								// timeout not allowed. hard exit
								GetLogger().Debug("timeout hit")
								manout.Error("Need Timeout", "waiting for a need timed out after ", timeOut, " milliseconds. you may increase timeoutNeeds in Options")
								systools.Exit(1)
							}, func(needTarget string, _ string, args map[string]string) bool {
								if script.Options.NoAutoRunNeeds {
									manout.Error("Need Task not started", "expected task ", target, " not running. autostart disabled")
									systools.Exit(1)
									return false
								}
								GetLogger().WithFields(logrus.Fields{
									"needs":   script.Needs,
									"current": needTarget,
								}).Info("executeTemplate found a need that is not stated already")
								// stopping for a couple of time
								// need to wait if these other task already started by
								// other options
								time.Sleep(500 * time.Millisecond)
								go executeTemplate(waitGroup, runAsync, runCfg, needTarget, args)
								return true
							})
						} else {
							// run needs in a sequence
							for _, targetNeed := range script.Needs {
								var args map[string]string = make(map[string]string) // no supported usage right now
								executionCode := executeTemplate(waitGroup, runAsync, runCfg, targetNeed, args)
								if executionCode != ExitOk {
									manout.Error("Need Task Error", "expected returncode ", ExitOk, " but got exit Code", executionCode)
									systools.Exit(1)
								}
							}
						}
					}
				} // end of experimental switch

				// targets that should be started as well
				// these targets running at the same time
				// so different to scope, we dont need to wait
				// right now until they ends
				var runTargetfutures []Future
				if Experimental {
					runTargetfutures = generateFuturesByTargetListAndExec(script.RunTargets, waitGroup, runCfg)
				} else {
					if len(script.RunTargets) > 0 {
						for _, runTrgt := range script.RunTargets {
							if runAsync {
								go executeTemplate(waitGroup, runAsync, runCfg, runTrgt, scopeVars)
							} else {
								executeTemplate(waitGroup, runAsync, runCfg, runTrgt, scopeVars)
							}
						}
						// workaround til the async runnig is refactored
						// now we need to give the subtask time to run and update the waitgroup
						duration := time.Second
						time.Sleep(duration)
					}
				}

				// check if we have script lines.
				// if not, we need at least to check
				// 'now' listener
				if len(script.Script) < 1 {
					GetLogger().Debug("no script lines defined. run listener anyway")
					listenerWatch(script, target, "", waitGroup, runAsync, runCfg)
					// workaround til the async runnig is refactored
					// now we need to give the subtask time to run and update the waitgroup
					duration := time.Second
					time.Sleep(duration)
				}

				// preparing codelines by execute second level commands
				// that can affect the whole script
				abort, returnCode, _ = TryParse(script.Script, func(codeLine string) (bool, int) {
					lineAbort, lineExitCode := lineExecuter(waitGroup, runAsync, stopReason, runCfg, colorCode, bgCode, codeLine, target, scopeVars, script)
					return lineExitCode, lineAbort
				})
				if abort {
					GetLogger().Debug("abort reason found ")
				}

				// waitin until the any target that runns also is done
				if Experimental && len(runTargetfutures) > 0 {
					CtxOut(messageCmdCtrl, LabelFY("wait targets"), "waiting until beside running targets are done")
					trgtRes := WaitAtGroup(runTargetfutures)
					CtxOut(messageCmdCtrl, LabelFY("wait targets"), "waiting done", trgtRes)
				}
				// next are tarets they runs afterwards the regular
				// script os done
				GetLogger().WithFields(logrus.Fields{
					"current-target": target,
					"nexts":          script.Next,
				}).Debug("executeTemplate next definition")
				if Experimental {
					nextfutures := generateFuturesByTargetListAndExec(script.Next, waitGroup, runCfg)
					nextRes := WaitAtGroup(nextfutures)
					CtxOut(messageCmdCtrl, LabelFY("wait next"), "waiting done", nextRes)

				} else {
					for _, nextTarget := range script.Next {
						if script.Options.Displaycmd {
							CtxOut(LabelFY("next"), InfoF(nextTarget))
						}
						/* ---- something is wrong with my logic dependig execution not in a sequence (useWaitGroup == true)
						if useWaitGroup {
							go executeTemplate(waitGroup, useWaitGroup, runCfg, nextTarget)

						} else {
							executeTemplate(waitGroup, useWaitGroup, runCfg, nextTarget)
						}*/

						// for now we execute without a waitgroup
						executeTemplate(waitGroup, runAsync, runCfg, nextTarget, scopeVars)
					}
				}

				//return returnCode
				// back to old dir if workpace usage was set
				if backToDir != "" {
					os.Chdir(backToDir)
				}
			}

		}
		if !targetFound {
			CtxOut(manout.MessageCln(manout.ForeYellow, "target not defined: ", manout.ForeWhite, target))
			GetLogger().Error("Target can not be found: ", target)
		}

		GetLogger().WithFields(logrus.Fields{
			"target": target,
		}).Info("executeTemplate. target do not contains tasks")
		return returnCode
	}
	return ExitNoCode
}

func defaultString(line string, defaultString string) string {
	if line == "" {
		return defaultString
	}
	return line
}

func handleFileImportsToVars(imports []string) {
	for _, filenameFull := range imports {
		var keyname string
		parts := strings.Split(filenameFull, " ")
		filename := parts[0]
		if len(parts) > 1 {
			keyname = parts[1]
		}

		dirhandle.FileTypeHandler(filename, func(jsonBaseName string) {
			GetLogger().Debug("loading json File as second level variables:", filename)
			if keyname == "" {
				keyname = jsonBaseName
			}
			ImportDataFromJSONFile(keyname, filename)

		}, func(yamlBaseName string) {
			GetLogger().Debug("loading yaml File: as second level variables", filename)
			if keyname == "" {
				keyname = yamlBaseName
			}
			ImportDataFromYAMLFile(keyname, filename)

		}, func(path string, err error) {
			GetLogger().Errorln("file not exists:", err)
			manout.Error("varibales file not exists:", path, err)
			systools.Exit(1)
		})
	}
}
