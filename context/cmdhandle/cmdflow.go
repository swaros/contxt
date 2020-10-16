package cmdhandle

import (
	"fmt"
	"os"
	"strings"
	"sync"

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
)

// RunTargets executes multiple targets
func RunTargets(targets string) {
	allTargets := strings.Split(targets, ",")
	template, templatePath, exists := GetTemplate()

	var runSequencially = false
	if exists {
		runSequencially = template.Config.Sequencially
	}

	if runSequencially == false {
		// run in thread
		var wg sync.WaitGroup
		for _, runTarget := range allTargets {
			wg.Add(1)
			go ExecuteTemplateWorker(&wg, runTarget, templatePath)
		}
		wg.Wait()

	} else {
		// trun one by one
		for _, runTarget := range allTargets {
			//ExecCurrentPathTemplate(runTarget)
			ExecPathFile(templatePath, runTarget)
		}
	}

}

func executeTemplate(runCfg configure.RunConfig, target string) int {
	if len(runCfg.Task) > 0 {
		colorCode := systools.CreateColorCode()
		bgCode := systools.CurrentBgColor
		SetPH("RUN.TARGET", target)
		for _, script := range runCfg.Task {
			// check if we have found the target
			if strings.EqualFold(target, script.ID) {
				// convert to stopReason struct
				stopReason := configure.StopReasons(script.Stopreasons)

				for _, codeLine := range script.Script {
					if script.Options.Displaycmd {
						fmt.Println(systools.Magenta(" RUN "), systools.Teal(target), systools.White(codeLine))
					}
					panelSize := 12
					if script.Options.Panelsize > 0 {
						panelSize = script.Options.Panelsize
					}
					var mainCommand = defaultString(script.Options.Maincmd, DefaultCommandFallBack)
					replacedLine := HandlePlaceHolder(codeLine)

					SetPH("RUN.SCRIPT_LINE", codeLine)
					execCode, execErr := ExecuteScriptLine(mainCommand, replacedLine, func(logLine string) bool {

						SetPH("RUN."+target+".LOG.LAST", logLine)
						// the watcher
						if script.Listener != nil {
							for _, listener := range script.Listener {
								listenReason := configure.StopReasons(listener.Trigger)
								triggerFound, triggerMessage := checkReason(listenReason, logLine)
								if triggerFound {
									SetPH("RUN."+target+".LOG.HIT", logLine)
									if script.Options.Displaycmd {
										fmt.Println(systools.Magenta("\tlistener hit"), systools.Yellow(triggerMessage), logLine)
									}
									actionDef := configure.Action(listener.Action)
									if actionDef.Target != "" {
										go executeTemplate(runCfg, actionDef.Target)
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
								fmt.Println(systools.Magenta(" STOP-HIT "), systools.Info(message))
							}
							return false
						}
						return true
					}, func(process *os.Process) {
						pidStr := fmt.Sprintf("%d", process.Pid)
						SetPH("RUN.PID", pidStr)
						SetPH("RUN."+target+".PID", pidStr)
						if script.Options.Displaycmd {
							fmt.Println(systools.Magenta(" PID "), systools.Teal(process.Pid))
						}
					})
					if execErr != nil {
						fmt.Println(systools.Warn(execErr))
					}
					if execCode == ExitByStopReason {
						return ExitByStopReason
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
