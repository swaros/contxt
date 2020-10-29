package cmdhandle

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
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
)

// RunTargets executes multiple targets
func RunTargets(targets string) {
	allTargets := strings.Split(targets, ",")
	template, templatePath, exists := GetTemplate()

	var runSequencially = false
	if exists {
		runSequencially = template.Config.Sequencially
		output.ColorEnabled = !template.Config.Coloroff
	}

	if template.Config.LogLevel != "" {
		setLogLevelByString(template.Config.LogLevel)
	}

	var wg sync.WaitGroup
	if runSequencially == false {
		// run in thread
		fmt.Println(output.MessageCln(output.ForeCyan, "thread runmode"))
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
	fmt.Println("done target run")
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

func executeTemplate(waitGroup *sync.WaitGroup, useWaitGroup bool, runCfg configure.RunConfig, target string) int {
	if useWaitGroup {
		waitGroup.Add(1)
		defer waitGroup.Done()
	}

	if len(runCfg.Task) > 0 {

		// main variables
		for keyName, variable := range runCfg.Config.Variables {
			SetPH(keyName, HandlePlaceHolder(variable))
		}

		colorCode := systools.CreateColorCode()
		bgCode := systools.CurrentBgColor
		SetPH("RUN.TARGET", target)
		for _, script := range runCfg.Task {
			// check if we have found the target
			if strings.EqualFold(target, script.ID) {

				// first get the task related variables
				for keyName, variable := range script.Variables {
					SetPH(keyName, HandlePlaceHolder(variable))
				}

				// convert to stopReason struct
				stopReason := configure.StopReasons(script.Stopreasons)

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
					execCode, execErr := ExecuteScriptLine(mainCommand, script.Options.Mainparams, replacedLine, func(logLine string) bool {

						SetPH("RUN."+target+".LOG.LAST", logLine)
						// the watcher
						if script.Listener != nil {
							for _, listener := range script.Listener {
								listenReason := configure.StopReasons(listener.Trigger)
								triggerFound, triggerMessage := checkReason(listenReason, logLine)
								if triggerFound {
									SetPH("RUN."+target+".LOG.HIT", logLine)
									if script.Options.Displaycmd {
										fmt.Println(output.MessageCln(output.ForeMagenta, "\tlistener hit", output.ForeYellow, triggerMessage, output.Reverse, logLine))
									}
									actionDef := configure.Action(listener.Action)
									if actionDef.Target != "" {
										if useWaitGroup {
											go executeTemplate(waitGroup, useWaitGroup, runCfg, actionDef.Target)

										} else {
											executeTemplate(waitGroup, useWaitGroup, runCfg, actionDef.Target)
										}
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
				}
				return ExitOk
			}

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
		message = fmt.Sprint("\treason match output len (", len(output), ") is less then ", stopReason.OnoutcountLess)
		return true, message
	}
	if stopReason.OnoutcountMore > 0 && stopReason.OnoutcountMore < len(output) {
		message = fmt.Sprint("\treason match output len (", len(output), ") is more then ", stopReason.OnoutcountMore)
		return true, message
	}

	for _, checkText := range stopReason.OnoutContains {
		if checkText != "" && strings.Contains(output, checkText) {
			message = fmt.Sprint("s\treason match because output contains ", checkText)
			return true, message
		}
	}

	return false, message
}
