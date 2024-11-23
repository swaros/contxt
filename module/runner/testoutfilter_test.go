package runner_test

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/swaros/contxt/module/systools"
)

// create a output handler to catch the created output while testing
type TestOutHandler struct {
	Msgs         []string
	logFile      string
	keepNewLines bool
}

func NewTestOutHandler() *TestOutHandler {
	return &TestOutHandler{}
}

func (t *TestOutHandler) SetLogFile(logFile string) {
	logFile = strings.ReplaceAll(logFile, ":", "_")
	logFile = strings.ReplaceAll(logFile, "-", "_")
	logFile = strings.ReplaceAll(logFile, "+", "_")
	t.logFile = logFile
}

func filterMessages(msgs []string) []string {
	filtered := []string{}
	for _, msg := range msgs {
		filteredStr := systools.PrintableChars(msg)
		filtered = append(filtered, filteredStr)

	}
	return filtered
}

// enable or disable the new line filter
func (t *TestOutHandler) SetKeepNewLines(keep bool) {
	t.keepNewLines = keep
}

func (t *TestOutHandler) Stream(msg ...interface{}) {
	t.Msgs = append(t.Msgs, fmt.Sprint(msg...))
}

func (t *TestOutHandler) StreamLn(msg ...interface{}) {
	t.Msgs = append(t.Msgs, fmt.Sprintln(msg...))
}

// get all the messages as a string
func (t *TestOutHandler) String() string {
	if t.keepNewLines {
		return fmt.Sprintln(strings.Join(t.Msgs, "\n"))
	}
	return fmt.Sprintln(strings.Join(filterMessages(t.Msgs), "\n"))
}

// get all the messages the are created
func (t *TestOutHandler) GetMessages() []string {
	return t.Msgs
}

// clear the messages
func (t *TestOutHandler) Clear() {
	t.Msgs = []string{}
}

func (t *TestOutHandler) ClearAndLog() {
	t.WriteToLogFile()
	t.Clear()
}

func (t *TestOutHandler) GetLogFile() string {
	return t.logFile
}

func (t *TestOutHandler) WriteToLogFile() error {
	if t.logFile == "" {
		return nil
	}
	f, err := os.OpenFile(t.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err == nil {
		defer f.Close()
		if _, err := io.WriteString(f, t.String()); err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

// check if the message is in the output
func (t *TestOutHandler) Contains(msg string) bool {
	for _, m := range t.Msgs {
		if m == msg {
			return true
		}
		if m == msg+"\n" {
			return true
		}
		if strings.Contains(m, msg) {
			return true
		}
		if systools.StrContains(m, msg) {
			return true
		}
	}
	return false
}

func (t *TestOutHandler) TestRegexPattern(pattern string) bool {
	for _, m := range t.Msgs {
		regexp, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		if regexp.MatchString(m) {
			return true
		}
	}
	return false
}

// check if the message is in the output
func (t *TestOutHandler) Get(msg string) []string {
	matches := []string{}
	for _, m := range t.Msgs {

		if strings.Contains(m, msg) {
			matches = append(matches, m)
		}
	}
	return matches
}
