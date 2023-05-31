package runner_test

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// create a output handler to catch the created output while testing
type TestOutHandler struct {
	Msgs    []string
	logFile string
}

func NewTestOutHandler() *TestOutHandler {
	return &TestOutHandler{}
}

func (t *TestOutHandler) SetLogFile(logFile string) {
	t.logFile = logFile
}

func (t *TestOutHandler) Stream(msg ...interface{}) {
	t.Msgs = append(t.Msgs, fmt.Sprint(msg...))
}

func (t *TestOutHandler) StreamLn(msg ...interface{}) {
	t.Msgs = append(t.Msgs, fmt.Sprintln(msg...))
}

// get all the messages as a string
func (t *TestOutHandler) String() string {
	return fmt.Sprintln(t.Msgs)
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
	}
	return false
}
