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

import "github.com/chzyer/readline"

type NativeCmd struct {
	Name          string
	Help          string
	ExecFunc      func(args []string) error
	CompleterFunc readline.DynamicCompleteFunc
}

func NewNativeCmd(name, help string, execFunc func(args []string) error) *NativeCmd {
	return &NativeCmd{
		Name:     name,
		Help:     help,
		ExecFunc: execFunc,
	}
}

func (t *NativeCmd) SetCompleterFunc(f readline.DynamicCompleteFunc) *NativeCmd {
	t.CompleterFunc = f
	return t
}

func (t *NativeCmd) Exec(args []string) error {
	return t.ExecFunc(args)
}

func (t *NativeCmd) GetHelp() string {
	return t.Help
}

func (t *NativeCmd) GetName() string {
	return t.Name
}
