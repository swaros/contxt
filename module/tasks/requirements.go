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

package tasks

import (
	"fmt"
	"os"
	"strings"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/mimiclog"
)

type Requires interface {
	CheckRequirements(require configure.Require) (bool, string)
	CheckReason(checkReason configure.Trigger, output string, e error) (bool, string)
}

type DefaultRequires struct {
	variables PlaceHolder // taking care about the variables
	logger    mimiclog.Logger
}

var globalStickerForLastRequirementCheckCreated DefaultRequires

func NewDefaultRequires(variables PlaceHolder, logger mimiclog.Logger) *DefaultRequires {
	newReqChk := &DefaultRequires{
		variables: variables,
		logger:    logger,
	}

	if newReqChk.variables == nil {
		panic(fmt.Errorf("no variable handler set"))
	}
	if newReqChk.logger == nil {
		newReqChk.logger = mimiclog.NewNullLogger()
	}
	globalStickerForLastRequirementCheckCreated = *newReqChk
	return newReqChk
}

func RequirementCheck(require configure.Require) (bool, string) {
	if globalStickerForLastRequirementCheckCreated.variables == nil {
		return true, ""
	} else {
		return globalStickerForLastRequirementCheckCreated.CheckRequirements(require)
	}
}

// test all requirements
// - operation system
// - variable matches
// - files exists
// - files not exists
// - environment variables
// returns bool and the message what is checked.
func (d *DefaultRequires) CheckRequirements(require configure.Require) (bool, string) {

	// check operating system
	if require.System != "" {
		match := d.StringMatchTest(require.System, configure.GetOs())
		if !match {
			return false, "operating system '" + configure.GetOs() + "' is not matching with '" + require.System + "'"
		}
	}

	// check file exists
	for _, fileExists := range require.Exists {
		fileExists = d.variables.HandlePlaceHolder(fileExists)
		fexists, err := dirhandle.Exists(fileExists)
		if err != nil || !fexists {

			return false, "required file (" + fileExists + ") not found "
		}
	}

	// check file not exists
	for _, fileNotExists := range require.NotExists {
		fileNotExists = d.variables.HandlePlaceHolder(fileNotExists)
		fexists, err := dirhandle.Exists(fileNotExists)
		if err != nil || fexists {
			return false, "unexpected file (" + fileNotExists + ")  found "
		}
	}

	// check environment variable is set
	for name, pattern := range require.Environment {
		envVar, envExists := os.LookupEnv(name)
		if !envExists || !d.StringMatchTest(pattern, envVar) {
			if envExists {
				return false, "environment variable[" + name + "] not matching with " + pattern
			}
			return false, "environment variable[" + name + "] not exists"
		}
	}

	// check variables
	for name, pattern := range require.Variables {
		defVar, defExists := d.variables.GetPHExists(name)
		if !defExists || !d.StringMatchTest(pattern, defVar) {
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
func (d *DefaultRequires) StringMatchTest(pattern, value string) bool {
	first := pattern
	maybeMatch := value
	if len(pattern) > 1 {
		maybeMatch = pattern[1:]
		first = pattern[0:1]
	}
	switch first {
	// string not empty
	case "?":
		logFields := mimiclog.Fields{"pattern": pattern, "value": value, "check": first, "result": (value != "")}
		d.logger.Debug("check anything then empty", logFields)
		return (value != "")

	// string equals
	case "=":
		logFields := mimiclog.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch == value)}
		d.logger.Debug("check equal", logFields)
		return (maybeMatch == value)

	// string not equals
	case "!":
		logFields := mimiclog.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch != value)}
		d.logger.Debug("check not equal", logFields)
		return (maybeMatch != value)

	// string greater
	case ">":
		logFields := mimiclog.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch < value)}
		d.logger.Debug("check greather then", logFields)
		return (value > maybeMatch)

	// sstring lower
	case "<":
		logFields := mimiclog.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch > value)}
		d.logger.Debug("check lower then", logFields)
		return (value < maybeMatch)

	// string not empty
	case "*":
		logFields := mimiclog.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (value != "")}
		d.logger.Debug("check empty *", logFields)
		return (value != "")

	// default is checking equals
	default:
		logFields := mimiclog.Fields{"pattern": pattern, "value": value, "check": first, "result": (pattern == value)}
		d.logger.Debug("default: check equal against plain values", logFields)
		return (pattern == value)
	}
}

// checks reasons that is used for some triggers
// returns bool and a message what trigger was matched and the reason
func (d *DefaultRequires) CheckReason(checkReason configure.Trigger, output string, e error) (bool, string) {
	logFields := mimiclog.Fields{
		"contains":   checkReason.OnoutContains,
		"onError":    checkReason.Onerror,
		"onLess":     checkReason.OnoutcountLess,
		"onMore":     checkReason.OnoutcountMore,
		"testing-at": output}
	d.logger.Debug("Checking Trigger", logFields)

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
		checkText = d.variables.HandlePlaceHolder(checkText)
		if checkText != "" && strings.Contains(output, checkText) {
			message = fmt.Sprint("reason match because output contains ", checkText)
			return true, message
		}
		if checkText != "" {
			logFields := mimiclog.Fields{
				"check": checkText,
				"with":  output,
				"from":  checkReason.OnoutContains,
			}
			d.logger.Debug("OnoutContains NO MATCH", logFields)
		}
	}

	return false, message
}
