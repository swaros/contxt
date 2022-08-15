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

package taskrun

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/configure"
	"github.com/swaros/contxt/dirhandle"
)

// test all requirements
// - operation system
// - variable matches
// - files exists
// - files not exists
// - environment variables
// returns bool and the message what is checked.
func checkRequirements(require configure.Require) (bool, string) {

	// check operating system
	if require.System != "" {
		match := StringMatchTest(require.System, configure.GetOs())
		if !match {
			return false, "operating system '" + configure.GetOs() + "' is not matching with '" + require.System + "'"
		}
	}

	// check file exists
	for _, fileExists := range require.Exists {
		fileExists = HandlePlaceHolder(fileExists)
		fexists, err := dirhandle.Exists(fileExists)
		GetLogger().WithFields(logrus.Fields{
			"path":   fileExists,
			"result": fexists,
		}).Debug("path exists? result=true means valid for require")
		if err != nil || !fexists {

			return false, "required file (" + fileExists + ") not found "
		}
	}

	// check file not exists
	for _, fileNotExists := range require.NotExists {
		fileNotExists = HandlePlaceHolder(fileNotExists)
		fexists, err := dirhandle.Exists(fileNotExists)
		GetLogger().WithFields(logrus.Fields{
			"path":   fileNotExists,
			"result": fexists,
		}).Debug("path NOT exists? result=true means not valid for require")
		if err != nil || fexists {
			return false, "unexpected file (" + fileNotExists + ")  found "
		}
	}

	// check environment variable is set
	for name, pattern := range require.Environment {
		envVar, envExists := os.LookupEnv(name)
		if !envExists || !StringMatchTest(pattern, envVar) {
			if envExists {
				return false, "environment variable[" + name + "] not matching with " + pattern
			}
			return false, "environment variable[" + name + "] not exists"
		}
	}

	// check variables
	for name, pattern := range require.Variables {
		defVar, defExists := GetPHExists(name)
		if !defExists || !StringMatchTest(pattern, defVar) {
			if defExists {
				return false, "runtime variable[" + name + "] not matching with " + pattern
			}
			return false, "runtime variable[" + name + "] not exists "
		}
	}

	return true, ""
}

// StringMatchTest test a pattern and a value.
// in this example: myvar: "=hello"
// the patter is "=hello" and the value should be "hello" for a match
func StringMatchTest(pattern, value string) bool {
	first := pattern
	maybeMatch := value
	if len(pattern) > 1 {
		maybeMatch = pattern[1:]
		first = pattern[0:1]
	}
	switch first {
	// string not empty
	case "?":
		GetLogger().WithFields(logrus.Fields{"pattern": pattern, "value": value, "check": first, "result": (value != "")}).Debug("check anything then empty")
		return (value != "")

	// string equals
	case "=":
		GetLogger().WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch == value)}).Debug("check equal")
		return (maybeMatch == value)

	// string not equals
	case "!":
		GetLogger().WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch != value)}).Debug("check not equal")
		return (maybeMatch != value)

	// string greater
	case ">":
		GetLogger().WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch < value)}).Debug("check greather then")
		return (value > maybeMatch)

	// sstring lower
	case "<":
		GetLogger().WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch > value)}).Debug("check lower then")
		return (value < maybeMatch)

	// string not empty
	case "*":
		GetLogger().WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (value != "")}).Debug("check empty *")
		return (value != "")

	// default is checking equals
	default:
		GetLogger().WithFields(logrus.Fields{"pattern": pattern, "value": value, "check": first, "result": (pattern == value)}).Debug("default: check equal against plain values")
		return (pattern == value)
	}
}

// checks reasons that is used for some triggers
// retunrs bool and a message what trigger was matched and the reason
func checkReason(checkReason configure.Trigger, output string, e error) (bool, string) {
	GetLogger().WithFields(logrus.Fields{
		"contains":   checkReason.OnoutContains,
		"onError":    checkReason.Onerror,
		"onLess":     checkReason.OnoutcountLess,
		"onMore":     checkReason.OnoutcountMore,
		"testing-at": output,
	}).Debug("Checking Trigger")

	var message = ""
	// now means forcing this trigger
	if checkReason.Now {
		message = "reason now match always"
		return true, message
	}
	// checks a error happens
	if checkReason.Onerror && e != nil {
		message = fmt.Sprint("reason match because a error happen (", e, ")  ")
		return true, message
	}

	// checks if the output from comand contains les then X chars
	if checkReason.OnoutcountLess > 0 && checkReason.OnoutcountLess > len(output) {
		message = fmt.Sprint("reason match output len (", len(output), ") is less then ", checkReason.OnoutcountLess)
		return true, message
	}
	// checks if the output contains more then X chars
	if checkReason.OnoutcountMore > 0 && checkReason.OnoutcountMore < len(output) {
		message = fmt.Sprint("reason match output len (", len(output), ") is more then ", checkReason.OnoutcountMore)
		return true, message
	}

	// checks if the output contains one of the defined text-lines
	for _, checkText := range checkReason.OnoutContains {
		checkText = HandlePlaceHolder(checkText)
		if checkText != "" && strings.Contains(output, checkText) {
			message = fmt.Sprint("reason match because output contains ", checkText)
			return true, message
		}
		if checkText != "" {
			GetLogger().WithFields(logrus.Fields{
				"check": checkText,
				"with":  output,
				"from":  checkReason.OnoutContains,
			}).Debug("OnoutContains NO MATCH")
		}
	}

	return false, message
}
