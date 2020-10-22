package systools

import (
	"fmt"

	"github.com/swaros/contxt/context/output"
)

var lastUsedColor = 2
var lastUsedBgColor = 4

// CurrentColor current used foreground color
var CurrentColor = "36"

// CurrentBgColor current used background color
var CurrentBgColor = "40"

var colorCodes = map[int]string{
	0:  "31",
	1:  "32",
	2:  "33",
	3:  "34",
	4:  "35",
	5:  "36",
	6:  "37",
	7:  "90",
	8:  "91",
	9:  "92",
	10: "93",
	11: "94",
	12: "95",
	13: "96",
	14: "97",
}

var colorBgCodes = map[int]string{
	0:  "40",
	1:  "41",
	2:  "42",
	3:  "43",
	4:  "44",
	5:  "45",
	6:  "46",
	7:  "47",
	8:  "100",
	9:  "101",
	10: "102",
	11: "103",
	12: "104",
	13: "105",
	14: "106",
}

// LabelColor contains fore and background color
type LabelColor struct {
	Color   string
	BgColor string
}

// CreateColor defines a random color and returns a id
func CreateColor() string {
	CreateColorCode()
	return colorFormatNumber(colorCodes[lastUsedColor])
}

// CreateBgColor defines a random color and returns a id
func CreateBgColor() string {
	CreateBgColorCode()
	return colorBgCodes[lastUsedBgColor]
}

// CreateColorCode returns the colorcode by a random number
func CreateColorCode() string {
	lastUsedColor++
	if lastUsedColor >= len(colorCodes) {
		lastUsedColor = 0
		CreateBgColor()
	}
	CurrentColor = colorCodes[lastUsedColor]
	return CurrentColor
}

// CreateBgColorCode returns the colorcode by a random number
func CreateBgColorCode() string {
	lastUsedBgColor++
	if lastUsedBgColor >= len(colorBgCodes) {
		lastUsedBgColor = 0
	}

	CurrentBgColor = colorBgCodes[lastUsedBgColor]
	return CurrentBgColor
}

func colorFormatNumber(code string) string {
	if !output.ColorEnabled {
		return "%s"
	}
	return "\033[1;" + code + "m%s"
}

func colorFormatWithBg(code string, bg string) string {
	if !output.ColorEnabled {
		return "%s"
	}
	return "\033[1;" + code + "m\033[" + bg + "m%s"
}

// PrintColored formats string colored by the color id
func PrintColored(code string, outputs string) string {
	if !output.ColorEnabled {
		return outputs
	}
	return fmt.Sprintf(colorFormatNumber(code), outputs)
}

// PrintColoredBg formats string colored by the color id including background
func PrintColoredBg(code string, bgCode string, outputs string) string {
	if !output.ColorEnabled {
		return outputs
	}
	return fmt.Sprintf(colorFormatWithBg(code, bgCode), outputs)
}

// GetWhiteBg Get the White Background
func GetWhiteBg() string {
	return "\033[107m"
}

// GetGreyBg Get the DarkGrey Background
func GetGreyBg() string {
	return "\033[100m"
}

// GetCodeBg Get the Background from color Code
func GetCodeBg(code string) string {
	return "\033[" + code + "m"
}

func getCodeFg(code string) string {

	return "\033[1;" + code + "m"

}

// GetDefaultBg Get the Default Background
func GetDefaultBg() string {
	return "\033[49m"
}

// GetReset gets the reset code
func GetReset() string {
	return "\033[0m"
}

// ResetColors resets terminal colors
// if print false you get the ansi code only
func ResetColors(print bool) string {
	if print {
		fmt.Print(GetReset(), GetDefaultBg())
	}
	return GetReset() + GetDefaultBg()
}

// LabelPrint prints message by using current fore and background
func LabelPrint(message string, attribute int) string {

	//outstr := fmt.Sprintf("\033[%s;%sm%s\033[0m", currentColor, currentBgColor, message)
	outstr := fmt.Sprintf("\033[%d;%s;%sm %s \033[0m", attribute, CurrentBgColor, CurrentColor, message)
	return outstr
}

// LabelPrintWithArg prints message by using current fore and background
func LabelPrintWithArg(message string, fg string, bg string, attribute int) string {
	if !output.ColorEnabled {
		return message
	}
	//outstr := fmt.Sprintf("\033[%s;%sm%s\033[0m", currentColor, currentBgColor, message)
	outstr := fmt.Sprintf("\033[%d;%s;%sm %s \033[0m", attribute, fg, bg, message)
	return outstr
}

// PadString Returns max len string filled with spaces
func PadString(line string, max int) string {
	if len(line) > max {
		runes := []rune(line)
		safeSubstring := string(runes[0:max])
		return safeSubstring
	}
	diff := max - len(line)
	for i := 0; i < diff; i++ {
		line = line + " "
	}
	return line
}

// PadStringToR Returns max len string filled with spaces right placed
func PadStringToR(line string, max int) string {
	if len(line) > max {
		runes := []rune(line)
		safeSubstring := string(runes[0:max])
		return safeSubstring
	}
	diff := max - len(line)
	for i := 0; i < diff; i++ {
		line = " " + line
	}
	return line
}

/*
// TestPrintColoredChanges is for testing the color formats
func TestPrintColoredChanges() {
	for i := 0; i < 500; i++ {
		CreateColorCode()
		//outpt := PrintColored(colorCode, "test ["+colorCode+"] last ("+lastCode+")")
		labelOut := LabelPrint("\t label print ", 2)
		fmt.Println(labelOut, "\t", currentColor, currentBgColor)

	}
	ResetColors(true)
}
*/
