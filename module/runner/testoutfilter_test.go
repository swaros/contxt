package runner_test

import "fmt"

// create a output handler to catch the created output while testing
type TestOutHandler struct {
	Msgs []string
}

func NewTestOutHandler() *TestOutHandler {
	return &TestOutHandler{}
}

func (t *TestOutHandler) Stream(msg ...interface{}) {
	t.Msgs = append(t.Msgs, fmt.Sprint(msg...))
}

func (t *TestOutHandler) StreamLn(msg ...interface{}) {
	t.Msgs = append(t.Msgs, fmt.Sprintln(msg...))
}

// get all the messages as a string
func (t *TestOutHandler) String() string {
	return fmt.Sprint(t.Msgs)
}

// get all the messages the are created
func (t *TestOutHandler) GetMessages() []string {
	return t.Msgs
}

// clear the messages
func (t *TestOutHandler) Clear() {
	t.Msgs = []string{}
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
	}
	return false
}
