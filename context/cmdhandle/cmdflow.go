package cmdhandle

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/swaros/contxt/context/systools"

	"github.com/swaros/contxt/context/configure"
)

// RunTargets executes multiple targets
func RunTargets(targets string) {
	allTargets := strings.Split(targets, ",")
	template, exists := GetTemplate()

	var runSequencially = false
	if exists {
		runSequencially = template.Config.Sequencially
	}

	if runSequencially == false {
		// run in thread
		var wg sync.WaitGroup
		for _, runTarget := range allTargets {
			wg.Add(1)
			go ExecuteTemplateWorker(&wg, runTarget)
		}
		wg.Wait()

	} else {
		// trun one by one
		for _, runTarget := range allTargets {
			ExecCurrentPathTemplate(runTarget)
		}
	}

}

func executeTemplate(runCfg configure.RunConfig, target string) {
	colorCode := systools.CreateColorCode()
	bgCode := systools.CreateBgColor()
	defaultFormat := systools.GetCodeBg(bgCode) + systools.PrintColored(colorCode, systools.PadStringToR(target+" :", 8)) + systools.GetReset() + "%s\n"
	for _, script := range runCfg.Task {
		// check if we have found the target
		if strings.EqualFold(target, script.ID) {
			// convert to stopReason struct
			stopReason := configure.StopReasons(script.Stopreasons)

			for _, codeLine := range script.Script {
				if script.Options.Displaycmd {
					fmt.Println(systools.Magenta(" RUN "), systools.Teal(target), systools.White(codeLine))
				}

				var mainCommand = defaultString(script.Options.Maincmd, DefaultCommandFallBack)
				ExecuteScriptLine(mainCommand, codeLine, func(logLine string) bool {
					// the watcher
					if script.Listener != nil {
						for _, listener := range script.Listener {
							listenReason := configure.StopReasons(listener.Trigger)
							triggerFound, triggerMessage := checkReason(listenReason, logLine)
							if triggerFound {
								fmt.Println(systools.Magenta("\tlistener hit"), systools.Yellow(triggerMessage), logLine)
								actionDef := configure.Action(listener.Action)
								if actionDef.Target != "" {
									go executeTemplate(runCfg, actionDef.Target)
								}
							}
						}
					}

					// print the output by configuration
					if script.Options.Hideout == false {

						//fmt.Printf(defaultString(script.Options.Format, defaultFormat), systools.White(logLine))
						fmt.Printf(defaultString(script.Options.Format, defaultFormat), systools.PrintColored(colorCode, logLine))

					}
					// do we found a defined reason to stop execution
					stopReasonFound, message := checkReason(stopReason, logLine)
					if stopReasonFound {
						if script.Options.Displaycmd {
							fmt.Println(systools.Teal(" HIT "), systools.Info(message))
						}
						return false
					}
					return true
				}, func(process *os.Process) {
					if script.Options.Displaycmd {
						fmt.Println(systools.Magenta(" PID "), systools.Teal(process.Pid))
					}
				})
			}
		}

	}
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
