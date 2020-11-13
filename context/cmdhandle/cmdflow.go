package cmdhandle

import (
	"fmt"
	"os"
	"strings"
	"sync"

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
)

// RunTargets executes multiple targets
func RunTargets(targets string) {
	allTargets := strings.Split(targets, ",")
	template, templatePath, exists := GetTemplate()

	var runSequencially = false
	if exists {
		runSequencially = template.Config.Sequencially
		if template.Config.Coloroff {
			output.ColorEnabled = false
		}
	}

	if len(template.Config.Imports) > 0 {
		handleFileImportsToVars(template.Config.Imports)
	}

	if template.Config.Loglevel != "" {
		setLogLevelByString(template.Config.Loglevel)
	}

	var wg sync.WaitGroup
	if runSequencially == false {
		// run in thread
		for _, runTarget := range allTargets {
			wg.Add(1)
			fmt.Println(output.MessageCln(output.ForeBlue, "[exec:async] ", output.BoldTag, runTarget, " ", output.ForeWhite, templatePath))
			go ExecuteTemplateWorker(&wg, true, runTarget, template)
		}
		wg.Wait()
	} else {
		// trun one by one
		fmt.Println("Sequencially runmode")
		for _, runTarget := range allTargets {
			fmt.Println(output.MessageCln(output.ForeBlue, "[exec:seq] ", output.BoldTag, runTarget, " ", output.ForeWhite, templatePath))
			ExecPathFile(&wg, false, template, runTarget)
		}
	}
	fmt.Println(output.MessageCln(output.ForeBlue, "[done] ", output.BoldTag, targets))
}

func setLogLevelByString(loglevel string) {
	switch strings.ToUpper(loglevel) {
	case "DEBUG":
		GetLogger().SetLevel(logrus.DebugLevel)
		break
	case "WARN":
		GetLogger().SetLevel(logrus.WarnLevel)
		break
	case "ERROR":
		GetLogger().SetLevel(logrus.ErrorLevel)
		break
	case "FATAL":
		GetLogger().SetLevel(logrus.FatalLevel)
		break
	case "TRACE":
		GetLogger().SetLevel(logrus.TraceLevel)
		break
	case "INFO":
		GetLogger().SetLevel(logrus.InfoLevel)
		break
	default:
		GetLogger().Fatal("unkown log level in config section: ", loglevel)
	}

}

func checkRequirements(require configure.Require) (bool, string) {
	// check file exists
	for _, fileExists := range require.Exists {
		fexists, err := dirhandle.Exists(fileExists)
		if err != nil || fexists == false {

			return false, "required file (" + fileExists + ") not found "
		}
	}

	// check file not exists
	for _, fileNotExists := range require.NotExists {
		fexists, err := dirhandle.Exists(fileNotExists)
		if err != nil || fexists == true {
			return false, "unexpected file (" + fileNotExists + ")  found "
		}
	}
	// check environment variable is set

	for name, value := range require.Environment {
		envVar := os.Getenv(name)
		if envVar != value {
			return false, "environment variable[" + name + "] not matching with " + value
		}
	}

	// check variables
	for name, value := range require.Variables {
		defVar := GetPH(name)
		if defVar != value {
			return false, "runtime variable variable[" + name + "] not matching with " + value
		}
	}

	return true, ""
}

func executeTemplate(waitGroup *sync.WaitGroup, useWaitGroup bool, runCfg configure.RunConfig, target string) int {
	if useWaitGroup {
		waitGroup.Add(1)
		GetLogger().WithFields(logrus.Fields{
			"waitgroup": waitGroup,
		}).Debug("starting async")
		defer waitGroup.Done()
	}

	GetLogger().WithFields(logrus.Fields{
		"target": target,
	}).Info("LOOKING for target")

	if len(runCfg.Task) > 0 {

		// main variables
		for keyName, variable := range runCfg.Config.Variables {
			SetPH(keyName, HandlePlaceHolder(variable))
		}

		colorCode := systools.CreateColorCode()
		bgCode := systools.CurrentBgColor
		SetPH("RUN.TARGET", target)
		targetFound := false
		for _, script := range runCfg.Task {
			// check if we have found the target
			if strings.EqualFold(target, script.ID) {
				GetLogger().WithFields(logrus.Fields{
					"target": target,
				}).Info("EXECUTE target")
				targetFound = true
				// first get the task related variables
				for keyName, variable := range script.Variables {
					SetPH(keyName, HandlePlaceHolder(variable))
				}

				// convert to stopReason struct
				stopReason := configure.StopReasons(script.Stopreasons)

				// check requirements
				canRun, message := checkRequirements(script.Requires)
				if canRun == false {
					if script.Options.Displaycmd {
						fmt.Println(output.MessageCln(output.ForeYellow, " [require] ", output.ForeBlue, message))
					}
					return ExitByRequirement
				}

				for _, codeLine := range script.Script {

					panelSize := 12
					if script.Options.Panelsize > 0 {
						panelSize = script.Options.Panelsize
					}
					var mainCommand = defaultString(script.Options.Maincmd, DefaultCommandFallBack)
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
							for _, listener := range script.Listener {
								listenReason := configure.StopReasons(listener.Trigger)
								triggerFound, triggerMessage := checkReason(listenReason, logLine)
								if triggerFound {
									SetPH("RUN."+target+".LOG.HIT", logLine)
									if script.Options.Displaycmd {
										fmt.Println(output.MessageCln(output.ForeCyan, "[trigger]\t", output.ForeYellow, triggerMessage, output.Dim, " ", logLine))
									}
									actionDef := configure.Action(listener.Action)
									if actionDef.Target != "" {
										if script.Options.Displaycmd {
											fmt.Println(output.MessageCln(output.ForeCyan, "[trigger]\t ", output.ForeGreen, "target:", output.ForeLightGreen, actionDef.Target))
										}
										GetLogger().WithFields(logrus.Fields{
											"trigger": triggerMessage,
											"target":  actionDef.Target,
										}).Info("trigger called")

										if useWaitGroup {
											go executeTemplate(waitGroup, useWaitGroup, runCfg, actionDef.Target)

										} else {
											executeTemplate(waitGroup, useWaitGroup, runCfg, actionDef.Target)
										}
									} else {
										GetLogger().WithFields(logrus.Fields{
											"trigger": triggerMessage,
											"output":  logLine,
										}).Warn("trigger defined without any target")
									}
								}
							}
						}

						// print the output by configuration
						if script.Options.Hideout == false {
							if script.Options.Format != "" {
								fmt.Printf(script.Options.Format, logLine)
							} else {
								foreColor := defaultString(script.Options.Colorcode, colorCode)
								bgColor := defaultString(script.Options.Bgcolorcode, bgCode)
								labelStr := systools.LabelPrintWithArg(systools.PadStringToR(target+" :", panelSize), foreColor, bgColor, 1)

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
						return ExitByStopReason
					case ExitCmdError:
						if script.Options.IgnoreCmdError {
							if script.Stopreasons.Onerror {
								return ExitByStopReason
							}
							fmt.Println(output.MessageCln(output.ForeYellow, "NOTE!\t", output.BackLightYellow, output.ForeDarkGrey, " a script execution was failing. no stopreason is set so execution will continued "))
							fmt.Println(output.MessageCln("\t", output.BackLightYellow, output.ForeDarkGrey, " if this is expected you can ignore this message.                                 "))
							fmt.Println(output.MessageCln("\t", output.BackLightYellow, output.ForeDarkGrey, " but you should handle error cases                                                "))
							fmt.Println("\ttarget :\t", output.MessageCln(output.ForeYellow, target))
							fmt.Println("\tcommand:\t", output.MessageCln(output.ForeYellow, codeLine))
							return ExitOk
						}
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
						return ExitCmdError

					}
				}
				// executes next targets if there some defined
				GetLogger().WithFields(logrus.Fields{
					"current-target": target,
					"nexts":          script.Next,
				}).Debug("next definition")
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
				return ExitOk
			}

		}

		if !targetFound {
			fmt.Println(output.MessageCln(output.ForeRed, "target not found: ", output.BackRed, output.ForeWhite, target))
			GetLogger().Error("Target can not be found: ", target)
		}
	}
	return ExitNoCode
}

func defaultString(line string, defaultString string) string {
	if line == "" {
		return defaultString
	}
	return line
}

func stringContains(findInHere string, matches []string) bool {
	for _, check := range matches {
		if check != "" && strings.Contains(findInHere, check) {
			return true
		}
	}
	return false
}

func checkReason(stopReason configure.StopReasons, output string) (bool, string) {
	var message = ""
	if stopReason.OnoutcountLess > 0 && stopReason.OnoutcountLess > len(output) {
		message = fmt.Sprint("reason match output len (", len(output), ") is less then ", stopReason.OnoutcountLess)
		return true, message
	}
	if stopReason.OnoutcountMore > 0 && stopReason.OnoutcountMore < len(output) {
		message = fmt.Sprint("reason match output len (", len(output), ") is more then ", stopReason.OnoutcountMore)
		return true, message
	}

	for _, checkText := range stopReason.OnoutContains {
		if checkText != "" && strings.Contains(output, checkText) {
			message = fmt.Sprint("reason match because output contains ", checkText)
			return true, message
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
			GetLogger().Debug("loading json File:", filename)
			if keyname == "" {
				keyname = jsonBaseName
			}
			ImportDataFromJSONFile(keyname, filename)

		}, func(yamlBaseName string) {
			GetLogger().Debug("loading yaml File:", filename)
			if keyname == "" {
				keyname = yamlBaseName
			}
			ImportDataFromYAMLFile(keyname, filename)

		}, func(path string, err error) {
			GetLogger().Errorln("file not exists:", err)
		})
	}
}
