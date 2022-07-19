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
	template, templatePath, exists := GetTemplate()
	if !exists {
		return
	}

	if template.Config.Loglevel != "" {
		setLogLevelByString(template.Config.Loglevel)
	}

	GetLogger().WithField("targets", allTargets).Info("SHARED START run targets...")

	// handle all shared usages
	if len(template.Config.Use) > 0 {
		GetLogger().WithField("uses", template.Config.Use).Info("found external dependecy")
		fmt.Println(manout.MessageCln(manout.ForeCyan, "[SHARED loop]"))
		for _, shared := range template.Config.Use {
			fmt.Println(manout.MessageCln(manout.ForeCyan, "[SHARED CONTXT][", manout.ForeBlue, shared, manout.ForeCyan, "] "))
			externalPath := HandleUsecase(shared)
			GetLogger().WithField("path", externalPath).Info("shared contxt location")
			currentDir, _ := dirhandle.Current()
			os.Chdir(externalPath)
			for _, runTarget := range allTargets {
				fmt.Println(manout.MessageCln(manout.ForeCyan, manout.ForeGreen, runTarget, manout.ForeYellow, " [ external:", manout.ForeWhite, externalPath, manout.ForeYellow, "] ", manout.ForeDarkGrey, templatePath))
				RunTargets(runTarget, false)
				fmt.Println(manout.MessageCln(manout.ForeCyan, "["+manout.ForeBlue, shared+"] ", manout.ForeGreen, runTarget, " DONE"))
			}
			os.Chdir(currentDir)
		}
		fmt.Println(manout.MessageCln(manout.ForeCyan, "[SHARED done]"))
	}
	GetLogger().WithField("targets", allTargets).Info("  SHARED DONE run targets...")
}

// RunTargets executes multiple targets
// the targets string can have multiple targets
// seperated by comma
func RunTargets(targets string, sharedRun bool) {

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
	template, templatePath, exists := GetTemplate()
	GetLogger().WithField("targets", allTargets).Info("run targets...")
	var runSequencially = false
	if exists {
		runSequencially = template.Config.Sequencially
		if template.Config.Coloroff {
			manout.ColorEnabled = false
		}
	}

	if template.Config.Loglevel != "" {
		setLogLevelByString(template.Config.Loglevel)
	}

	var wg sync.WaitGroup

	// handle all imports
	if len(template.Config.Imports) > 0 {
		GetLogger().WithField("Import", template.Config.Imports).Info("import second level vars")
		handleFileImportsToVars(template.Config.Imports)
	} else {
		GetLogger().Info("No second level Variables defined")
	}

	// experimental usage of taskrunner
	if Experimental {
		if runSequencially {
			for _, trgt := range allTargets {
				ExecPathFile(&wg, !runSequencially, template, trgt)
			}
		} else {
			var futuresExecs []FutureStack
			for _, trgt := range allTargets {
				futuresExecs = append(futuresExecs, FutureStack{
					AwaitFunc: func(ctx context.Context) interface{} {
						ctxTarget := ctx.Value(CtxKey{}).(string)
						return ExecPathFile(&wg, !runSequencially, template, ctxTarget)
					},
					Argument: trgt,
				})
			}
			futures := ExecFutureGroup(futuresExecs)
			WaitAtGroup(futures)
		}

	} else {
		// NONE experimental usage of taskrunner
		if !runSequencially {
			// run in thread
			for _, runTarget := range allTargets {
				SetPH("CTX_TARGET", runTarget)
				wg.Add(1)
				fmt.Println(manout.MessageCln(manout.ForeBlue, "[exec:async] ", manout.BoldTag, runTarget, " ", manout.ForeWhite, templatePath))
				go ExecuteTemplateWorker(&wg, true, runTarget, template)
			}
			wg.Wait()
		} else {
			// trun one by one
			for _, runTarget := range allTargets {
				SetPH("CTX_TARGET", runTarget)
				fmt.Println(manout.MessageCln(manout.ForeBlue, "[exec:seq] ", manout.BoldTag, runTarget, " ", manout.ForeWhite, templatePath))
				exitCode := ExecPathFile(&wg, false, template, runTarget)
				GetLogger().WithField("exitcode", exitCode).Info("RunTarget [Sequencially runmode] done with exitcode")
			}
		}
	}
	fmt.Println(manout.MessageCln(manout.ForeBlue, "[done] ", manout.BoldTag, targets))
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
			listenReason := listener.Trigger
			triggerFound, triggerMessage := checkReason(listenReason, logLine)
			if triggerFound {
				SetPH("RUN."+target+".LOG.HIT", logLine)
				if script.Options.Displaycmd {
					fmt.Println(manout.MessageCln(manout.ForeCyan, "[trigger]\t", manout.ForeYellow, triggerMessage, manout.Dim, " ", logLine))
				}

				// did this trigger something?
				someReactionTriggered := false
				// extract action
				actionDef := configure.Action(listener.Action)

				// checking script
				if len(actionDef.Script) > 0 {
					someReactionTriggered = true
					var dummyArgs map[string]string = make(map[string]string)
					for _, triggerScript := range actionDef.Script {
						GetLogger().WithFields(logrus.Fields{
							"cmd": triggerScript,
						}).Debug("TRIGGER SCRIPT ACTION")
						lineExecuter(waitGroup, useWaitGroup, script.Stopreasons, runCfg, "93", "46", triggerScript, target, dummyArgs, script)
					}

				}

				if actionDef.Target != "" {
					someReactionTriggered = true
					GetLogger().WithFields(logrus.Fields{
						"target": actionDef.Target,
					}).Debug("TRIGGER ACTION")

					if script.Options.Displaycmd {
						fmt.Println(manout.MessageCln(manout.ForeCyan, "[trigger]\t ", manout.ForeGreen, "target:", manout.ForeLightGreen, actionDef.Target))
					}

					hitKeyTargets := "RUN.LISTENER." + target + ".HIT.TARGETS"
					lastHitTargets := GetPH(hitKeyTargets)
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

func lineExecuter(
	waitGroup *sync.WaitGroup,
	useWaitGroup bool,
	stopReason configure.Trigger,
	runCfg configure.RunConfig,
	colorCode, bgCode,
	codeLine,
	target string,
	arguments map[string]string,
	script configure.Task) (int, bool) {
	panelSize := 12
	if script.Options.Panelsize > 0 {
		panelSize = script.Options.Panelsize
	}
	var mainCommand = defaultString(script.Options.Maincmd, DefaultCommandFallBack)
	if configure.GetOs() == "windows" {
		mainCommand = defaultString(script.Options.Maincmd, DefaultCommandFallBackWindows)
	}
	replacedLine := HandlePlaceHolderWithScope(codeLine, arguments)
	if script.Options.Displaycmd {
		fmt.Println(manout.MessageCln(manout.Dim, manout.ForeYellow, " [cmd] ", manout.ResetDim, manout.ForeCyan, target, manout.ForeDarkGrey, " \t :> ", manout.BoldTag, manout.ForeBlue, replacedLine))
	}

	SetPH("RUN.SCRIPT_LINE", replacedLine)

	// here we execute the current script line
	execCode, realExitCode, execErr := ExecuteScriptLine(mainCommand, script.Options.Mainparams, replacedLine, func(logLine string) bool {

		SetPH("RUN."+target+".LOG.LAST", logLine)
		// the watcher
		if script.Listener != nil {

			GetLogger().WithFields(logrus.Fields{
				"cnt":      len(script.Listener),
				"listener": script.Listener,
			}).Debug("CHECK Listener")
			listenerWatch(script, target, logLine, waitGroup, useWaitGroup, runCfg)
		}

		// print the output by configuration
		if !script.Options.Hideout {
			foreColor := defaultString(script.Options.Colorcode, colorCode)
			bgColor := defaultString(script.Options.Bgcolorcode, bgCode)
			labelStr := systools.LabelPrintWithArg(systools.PadStringToR(target+" :", panelSize), foreColor, bgColor, 1)
			if script.Options.Format != "" {
				format := HandlePlaceHolderWithScope(script.Options.Format, script.Variables)
				fomatedOutStr := manout.Message(fmt.Sprintf(format, target))
				labelStr = systools.LabelPrintWithArg(fomatedOutStr, foreColor, bgColor, 1)
			}

			outStr := systools.LabelPrintWithArg(logLine, colorCode, "39", 2)
			if script.Options.Stickcursor {
				fmt.Print("\033[G\033[K")
			}
			// prints the codeline
			fmt.Println(labelStr, outStr)
			if script.Options.Stickcursor {
				fmt.Print("\033[A")

			}
		}
		// do we found a defined reason to stop execution
		stopReasonFound, message := checkReason(stopReason, logLine)
		if stopReasonFound {
			if script.Options.Displaycmd {
				fmt.Println(manout.MessageCln(manout.ForeLightCyan, " STOP-HIT ", manout.ForeWhite, manout.BackBlue, message))
			}
			return false
		}
		return true
	}, func(process *os.Process) {
		pidStr := fmt.Sprintf("%d", process.Pid)
		SetPH("RUN.PID", pidStr)
		SetPH("RUN."+target+".PID", pidStr)
		if script.Options.Displaycmd {
			fmt.Println(manout.MessageCln(manout.ForeYellow, " [pid] ", manout.ForeBlue, process.Pid))
		}
	})
	if execErr != nil {
		if script.Options.Displaycmd {
			fmt.Println(manout.MessageCln(manout.ForeRed, "execution error: ", manout.BackRed, manout.ForeWhite, execErr))
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
			fmt.Println(manout.MessageCln(manout.ForeYellow, "NOTE!\t", manout.BackLightYellow, manout.ForeDarkGrey, " a script execution was failing. no stopreason is set so execution will continued "))
			fmt.Println(manout.MessageCln("\t", manout.BackLightYellow, manout.ForeDarkGrey, " if this is expected you can ignore this message.                                 "))
			fmt.Println(manout.MessageCln("\t", manout.BackLightYellow, manout.ForeDarkGrey, " but you should handle error cases                                                "))
			fmt.Println("\ttarget :\t", manout.MessageCln(manout.ForeYellow, target))
			fmt.Println("\tcommand:\t", manout.MessageCln(manout.ForeYellow, codeLine))

		} else {
			errMsg := " = exit code from command: "
			lastMessage := manout.MessageCln(manout.BackRed, manout.ForeYellow, realExitCode, manout.CleanTag, manout.ForeLightRed, errMsg, manout.ForeWhite, codeLine)
			fmt.Println("\t Exit ", lastMessage)
			fmt.Println()
			fmt.Println("\t check the command. if this command can fail you may fit the execution rules. see options:")
			fmt.Println("\t you may disable a hard exit on error by setting ignoreCmdError: true")
			fmt.Println("\t if you do so, a Note will remind you, that a error is happend in this case.")
			fmt.Println()
			GetLogger().Error("runtime error:", execErr, "exit", realExitCode)
			os.Exit(realExitCode)
			// returns the error code
			return ExitCmdError, true
		}
	case ExitOk:
		return ExitOk, false
	}
	return ExitNoCode, true
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
		colorCode := systools.CreateColorCode()
		bgCode := systools.CurrentBgColor

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
		for _, script := range taskList {
			if strings.EqualFold(target, script.ID) {
				GetLogger().WithFields(logrus.Fields{
					"target":    target,
					"scopeVars": scopeVars,
				}).Info("executeTemplate EXECUTE target")
				targetFound = true

				stopReason := script.Stopreasons
				// check requirements
				canRun, message := checkRequirements(script.Requires)
				if !canRun {
					GetLogger().WithFields(logrus.Fields{
						"target": target,
						"reason": message,
					}).Info("executeTemplate IGNORE because requirements not matching")
					if script.Options.Displaycmd {
						fmt.Println(manout.MessageCln(manout.ForeYellow, " [require] ", manout.ForeBlue, message))
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
						os.Exit(10)
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

						GetLogger().WithField("needs", script.Needs).Debug("Needs for the script")
						if runAsync {
							var needExecs []FutureStack
							for _, needTarget := range script.Needs {
								GetLogger().Debug("need name should be added " + needTarget)
								needExecs = append(needExecs, FutureStack{
									AwaitFunc: func(ctx context.Context) interface{} {
										argNeed := ctx.Value(CtxKey{}).(string)
										_, argmap := StringSplitArgs(argNeed, "arg")
										GetLogger().Debug("add need task " + argNeed)
										return executeTemplate(waitGroup, true, runCfg, argNeed, argmap)
									},
									Argument: needTarget})

							}
							futures := ExecFutureGroup(needExecs)
							results := WaitAtGroup(futures)
							GetLogger().WithField("result", results).Debug("needs result")
						} else {
							for _, needTarget := range script.Needs {
								_, argmap := StringSplitArgs(needTarget, "arg")
								executeTemplate(waitGroup, false, runCfg, needTarget, argmap)
							}
						}

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
								os.Exit(1)
							}, func(needTarget string, _ string, args map[string]string) bool {
								if script.Options.NoAutoRunNeeds {
									manout.Error("Need Task not started", "expected task ", target, " not running. autostart disabled")
									os.Exit(1)
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
									os.Exit(1)
								}
							}
						}
					}
				} // end of experimental switch

				// targets that should be started as well
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
				// executes next targets if there some defined
				GetLogger().WithFields(logrus.Fields{
					"current-target": target,
					"nexts":          script.Next,
				}).Debug("executeTemplate next definition")
				for _, nextTarget := range script.Next {
					if script.Options.Displaycmd {
						fmt.Println(manout.MessageCln(manout.ForeYellow, " [next] ", manout.ForeBlue, nextTarget))
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

				//return returnCode
				// back to old dir if workpace usage was set
				if backToDir != "" {
					os.Chdir(backToDir)
				}
			}

		}
		if !targetFound {
			fmt.Println(manout.MessageCln(manout.ForeYellow, "target not defined: ", manout.ForeWhite, target))
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
			os.Exit(1)
		})
	}
}
