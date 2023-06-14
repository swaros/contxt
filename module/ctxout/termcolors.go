// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

 package ctxout

import "strings"

const (
	// terminal reser code
	ResetCode = "\033[0m"
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
	BoldTag = "<b>"
	// Dim dim
	Dim = "<dim>"
	// Underlined tag
	Underlined = "<u>"
	// Reverse tag
	Reverse = "<r>"
	// Hidden tag
	Hidden = "<hide>"
	// ResetBold tag
	ResetBold = "</b>"
	// ResetDim tag
	ResetDim = "</dim>"
	// ResetUnderline tag
	ResetUnderline = "</u>"
	//ResetReverse tag
	ResetReverse = "</r>"
	//ResetHidden tag
	ResetHidden = "</hide>"

	// BackBlack black Background color
	BackBlack = "<b:black>"
	// BackRed red Background color
	BackRed = "<b:red>"
	// BackGreen red Background color
	BackGreen = "<b:green>"
	// BackYellow red Background color
	BackYellow = "<b:yellow>"
	// BackBlue red Background color
	BackBlue = "<b:blue>"
	// BackMagenta red Background color
	BackMagenta = "<b:magenta>"
	// BackCyan red Background color
	BackCyan = "<b:cyan>"
	// BackLightGrey red Background color
	BackLightGrey = "<b:light-grey>"
	// BackDarkGrey red Background color
	BackDarkGrey = "<b:dark-grey>"
	// BackLightRed red Background color
	BackLightRed = "<b:light-red>"
	// BackLightGreen red Background color
	BackLightGreen = "<b:light-green>"
	// BackLightYellow red Background color
	BackLightYellow = "<b:light-yellow>"
	// BackLightBlue red Background color
	BackLightBlue = "<b:light-blue>"
	// BackLightCyan red Background color
	BackLightCyan = "<b:light-cyan>"
	// BackLightMagenta red Background color
	BackLightMagenta = "<b:light-magenta>"
	// BackWhite red Background color
	BackWhite = "<b:white>"
)

type BasicColors struct {
	parser Markup
	// Foreground colors
	Red, Green, Yellow, Blue, Magenta, Cyan, LightGrey, DarkGrey, LightRed, LightGreen, LightYellow, LightBlue, LightCyan, LightMagenta, White string
	// Background colors
	BackBlack, BackRed, BackGreen, BackYellow, BackBlue, BackMagenta, BackCyan, BackLightGrey, BackDarkGrey, BackLightRed, BackLightGreen, BackLightYellow, BackLightBlue, BackLightCyan, BackLightMagenta, BackWhite string
	// Bold, Dim, Underlined, Reverse, Hidden string
}

func NewBasicColors() *BasicColors {
	bc := &BasicColors{}
	bc.InitColors()
	bc.parser = *NewMarkup()
	return bc
}

func (bc *BasicColors) InitColors() {
	bc.Red = ForeRed
	bc.Green = ForeGreen
	bc.Yellow = ForeYellow
	bc.Blue = ForeBlue
	bc.Magenta = ForeMagenta
	bc.Cyan = ForeCyan
	bc.LightGrey = ForeLightGrey
	bc.DarkGrey = ForeDarkGrey
	bc.LightRed = ForeLightRed
	bc.LightGreen = ForeLightGreen
	bc.LightYellow = ForeLightYellow
	bc.LightBlue = ForeLightBlue
	bc.LightCyan = ForeLightCyan
	bc.LightMagenta = ForeLightMagenta
	bc.White = ForeWhite
	bc.BackBlack = BackBlack
	bc.BackRed = BackRed
	bc.BackGreen = BackGreen
	bc.BackYellow = BackYellow
	bc.BackBlue = BackBlue
	bc.BackMagenta = BackMagenta
	bc.BackCyan = BackCyan
	bc.BackLightGrey = BackLightGrey
	bc.BackDarkGrey = BackDarkGrey
	bc.BackLightRed = BackLightRed
	bc.BackLightGreen = BackLightGreen
	bc.BackLightYellow = BackLightYellow
	bc.BackLightBlue = BackLightBlue
	bc.BackLightCyan = BackLightCyan
	bc.BackLightMagenta = BackLightMagenta
	bc.BackWhite = BackWhite
}

func (bc *BasicColors) IsBasicColor(text string) bool {
	return HaveBasicColors(text)
}

func (bc *BasicColors) ParseText(text string, handleColorCode func(m Parsed) string) string {
	markups := bc.parser.Parse(text)
	newText := ""
	for _, m := range markups {

		newText += handleColorCode(m)

	}
	return newText
}

func HaveBasicColors(input string) bool {
	return strings.Contains(input, ForeRed) ||
		strings.Contains(input, ForeGreen) ||
		strings.Contains(input, ForeYellow) ||
		strings.Contains(input, ForeBlue) ||
		strings.Contains(input, ForeMagenta) ||
		strings.Contains(input, ForeCyan) ||
		strings.Contains(input, ForeLightGrey) ||
		strings.Contains(input, ForeDarkGrey) ||
		strings.Contains(input, ForeLightRed) ||
		strings.Contains(input, ForeLightGreen) ||
		strings.Contains(input, ForeLightYellow) ||
		strings.Contains(input, ForeLightBlue) ||
		strings.Contains(input, ForeLightCyan) ||
		strings.Contains(input, ForeLightMagenta) ||
		strings.Contains(input, ForeWhite) ||
		strings.Contains(input, BackBlack) ||
		strings.Contains(input, BackRed) ||
		strings.Contains(input, BackGreen) ||
		strings.Contains(input, BackYellow) ||
		strings.Contains(input, BackBlue) ||
		strings.Contains(input, BackMagenta) ||
		strings.Contains(input, BackCyan) ||
		strings.Contains(input, BackLightGrey) ||
		strings.Contains(input, BackDarkGrey) ||
		strings.Contains(input, BackLightRed) ||
		strings.Contains(input, BackLightGreen) ||
		strings.Contains(input, BackLightYellow) ||
		strings.Contains(input, BackLightBlue) ||
		strings.Contains(input, BackLightCyan) ||
		strings.Contains(input, BackLightMagenta) ||
		strings.Contains(input, BackWhite)
}
