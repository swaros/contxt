package cmdhandle

import (
	"fmt"
	"os"
	"strings"

	"github.com/swaros/contxt/context/systools"

	"github.com/swaros/contxt/context/configure"
)

func executeTemplate(runCfg configure.RunConfig, target string) {
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

					// print the output by configuration
					if script.Options.Hideout == false {
						fmt.Printf(defaultString(
							script.Options.Format,
							systools.Yellow(target)+"\t"+systools.Teal("|")+"%s\n"), systools.White(logLine))
					}
					// do we found a defined reason to stop execution
					stopReasonFound := checkStopReason(stopReason, logLine)
					if stopReasonFound {
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

func checkStopReason(stopReason configure.StopReasons, output string) bool {
	var stopReasonBool = false
	if stopReason.OnoutcountLess > 0 && stopReason.OnoutcountLess > len(output) {
		fmt.Println("stopped because output len (", len(output), ") is less then ", stopReason.OnoutcountLess)
		stopReasonBool = true
	}
	if stopReason.OnoutcountMore > 0 && stopReason.OnoutcountMore < len(output) {
		fmt.Println("stopped because output len (", len(output), ") is more then ", stopReason.OnoutcountMore)
		stopReasonBool = true
	}

	for _, checkText := range stopReason.OnoutContains {
		if checkText != "" && strings.Contains(output, checkText) {
			fmt.Println("stoppend because output contains ", checkText)
			stopReasonBool = true
		}
	}

	return stopReasonBool
}
