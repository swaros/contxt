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

package systools

import (
	"fmt"
	"os"

	"github.com/swaros/manout"
	"golang.org/x/term"
)

var lastUsedForeColorIndex = 1
var lastUsedBgColorIndex = 2

// CurrentColor current used foreground color
var CurrentColor = "32"

// CurrentBgColor current used background color
var CurrentBgColor = "42"

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

// CreateBgColor defines a random color and returns a id
func CreateBgColor() string {
	CreateBgColorCode()
	return colorBgCodes[lastUsedBgColorIndex]
}

// CreateColorCode returns the colorcode by a random number
func CreateColorCode() (string, string) {
	lastUsedForeColorIndex++
	if lastUsedForeColorIndex >= len(colorCodes) {
		lastUsedForeColorIndex = 0
		CreateBgColor()
	}
	CurrentColor = colorCodes[lastUsedForeColorIndex]
	// hardcoded way to check if a backgroundcolor
	// is usable with forground color
	// might be replaced in future with something else
	for !colorCombinationisFine() {
		CreateColorCode()
	}
	return CurrentColor, CurrentBgColor
}

// CreateBgColorCode returns the colorcode by a random number
func CreateBgColorCode() string {
	lastUsedBgColorIndex++
	if lastUsedBgColorIndex >= len(colorBgCodes) {
		lastUsedBgColorIndex = 0
	}

	CurrentBgColor = colorBgCodes[lastUsedBgColorIndex]
	return CurrentBgColor
}

func colorCombinationisFine() bool {
	return true
	/*
		switch CurrentBgColor {

		case "47":
			{
				switch CurrentColor {
				case "97", "37", "36", "96", "93":
					return false
				}
			}
		case "100":
			{
				switch CurrentColor {
				case "32", "36", "95", "96", "37", "93", "97", "90", "94":
					return false
				}
			}
		case "101":
			{
				switch CurrentColor {
				case "32", "33", "37", "94", "95", "96", "31", "34", "90", "91":
					return false
				}
			}
		case "102":
			{
				switch CurrentColor {
				case "97", "32", "92":
					return false
				}
			}
		case "103":
			{
				switch CurrentColor {
				case "35", "93", "37", "97", "33":
					return false
				}
			}
		case "104":
			{
				switch CurrentColor {
				case "37", "92", "93", "96", "97", "31", "90", "91", "34", "35", "94", "95":
					return false
				}
			}
		case "105":
			{
				switch CurrentColor {
				case "37", "92", "93", "96", "97", "31", "90", "91", "34", "35", "95", "94":
					return false
				}
			}
		case "106":
			{
				switch CurrentColor {
				case "97", "36", "96":
					return false
				}
			}
		case "40":
			{
				switch CurrentColor {
				case "31", "34", "90", "91", "35":
					return false
				}
			}
		case "41":
			{
				switch CurrentColor {
				case "31", "32", "34", "36", "91", "95", "93", "35", "37", "90":
					return false
				}
			}
		case "42":
			{
				switch CurrentColor {
				case "31", "34", "37", "90", "92", "93", "96", "97", "94":
					return false
				}
			}
		case "43":
			{
				switch CurrentColor {
				case "92", "93", "97", "37", "32", "90":
					return false
				}
			}
		case "44":
			{
				switch CurrentColor {
				case "32", "33", "36", "37", "94", "95", "93", "96", "97", "35", "31":
					return false
				}
			}
		case "45":
			{
				switch CurrentColor {
				case "35", "95":
					return false
				}
			}
		case "46":
			{
				switch CurrentColor {
				case "37", "92", "96", "90", "31":
					return false
				}
			}

		}
		return true*/
}

func colorFormatNumber(code string) string {
	if !manout.ColorEnabled {
		return "%s"
	}
	return "\033[1;" + code + "m%s"
}

// PrintColored formats string colored by the color id
func PrintColored(code string, outputs string) string {
	if !manout.ColorEnabled {
		return outputs
	}
	return fmt.Sprintf(colorFormatNumber(code), outputs)
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
	outstr := fmt.Sprintf("\033[%d;%s;%sm %s \033[0m", attribute, CurrentBgColor, CurrentColor, message)
	return outstr
}

// LabelPrintWithArg prints message by using current fore and background
func LabelPrintWithArg(message string, fg string, bg string, attribute int) string {
	if !manout.ColorEnabled {
		return message
	}
	//sign := "\u2807"
	//sign := "█▇▆▅▄▃▂▁"
	sign := "▓▒░"
	preSign := "░▒▓"
	colorblock := fmt.Sprintf("\033[%d;%s;%sm%s\033[0m ", attribute, fg, bg, sign)
	colorblockPre := fmt.Sprintf("\033[%d;%s;%sm%s\033[0m", attribute, fg, bg, preSign)
	coloredMsg := fmt.Sprintf("\033[%d;%sm%s", attribute, fg, message)
	return colorblockPre + coloredMsg + colorblock // + "⇨"
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

func TestColorCombinations() {
	width := 0

	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w

	}
	outLen := 0
	startColComb := ""
	for i := 0; i < 5000; i++ {
		CreateColorCode()
		origStr := "b" + CurrentBgColor + "f" + CurrentColor + ""
		if i == 0 {
			startColComb = origStr
		} else {
			if startColComb == origStr {
				return
			}
		}
		str := LabelPrintWithArg("TEST", CurrentColor, CurrentBgColor, 1)
		if width == 0 {
			fmt.Println(origStr, str)
		} else {
			outLen += (len(origStr) + 6)
			if outLen >= width {
				fmt.Println()
				outLen = (len(origStr) + 6)
			}
			fmt.Print(origStr, str)
		}

	}
	ResetColors(true)
}
