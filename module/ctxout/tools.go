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

import (
	"regexp"
	"strings"

	"github.com/muesli/reflow/ansi"
	"github.com/rivo/uniseg"
)

// UniseqLen returns the length of a string, but only counts printable characters
// it ignores ANSI escape codes
// and also ignores the length of the ANSI escape codes
func UniseqLen(s string) int {
	return uniseg.StringWidth(s)
}

// returns the length of a string, but only counts printable characters
// it ignores ANSI escape codes
// makes use of the ansi package from muesli
func VisibleLen(s string) int {
	return ansi.PrintableRuneWidth(s)
}

func StringPure(s string) string {
	removeChars := []string{
		"\n",
		"\t",
		"\r",
		"\x08",
	}
	for _, c := range removeChars {
		s = strings.ReplaceAll(s, c, "")
	}
	return StringCleanEscapeCodes(s)
}

func StringCleanEscapeCodes(s string) string {
	match := "[^\x08]\x08"
	match += "|\\x1b\\[[0-9;]*[a-zA-Z]"
	match += "|\\x1b\\[[0-9;]*m"

	re := regexp.MustCompile(match)
	return re.ReplaceAllString(s, "")
}

// splits a string into two parts. one part is the cut string, the other part is the rest
// the cut string is the first part of the string with the given size
// the rest is the remaining part of the string
func StringCut(s string, size int) (cutStr string, rest string) {
	if size <= 0 {
		return "", s
	}
	if size >= VisibleLen(s) {
		return s, ""
	}
	left := s[:size]
	right := s[size:]
	return left, right
}

// splits a string into two parts. one part is the cut string, the other part is the rest
// the cut string is the last part of the string with the given size starting from the right side
// the rest is the remaining part of the string
// so if you have a string "1234567890" and you call StringCutFromRight("1234567890", 5)
// then the result will be "67890" and "12345"
func StringCutFromRight(s string, size int) (cutStr string, rest string) {
	if size <= 0 {
		return "", s
	}
	if size >= VisibleLen(s) {
		return s, ""
	}
	left := s[:len(s)-size]
	right := s[len(s)-size:]
	return right, left
}
