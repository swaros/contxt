package cmdhandle

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/context/dirhandle"
	"github.com/swaros/contxt/context/output"

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
		fmt.Println(output.MessageCln(output.ForeCyan, "[SHARED loop]"))
		for _, shared := range template.Config.Use {
			fmt.Println(output.MessageCln(output.ForeCyan, "[...SHARED CONTXT >>>][", output.ForeBlue, shared+"] "))
			externalPath := HandleUsecase(shared)
			GetLogger().WithField("path", externalPath).Info("shared contxt location")
			currentDir, _ := dirhandle.Current()
			os.Chdir(externalPath)
			for _, runTarget := range allTargets {
				fmt.Println(output.MessageCln(output.ForeCyan, output.ForeGreen, runTarget, output.ForeYellow, " [ external:", output.ForeWhite, externalPath, output.ForeYellow, "] ", output.ForeDarkGrey, templatePath))
				RunTargets(runTarget, false)
				fmt.Println(output.MessageCln(output.ForeCyan, output.ForeBlue, shared+"] ", output.ForeGreen, runTarget, " DONE"))
			}
			os.Chdir(currentDir)
		}
		fmt.Println(output.MessageCln(output.ForeCyan, "[SHARED done]"))
	}
	GetLogger().WithField("targets", allTargets).Info("  SHARED DONE run targets...")
}

// RunTargets executes multiple targets
func RunTargets(targets string, sharedRun bool) {

	SetPH("CTX_TARGETS", targets)

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
			output.ColorEnabled = false
		}
	}

	if template.Config.Loglevel != "" {
		setLogLevelByString(template.Config.Loglevel)
	}

	var wg sync.WaitGroup

	// handle all shared usages
	/*
		if len(template.Config.Use) > 0 {
			GetLogger().WithField("uses", template.Config.Use).Info("found external dependecy")
			for _, shared := range template.Config.Use {
				externalPath := HandleUsecase(shared)
				GetLogger().WithField("path", externalPath).Info("shared contxt location")
				currentDir, _ := dirhandle.Current()
				os.Chdir(externalPath)
				for _, runTarget := range allTargets {
					fmt.Println(output.MessageCln(output.ForeCyan, "[SHARED CONTXT ", output.BoldTag, shared, "] ", runTarget, " ", output.ForeWhite, templatePath))
					RunTargets(runTarget)
				}
				os.Chdir(currentDir)
			}
		}
	*/

	// handle all imports
	if len(template.Config.Imports) > 0 {
		GetLogger().WithField("Import", template.Config.Imports).Info("import second level vars")
		handleFileImportsToVars(template.Config.Imports)
	} else {
		GetLogger().Info("No second level Variables defined")
	}

	if !runSequencially {
		// run in thread
		for _, runTarget := range allTargets {
			SetPH("CTX_TARGET", runTarget)
			wg.Add(1)
			fmt.Println(output.MessageCln(output.ForeBlue, "[exec:async] ", output.BoldTag, runTarget, " ", output.ForeWhite, templatePath))
			go ExecuteTemplateWorker(&wg, true, runTarget, template)
		}
		wg.Wait()
	} else {
		// trun one by one
		for _, runTarget := range allTargets {
			SetPH("CTX_TARGET", runTarget)
			fmt.Println(output.MessageCln(output.ForeBlue, "[exec:seq] ", output.BoldTag, runTarget, " ", output.ForeWhite, templatePath))
			exitCode := ExecPathFile(&wg, false, template, runTarget)
			GetLogger().WithField("exitcode", exitCode).Info("RunTarget [Sequencially runmode] done with exitcode")
		}
	}
	fmt.Println(output.MessageCln(output.ForeBlue, "[done] ", output.BoldTag, targets))
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

func checkRequirements(require configure.Require) (bool, string) {
	// check operating system
	if require.System != "" {
		match := StringMatchTest(require.System, configure.GetOs())
		if !match {
			return false, "operating system '" + configure.GetOs() + "' is not matching with '" + require.System + "'"
		}
	}

	// check file exists
	for _, fileExists := range require.Exists {
		fileExists = handlePlaceHolder(fileExists)
		fexists, err := dirhandle.Exists(fileExists)
		GetLogger().WithFields(logrus.Fields{
			"path":   fileExists,
			"result": fexists,
		}).Debug("path exists? result=true means valid for require")
		if err != nil || !fexists {

			return false, "required file (" + fileExists + ") not found "
		}
	}

	// check file not exists
	for _, fileNotExists := range require.NotExists {
		fileNotExists = handlePlaceHolder(fileNotExists)
		fexists, err := dirhandle.Exists(fileNotExists)
		GetLogger().WithFields(logrus.Fields{
			"path":   fileNotExists,
			"result": fexists,
		}).Debug("path NOT exists? result=true means not valid for require")
		if err != nil || fexists {
			return false, "unexpected file (" + fileNotExists + ")  found "
		}
	}
	// check environment variable is set

	for name, pattern := range require.Environment {
		envVar, envExists := os.LookupEnv(name)
		if !envExists || !StringMatchTest(pattern, envVar) {
			if envExists {
				return false, "environment variable[" + name + "] not matching with " + pattern
			}
			return false, "environment variable[" + name + "] not exists"
		}
	}

	// check variables
	for name, pattern := range require.Variables {
		defVar, defExists := GetPHExists(name)
		if !defExists || !StringMatchTest(pattern, defVar) {
			if defExists {
				return false, "runtime variable[" + name + "] not matching with " + pattern
			}
			return false, "runtime variable[" + name + "] not exists "
		}
	}

	return true, ""
}

// StringMatchTest test a pattern and a value.
// in this example: myvar: "=hello"
// the patter is "=hello" and the value should be "hello" for a match
func StringMatchTest(pattern, value string) bool {
	first := pattern
	maybeMatch := value
	if len(pattern) > 1 {
		maybeMatch = pattern[1:]
		first = pattern[0:1]
	}
	switch first {
	case "?":
		GetLogger().WithFields(logrus.Fields{"pattern": pattern, "value": value, "check": first, "result": (value != "")}).Debug("check anything then empty")
		return (value != "")
	case "=":
		GetLogger().WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch == value)}).Debug("check equal")
		return (maybeMatch == value)
	case "!":
		GetLogger().WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch != value)}).Debug("check not equal")
		return (maybeMatch != value)
	case ">":
		GetLogger().WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch < value)}).Debug("check greather then")
		return (value > maybeMatch)
	case "<":
		GetLogger().WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch > value)}).Debug("check lower then")
		return (value < maybeMatch)
	case "*":
		GetLogger().WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (value != "")}).Debug("check empty *")
		return (value != "")
	default:
		GetLogger().WithFields(logrus.Fields{"pattern": pattern, "value": value, "check": first, "result": (pattern == value)}).Debug("default: check equal against plain values")
		return (pattern == value)
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
					fmt.Println(output.MessageCln(output.ForeCyan, "[trigger]\t", output.ForeYellow, triggerMessage, output.Dim, " ", logLine))
				}

				// did this trigger something?
				someReactionTriggered := false
				// extract action
				actionDef := configure.Action(listener.Action)

				// checking script
				if len(actionDef.Script) > 0 {
					someReactionTriggered = true
					for _, triggerScript := range actionDef.Script {
						GetLogger().WithFields(logrus.Fields{
							"cmd": triggerScript,
						}).Debug("TRIGGER SCRIPT ACTION")
						lineExecuter(waitGroup, useWaitGroup, script.Stopreasons, runCfg, "93", "46", triggerScript, target, script)
					}

				}

				if actionDef.Target != "" {
					someReactionTriggered = true
					GetLogger().WithFields(logrus.Fields{
						"target": actionDef.Target,
					}).Debug("TRIGGER ACTION")

					if script.Options.Displaycmd {
						fmt.Println(output.MessageCln(output.ForeCyan, "[trigger]\t ", output.ForeGreen, "target:", output.ForeLightGreen, actionDef.Target))
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

					if useWaitGroup {
						GetLogger().WithFields(logrus.Fields{
							"target": actionDef.Target,
						}).Info("RUN ASYNC")

						go executeTemplate(waitGroup, useWaitGroup, runCfg, actionDef.Target)

					} else {
						GetLogger().WithFields(logrus.Fields{
							"target": actionDef.Target,
						}).Info("RUN SEQUENCE")
						executeTemplate(waitGroup, useWaitGroup, runCfg, actionDef.Target)
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

func lineExecuter(waitGroup *sync.WaitGroup, useWaitGroup bool, stopReason configure.Trigger, runCfg configure.RunConfig, colorCode, bgCode, codeLine, target string, script configure.Task) (int, bool) {
	panelSize := 12
	if script.Options.Panelsize > 0 {
		panelSize = script.Options.Panelsize
	}
	var mainCommand = defaultString(script.Options.Maincmd, DefaultCommandFallBack)
	if configure.GetOs() == "windows" {
		mainCommand = defaultString(script.Options.Maincmd, DefaultCommandFallBackWindows)
	}
	replacedLine := HandlePlaceHolder(codeLine)
	if script.Options.Displaycmd {
		fmt.Println(output.MessageCln(output.Dim, output.ForeYellow, " [cmd] ", output.ResetDim, output.ForeCyan, target, output.ForeDarkGrey, " \t :> ", output.BoldTag, output.ForeBlue, replacedLine))
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
				format := handlePlaceHolder(script.Options.Format)
				fomatedOutStr := output.Message(fmt.Sprintf(format, target))
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
				fmt.Println(output.MessageCln(output.ForeLightCyan, " STOP-HIT ", output.ForeWhite, output.BackBlue, message))
			}
			return false
		}
		return true
	}, func(process *os.Process) {
		pidStr := fmt.Sprintf("%d", process.Pid)
		SetPH("RUN.PID", pidStr)
		SetPH("RUN."+target+".PID", pidStr)
		if script.Options.Displaycmd {
			fmt.Println(output.MessageCln(output.ForeYellow, " [pid] ", output.ForeBlue, process.Pid))
		}
	})
	if execErr != nil {
		if script.Options.Displaycmd {
			fmt.Println(output.MessageCln(output.ForeRed, "execution error: ", output.BackRed, output.ForeWhite, execErr))
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
			fmt.Println(output.MessageCln(output.ForeYellow, "NOTE!\t", output.BackLightYellow, output.ForeDarkGrey, " a script execution was failing. no stopreason is set so execution will continued "))
			fmt.Println(output.MessageCln("\t", output.BackLightYellow, output.ForeDarkGrey, " if this is expected you can ignore this message.                                 "))
			fmt.Println(output.MessageCln("\t", output.BackLightYellow, output.ForeDarkGrey, " but you should handle error cases                                                "))
			fmt.Println("\ttarget :\t", output.MessageCln(output.ForeYellow, target))
			fmt.Println("\tcommand:\t", output.MessageCln(output.ForeYellow, codeLine))

		} else {
			errMsg := " = exit code from command: "
			lastMessage := output.MessageCln(output.BackRed, output.ForeYellow, realExitCode, output.CleanTag, output.ForeLightRed, errMsg, output.ForeWhite, codeLine)
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

func executeTemplate(waitGroup *sync.WaitGroup, useWaitGroup bool, runCfg configure.RunConfig, target string) int {
	if useWaitGroup {
		waitGroup.Add(1)
		GetLogger().WithFields(logrus.Fields{
			"waitgroup": waitGroup,
		}).Debug("starting async")
		defer waitGroup.Done()
	}
	// check if task is already running

	if TaskRunning(target) {
		GetLogger().WithField("task", target).Warning("task would be triggered again while is already running. IGNORED")
		return ExitAlreadyRunning
	}
	incTaskCount(target)
	defer incTaskDoneCount(target)

	GetLogger().WithFields(logrus.Fields{
		"target": target,
	}).Info("executeTemplate LOOKING for target")

	if len(runCfg.Task) > 0 {

		// main variables
		for keyName, variable := range runCfg.Config.Variables {
			SetIfNotExists(keyName, HandlePlaceHolder(variable))
		}

		colorCode := systools.CreateColorCode()
		bgCode := systools.CurrentBgColor
		SetPH("RUN.TARGET", target)
		targetFound := false
		//for _, script := range runCfg.Task {
		script := mergeTargets(target, runCfg)
		// check if we have found the target
		if strings.EqualFold(target, script.ID) {
			GetLogger().WithFields(logrus.Fields{
				"target": target,
			}).Info("executeTemplate EXECUTE target")
			targetFound = true
			// first get the task related variables
			for keyName, variable := range script.Variables {
				SetPH(keyName, HandlePlaceHolder(variable))
			}

			stopReason := script.Stopreasons
			// check requirements
			canRun, message := checkRequirements(script.Requires)
			if !canRun {
				GetLogger().WithFields(logrus.Fields{
					"target": target,
					"reason": message,
				}).Info("executeTemplate IGNORE because requirements not matching")
				if script.Options.Displaycmd {
					fmt.Println(output.MessageCln(output.ForeYellow, " [require] ", output.ForeBlue, message))
				}
				return ExitByRequirement
			}

			// parsing codelines
			returnCode := ExitOk
			abort := false

			// checking needs
			if len(script.Needs) > 0 {
				GetLogger().WithFields(logrus.Fields{
					"needs": script.Needs,
				}).Info("executeTemplate NEEDS found")
				if useWaitGroup {
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
						output.Error("Need Timeout", "waiting for a need timed out after ", timeOut, " milliseconds. you may increase timeoutNeeds in Options")
						os.Exit(1)
					}, func(needTarget string) bool {
						if script.Options.NoAutoRunNeeds {
							output.Error("Need Task not started", "expected task ", target, " not running. autostart disbabled")
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
						go executeTemplate(waitGroup, useWaitGroup, runCfg, needTarget)
						return true
					})
				} else {
					// run needs in a sequence
					for _, targetNeed := range script.Needs {
						executionCode := executeTemplate(waitGroup, useWaitGroup, runCfg, targetNeed)
						if executionCode != ExitOk {
							output.Error("Need Task Error", "expected returncode ", ExitOk, " but got exit Code", executionCode)
							os.Exit(1)
						}
					}
				}
			}

			// targets that should be started as well
			if len(script.RunTargets) > 0 {
				for _, runTrgt := range script.RunTargets {
					if useWaitGroup {
						go executeTemplate(waitGroup, useWaitGroup, runCfg, runTrgt)
					} else {
						executeTemplate(waitGroup, useWaitGroup, runCfg, runTrgt)
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
				listenerWatch(script, target, "", waitGroup, useWaitGroup, runCfg)
				// workaround til the async runnig is refactored
				// now we need to give the subtask time to run and update the waitgroup
				duration := time.Second
				time.Sleep(duration)
			}

			// preparing codelines by execute second level commands
			// that can affect the whole script
			abort, returnCode, _ = TryParse(script.Script, func(codeLine string) (bool, int) {
				lineAbort, lineExitCode := lineExecuter(waitGroup, useWaitGroup, stopReason, runCfg, colorCode, bgCode, codeLine, target, script)
				return lineExitCode, lineAbort
			})
			if abort {
				GetLogger().Debug("abort reason found ")
			}
			/*
				for _, codeLine := range script.Script {
					if !abort {
						returnCode, abort = lineExecuter(waitGroup, useWaitGroup, stopReason, runCfg, colorCode, bgCode, codeLine, target, script)
					}

				}*/
			// executes next targets if there some defined
			GetLogger().WithFields(logrus.Fields{
				"current-target": target,
				"nexts":          script.Next,
			}).Debug("executeTemplate next definition")
			for _, nextTarget := range script.Next {
				if script.Options.Displaycmd {
					fmt.Println(output.MessageCln(output.ForeYellow, " [next] ", output.ForeBlue, nextTarget))
				}
				/* ---- something is wrong with my logic dependig execution not in a sequence (useWaitGroup == true)
				if useWaitGroup {
					go executeTemplate(waitGroup, useWaitGroup, runCfg, nextTarget)

				} else {
					executeTemplate(waitGroup, useWaitGroup, runCfg, nextTarget)
				}*/

				// for now we execute without a waitgroup
				executeTemplate(waitGroup, useWaitGroup, runCfg, nextTarget)
			}
			return returnCode
		}

		if !targetFound {
			fmt.Println(output.MessageCln(output.ForeYellow, "target not defined: ", output.ForeWhite, target))
			GetLogger().Error("Target can not be found: ", target)
		}
	}
	GetLogger().WithFields(logrus.Fields{
		"target": target,
	}).Info("executeTemplate. target do not contains tasks")

	return ExitNoCode
}

func defaultString(line string, defaultString string) string {
	if line == "" {
		return defaultString
	}
	return line
}

/*
func stringContains(findInHere string, matches []string) bool {
	for _, check := range matches {
		if check != "" && strings.Contains(findInHere, check) {
			return true
		}
	}
	return false
}*/

func checkReason(checkReason configure.Trigger, output string) (bool, string) {
	GetLogger().WithFields(logrus.Fields{
		"contains":   checkReason.OnoutContains,
		"onError":    checkReason.Onerror,
		"onLess":     checkReason.OnoutcountLess,
		"onMore":     checkReason.OnoutcountMore,
		"testing-at": output,
	}).Debug("Checking Trigger")

	var message = ""
	if checkReason.Now {
		message = "reason now match always"
		return true, message
	}
	if checkReason.OnoutcountLess > 0 && checkReason.OnoutcountLess > len(output) {
		message = fmt.Sprint("reason match output len (", len(output), ") is less then ", checkReason.OnoutcountLess)
		return true, message
	}
	if checkReason.OnoutcountMore > 0 && checkReason.OnoutcountMore < len(output) {
		message = fmt.Sprint("reason match output len (", len(output), ") is more then ", checkReason.OnoutcountMore)
		return true, message
	}

	for _, checkText := range checkReason.OnoutContains {
		if checkText != "" && strings.Contains(output, checkText) {
			message = fmt.Sprint("reason match because output contains ", checkText)
			return true, message
		}
		if checkText != "" {
			GetLogger().WithFields(logrus.Fields{
				"check": checkText,
				"with":  output,
				"from":  checkReason.OnoutContains,
			}).Debug("OnoutContains NO MATCH")
		}
	}

	return false, message
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
			output.Error("varibales file not exists:", err)
			os.Exit(1)
		})
	}
}
