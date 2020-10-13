package systools

import (
	"fmt"
	"math/rand"
)

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

// CreateColor defines a random color and returns a id
func CreateColor() string {
	rand.Seed(2)
	colorNumber := rand.Intn(len(colorCodes))
	return colorFormatNumber(colorCodes[colorNumber])
}

// CreateBgColor defines a random color and returns a id
func CreateBgColor() string {
	rand.Seed(4)
	colorNumber := rand.Intn(len(colorBgCodes))
	return colorBgCodes[colorNumber]
}

// CreateColorCode returns the colorcode by a random number
func CreateColorCode() string {
	colorNumber := rand.Intn(len(colorCodes))
	return colorCodes[colorNumber]
}

func colorFormatNumber(code string) string {
	return "\033[1;" + code + "m%s"
}

func colorFormatWithBg(code string, bg string) string {
	return "\033[1;" + code + "m\033[" + bg + "m%s"
}

// PrintColored formats string colored by the color id
func PrintColored(code string, output string) string {
	return fmt.Sprintf(colorFormatNumber(code), output)
}

// PrintColoredBg formats string colored by the color id including background
func PrintColoredBg(code string, bgCode string, output string) string {
	return fmt.Sprintf(colorFormatWithBg(code, bgCode), output)
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

// GetDefaultBg Get the Default Background
func GetDefaultBg() string {
	return "\033[49m"
}

// GetReset gets the reset code
func GetReset() string {
	return "\033[0m"
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
