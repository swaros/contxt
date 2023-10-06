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
