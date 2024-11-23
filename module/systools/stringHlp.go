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

// a collection of string helper functions

package systools

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// StringSubLeft returns the left part of a string
func StringSubLeft(path string, max int) string {
	if len := StrLen(path); len <= max {
		return path
	} else {
		runes := []rune(path)
		return string(runes[0:max])
	}
}

// StringSubRight returns the right part of a string
func StringSubRight(inStr string, max int) string {
	if len := StrLen(inStr); len <= max {
		return inStr
	} else {
		runes := []rune(inStr)
		left := len - max
		return string(runes[left:])
	}
}

func StrLen(str string) int {
	return len([]rune(str))
}

// CheckForCleanString checks if a string contains only
// accepted chars for a string. these are A-Z, a-z, 0-9, _ and -
// so "hello_my-friend" is a clean string
// "hello_my-friend!" is not a clean string
// there are chars they will just be replaced by a clean char
// and in this case no error will be returned
// so "hello.my:friend" will be "hello-my--friend"
// and no error will be returned
// but if after that the string contains any other char (what will be checked by a regex)
// an error will be returned
// this is usefull for keys in json, yaml or other config files
func CheckForCleanString(s string) (cleanString string, err error) {
	// replace some expected chars comming from version (dots) nd paths (/\:)
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ":", "--") // windows paths
	if isMatch := regexp.MustCompile(`^[A-Za-z0-9_-]*$`).MatchString(s); isMatch {
		return s, nil
	}
	return "", errors.New("string contains not accepted chars ")
}

// returns the string where any non printable chars are removed
// so "hello\x00my\x01friend" will be "hellomyfriend"
func PrintableChars(str string) string {
	var result []rune
	for _, r := range str {
		if r > 31 && r < 127 {
			result = append(result, r)
		}
	}
	return string(result)
}

// returns the string where any spaces, with an length of more than 1
// are replaced by one space
// so "hello   my friend" will be "hello my friend"
func TrimAllSpaces(s string) string {
	sp := strings.Split(s, " ")
	newSl := []string{}
	for _, v := range sp {
		if v != "" {
			newSl = append(newSl, v)
		}
	}
	return strings.Join(newSl, " ")
}

// this function removes any escape sequences from a string
func NoEscapeSequences(str string) string {
	// we need to remove any escape sequences from the string
	// for output without colors and countingchars they are visible
	// and will break the output

	var result []rune
	var escape bool
	for _, r := range str {
		if r == 27 {
			escape = true
		}
		if !escape {
			result = append(result, r)
		}
		if r == 109 {
			escape = false
		}
	}
	return string(result)
}

// this function creates a short label from a string
// by looking for the first char that is a letter until
// a non letter char is found.
// so hello-my-friend will be hmf
// or hello_my_friend will be hmf
// or hello my friend will be hmf
// also for escape sequences
// so \033[1;32mC.H.E.C.K\033[0m will be CHECK
func ShortLabel(label string, max int) string {
	label = FindStartChars(NoEscapeSequences(label))
	if len := StrLen(label); len <= max {
		return label
	} else {
		runes := []rune(label)
		return string(runes[0:max])
	}
}

// FindStartChars returns the first chars of a string
// that are a letter or a number
// so "hello my friend" will be "hmf"
func FindStartChars(str string) string {
	var result []rune
	hit := false
	for _, r := range str {
		if r > 64 && r < 91 {
			if !hit {
				result = append(result, r)
				hit = true
			}
		} else if r > 96 && r < 123 {
			if !hit {
				result = append(result, r)
				hit = true
			}
		} else if r > 47 && r < 58 {
			if !hit {
				result = append(result, r)
				hit = true
			}
		} else {
			hit = false
		}
	}
	return string(result)
}

// StringSplitArgs splits a string into a command and arguments
// the first argument is the command, the rest are arguments
// the prefix is used to create the map keys for the arguments
// so the first argument will be prefix0, the second prefix1 etc.
func StringSplitArgs(argLine string, prefix string) (string, map[string]string) {
	var args map[string]string = make(map[string]string)
	argArr := strings.Split(argLine, " ")
	for index, v := range argArr {
		if v == "" {
			continue
		}
		args[fmt.Sprintf("%s%v", prefix, index)] = v
	}
	return argArr[0], args
}

// SplitQuoted splits a string by a separator and respects quoted strings
// so 'hello my friend' will be ['hello my friend']
// and hello my friend will be ['hello', 'my', 'friend']
// this func is only using single quotes
func SplitQuoted(oristr string, sep string) []string {
	var result []string
	var placeHolder map[string]string = make(map[string]string)

	found := false
	re := regexp.MustCompile(`'[^']+'`)
	newStrs := re.FindAllString(oristr, -1)
	i := 0
	for _, s := range newStrs {
		pl := "[$" + strconv.Itoa(i) + "]"
		placeHolder[pl] = strings.ReplaceAll(s, `'`, "")
		oristr = strings.Replace(oristr, s, pl, 1)
		found = true
		i++
	}
	result = strings.Split(oristr, sep)
	if !found {
		return result
	}

	for index, val := range result {
		if orgStr, fnd := placeHolder[val]; fnd {
			result[index] = orgStr
		}
	}

	return result
}

// AnyToStrNoTabs just converts any tab (\t) to a space
// so "hello\tworld" will be "hello world"
func AnyToStrNoTabs(any interface{}) string {
	return strings.ReplaceAll(fmt.Sprintf("%v", any), "\t", " ")
}

// FillString fills a string with spaces to a given length
// but only if the string is shorter than the length
// so "hello" with a length of 10 will be "hello     "
// and "hello" with a length of 3 will be "hello"
func FillString(str string, length int) string {
	if len(str) > length {
		return str
	}
	return str + strings.Repeat(" ", length-len(str))
}

// StrContains checks if a string contains a substring
// by not using the strings.Contains function
// because the use template strings and the function
// would not work as expected
func StrContains(str string, substr string) bool {
	var byteArr []byte
	for _, v := range str {
		byteArr = append(byteArr, byte(v))
	}
	var byteArrSub []byte
	for _, v := range substr {
		byteArrSub = append(byteArrSub, byte(v))
	}
	if len(byteArrSub) == 0 {
		return len(byteArr) == 0 // empty string is contained in empty string
	}
	for i := 0; i < len(byteArr); i++ {
		if byteArr[i] == byteArrSub[0] && len(byteArr)-i >= len(byteArrSub) {
			for j := 0; j < len(byteArrSub); j++ {
				if byteArr[i+j] != byteArrSub[j] {
					break
				}
				if j == len(byteArrSub)-1 {
					return true
				}
			}
		}
	}
	return false
}
