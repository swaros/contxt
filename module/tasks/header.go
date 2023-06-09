// Copyright (c) 2023 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// # Licensed under the MIT License
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
package tasks

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
)

var (
	emptyMainCmdSetter MainCmdSetter = emptyCmd{}
)

type targetExecuter struct {
	target          string
	arguments       map[string]string
	runCfg          configure.RunConfig
	mainCmd         string
	mainCmdArgs     []string
	phHandler       PlaceHolder
	outputHandler   func(msg ...interface{})
	requireHandler  Requires
	Logger          *logrus.Logger
	dataHandler     DataMapHandler
	watch           *Watchman
	commandFallback MainCmdSetter
}

type emptyCmd struct{}

func (e emptyCmd) GetMainCmd(cfg configure.Options) (string, []string) {
	return "", []string{}
}

func New(target string, arguments map[string]string, any ...interface{}) *targetExecuter {

	t := &targetExecuter{
		target:    target,
		arguments: arguments,
	}

	for i := 0; i < len(any); i++ {
		switch any[i].(type) {
		case configure.RunConfig:
			t.runCfg = any[i].(configure.RunConfig)

		case PlaceHolder:
			t.phHandler = any[i].(PlaceHolder)
			// check if if any[i] also implements the DataMapHandler interface
			// if so, and we do not have a data handler set yet
			// we set it to the one from the PlaceHolder
			if t.dataHandler == nil {
				if dm, ok := any[i].(DataMapHandler); ok {
					t.dataHandler = dm
				}
			}
		case func(msg ...interface{}):
			t.outputHandler = any[i].(func(msg ...interface{}))
		case Requires:
			t.requireHandler = any[i].(Requires)
		case DataMapHandler:
			t.dataHandler = any[i].(DataMapHandler)
			// check if if any[i] also implements the PlaceHolder interface
			// if so, and we do not have a placeholder handler set yet
			// we set it to the one from the DataMapHandler
			if t.phHandler == nil {
				if ph, ok := any[i].(PlaceHolder); ok {
					t.phHandler = ph
				}
			}
		case *Watchman:
			t.watch = any[i].(*Watchman)
		case MainCmdSetter:
			t.commandFallback = any[i].(MainCmdSetter)
		default:
			// print out the type of the given argument
			// so we can see what is wrong
			// and panic
			// so we can see the error
			// and fix it
			panic(fmt.Sprintf("Invalid type passed to New: %T", any[i]))
		}
	}

	t.reInitialize()
	return t
}

func (t *targetExecuter) SetMainCmd(mainCmd string, args ...string) *targetExecuter {
	t.mainCmd = mainCmd
	t.mainCmdArgs = args
	return t
}

// reInitialize is used to reinitialize the targetExecuter
// so it assigns the required fields depending the given arguments
// and also make sure, any required field is set
// if they can have a default value.
func (t *targetExecuter) reInitialize() {
	// this just returns the emptyCmd struct
	// so we can use it as a fallback
	// but will not usable so we have to warn the user
	if t.commandFallback == nil {
		t.commandFallback = emptyMainCmdSetter
		t.getLogger().Warn("No MainCmdSetter provided, using empty fallback")
	}
	// if no task watcher is set, we create a new one
	if t.watch == nil {
		t.watch = NewWatchman()
	}
	// assign the Tasks to the targetExecuter

}

func (t *targetExecuter) CopyToTarget(target string) *targetExecuter {
	copy := New(
		target,
		t.arguments,
		t.runCfg,
		t.phHandler,
		t.outputHandler,
		t.requireHandler,
		t.dataHandler,
		t.watch,
		t.commandFallback,
	)

	return copy
}

func (t *targetExecuter) SetLogger(logger *logrus.Logger) *targetExecuter {
	t.Logger = logger
	return t
}

func (t *targetExecuter) SetDataHandler(handler DataMapHandler) *targetExecuter {
	t.dataHandler = handler
	return t
}

func (t *targetExecuter) SetPlaceholderHandler(handler PlaceHolder) *targetExecuter {
	t.phHandler = handler
	return t
}

func (t *targetExecuter) SetWatchman(watch *Watchman) *targetExecuter {
	t.watch = watch
	return t
}
