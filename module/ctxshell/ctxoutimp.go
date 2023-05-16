package ctxshell

import "fmt"

func (t *Cshell) Stream(msg ...interface{}) {
	if t.rlInstance != nil {
		t.rlInstance.Stdout().Write([]byte(fmt.Sprint(msg...)))
	} else {
		fmt.Print(msg...)
	}
}

func (t *Cshell) StreamLn(msg ...interface{}) {
	if t.rlInstance != nil {
		t.rlInstance.Stdout().Write([]byte(fmt.Sprintln(msg...)))
	} else {
		fmt.Println(msg...)
	}
}
