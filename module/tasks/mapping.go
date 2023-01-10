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
	"os"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
)

func (t *targetExecuter) out(msg ...interface{}) {
	if t.outputHandler != nil {
		t.outputHandler(msg...)
	}
}

func (t *targetExecuter) getLogger() *logrus.Logger {
	if t.Logger == nil {
		t.Logger = logrus.New()
		t.Logger.SetFormatter(&logrus.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
		t.Logger.SetOutput(os.Stdout)
	}
	return t.Logger
}

func (t *targetExecuter) SetArgs(args map[string]string) {
	t.arguments = args
}

func (t *targetExecuter) setPh(name, value string) {
	if t.phHandler != nil {
		t.phHandler.SetPH(name, value)
	}
}

func (t *targetExecuter) getPh(input string) string {
	if t.phHandler != nil {
		return t.phHandler.GetPH(input)
	}
	return input
}

func (t *targetExecuter) checkReason(reason configure.Trigger, output string, e error) (bool, string) {
	if t.reasonCheck != nil {
		return t.reasonCheck(reason, output, e)
	}
	return false, ""
}

func (t *targetExecuter) checkRequirements(require configure.Require) (bool, string) {
	if t.checkReqs != nil {
		return t.checkReqs(require)
	}
	return false, "no requirement check handler set"
}

func (t *targetExecuter) GetWatch() *Watchman {
	return t.watch
}
