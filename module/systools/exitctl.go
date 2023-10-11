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

package systools

import (
	"os"
	"os/signal"
)

type ExitBehavior struct {
	proceedWithExit bool
}

var (
	Continue  ExitBehavior = ExitBehavior{proceedWithExit: true}
	Interrupt ExitBehavior = ExitBehavior{proceedWithExit: false}
)

// contains all listener they should be executed
// if we want to exit the app, so some cleanup can be executed.
var exitListener map[string]func(int) ExitBehavior = make(map[string]func(int) ExitBehavior)

// adds a callback as listener
func AddExitListener(name string, callbk func(int) ExitBehavior) {
	exitListener[name] = callbk
}

// Exit maps the os.Exit but
// executes all callbacks before
// it the exit was aborted, you will get
// false in return
func Exit(code int) bool {
	for _, listener := range exitListener {
		if behave := listener(code); !behave.proceedWithExit {
			return false
		}
	}
	os.Exit(code)
	return true
}

// WatchSigTerm adds a callback function
// that will be executed if the app receives
// a SIGTERM signal
// if no callback is given, the app will exit
// with code 0 and all registered callbacks will
// be executed
func WatchSigTerm(callback func(os.Signal)) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if callback != nil {
				callback(sig)
			} else {
				Exit(0)
			}
		}
	}() // exit
}
