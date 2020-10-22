package output

import (
	"fmt"
	"strings"
)

// ColorEnabled enables or disables the color usage
var ColorEnabled = true

const (
	resetCode = "\033[0m"
	// CleanTag is the tag to reset to default
	CleanTag = "</>"
	// ForeRed red foreground color
	ForeRed = "<f:red>"
	// ForeGreen red foreground color
	ForeGreen = "<f:green>"
	// ForeYellow red foreground color
	ForeYellow = "<f:yellow>"
	// ForeBlue red foreground color
	ForeBlue = "<f:blue>"
	// ForeMagenta red foreground color
	ForeMagenta = "<f:magenta>"
	// ForeCyan red foreground color
	ForeCyan = "<f:cyan>"
	// ForeLightGrey red foreground color
	ForeLightGrey = "<f:light-grey>"
	// ForeDarkGrey red foreground color
	ForeDarkGrey = "<f:dark-grey>"
	// ForeLightRed red foreground color
	ForeLightRed = "<f:light-red>"
	// ForeLightGreen red foreground color
	ForeLightGreen = "<f:light-green>"
	// ForeLightYellow red foreground color
	ForeLightYellow = "<f:light-yellow>"
	// ForeLightBlue red foreground color
	ForeLightBlue = "<f:light-blue>"
	// ForeLightCyan red foreground color
	ForeLightCyan = "<f:light-cyan>"
	// ForeLightMagenta red foreground color
	ForeLightMagenta = "<f:light-magenta>"
	// ForeWhite red foreground color
	ForeWhite = "<f:white>"

	// BoldTag writes bolder text
	BoldTag = "<bold>"
)

var tagMap = map[string]string{
	ForeRed:          "31",
	ForeGreen:        "32",
	ForeYellow:       "33",
	ForeBlue:         "34",
	ForeMagenta:      "35",
	ForeCyan:         "36",
	ForeLightGrey:    "37",
	ForeDarkGrey:     "90",
	ForeLightRed:     "91",
	ForeLightGreen:   "92",
	ForeLightYellow:  "93",
	ForeLightBlue:    "94",
	ForeLightMagenta: "95",
	ForeLightCyan:    "96",
	ForeWhite:        "97",
}

// MessageCln converts arguemnst to a fomated string and adding cleanup and newline code
func MessageCln(a ...interface{}) string {
	result := Message(a...)
	needToDo := strings.Contains(result, "\033[")
	if needToDo {
		result = Message(result, CleanTag)
	}
	return result
}

// Message get the message an handle the formating of them
func Message(a ...interface{}) string {
	stringResult := fmt.Sprint(a...)
	needToDo := strings.Contains(stringResult, "<")
	if needToDo {
		replaceed := buildColored(stringResult)
		return replaceed
	}
	return stringResult
}

func buildColored(origin string) string {

	for key, code := range tagMap {
		colCde := "\033[" + code + "m"
		if !ColorEnabled {
			colCde = ""
		}
		if strings.Contains(origin, key) {
			origin = strings.ReplaceAll(origin, key, colCde)
		}

		if strings.Contains(origin, CleanTag) {
			if !ColorEnabled {
				origin = strings.ReplaceAll(origin, CleanTag, "")
			} else {
				origin = strings.ReplaceAll(origin, CleanTag, resetCode)
			}
		}

	}

	return origin
}
