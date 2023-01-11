package tasks

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
)

type Requires interface {
	CheckRequirements(require configure.Require) (bool, string)
	CheckReason(checkReason configure.Trigger, output string, e error) (bool, string)
}

type DefaultRequires struct {
	variables PlaceHolder // taking care about the variables
	logger    *logrus.Logger
}

var globalStickerForLastRequirementCheckCreated DefaultRequires

func NewDefaultRequires(variables PlaceHolder, logger *logrus.Logger) *DefaultRequires {
	newReqChk := &DefaultRequires{
		variables: variables,
		logger:    logger,
	}

	if newReqChk.variables == nil {
		panic(fmt.Errorf("no variable handler set"))
	}
	if newReqChk.logger == nil {
		newReqChk.logger = logrus.New()
		newReqChk.logger.SetFormatter(&logrus.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
		newReqChk.logger.SetOutput(os.Stdout)
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
		d.logger.WithFields(logrus.Fields{"pattern": pattern, "value": value, "check": first, "result": (value != "")}).Debug("check anything then empty")
		return (value != "")

	// string equals
	case "=":
		d.logger.WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch == value)}).Debug("check equal")
		return (maybeMatch == value)

	// string not equals
	case "!":
		d.logger.WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch != value)}).Debug("check not equal")
		return (maybeMatch != value)

	// string greater
	case ">":
		d.logger.WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch < value)}).Debug("check greather then")
		return (value > maybeMatch)

	// sstring lower
	case "<":
		d.logger.WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (maybeMatch > value)}).Debug("check lower then")
		return (value < maybeMatch)

	// string not empty
	case "*":
		d.logger.WithFields(logrus.Fields{"pattern": maybeMatch, "value": value, "check": first, "result": (value != "")}).Debug("check empty *")
		return (value != "")

	// default is checking equals
	default:
		d.logger.WithFields(logrus.Fields{"pattern": pattern, "value": value, "check": first, "result": (pattern == value)}).Debug("default: check equal against plain values")
		return (pattern == value)
	}
}

// checks reasons that is used for some triggers
// returns bool and a message what trigger was matched and the reason
func (d *DefaultRequires) CheckReason(checkReason configure.Trigger, output string, e error) (bool, string) {
	d.logger.WithFields(logrus.Fields{
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
		checkText = d.variables.HandlePlaceHolder(checkText)
		if checkText != "" && strings.Contains(output, checkText) {
			message = fmt.Sprint("reason match because output contains ", checkText)
			return true, message
		}
		if checkText != "" {
			d.logger.WithFields(logrus.Fields{
				"check": checkText,
				"with":  output,
				"from":  checkReason.OnoutContains,
			}).Debug("OnoutContains NO MATCH")
		}
	}

	return false, message
}
