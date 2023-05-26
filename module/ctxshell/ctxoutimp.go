package ctxshell

import (
	"fmt"
	"time"
)

func (t *Cshell) Stream(msg ...interface{}) {
	if t.rlInstance != nil {
		t.messages.Push("stdout", fmt.Sprint(msg...))
	} else {
		fmt.Print(msg...)
	}
}

func (t *Cshell) StreamLn(msg ...interface{}) {
	if t.rlInstance != nil {
		t.messages.Push("stdout", fmt.Sprintln(msg...))
	} else {
		fmt.Println(msg...)
	}
}

func (t *Cshell) Streamf(format string, msg ...interface{}) {
	if t.rlInstance != nil {
		t.messages.Push("stdout", fmt.Sprintf(format, msg...))
	} else {
		fmt.Printf(format, msg...)
	}
}

func (t *Cshell) StartMessageProvider() error {
	if t.rlInstance == nil {
		return fmt.Errorf("no readline instance initialized")
	}

	done := make(chan struct{})
	go func() {

	loop:
		for {
			select {
			case <-time.After(time.Duration(t.tickTimerDuration)):
				msgs := t.messages.GetAllMessages()
				for _, msg := range msgs {
					t.rlInstance.Stdout().Write([]byte(msg.Msg))
				}
			case <-done:
				break loop
			}
		}
		done <- struct{}{}
	}()
	return nil
}
