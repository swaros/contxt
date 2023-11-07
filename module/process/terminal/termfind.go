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

package terminal

import (
	"errors"
	"runtime"
)

type TermFind struct {
	matchingOs     []string // the os to match
	cmd            string   // the command to run
	args           []string // the arguments to pass to the command and execute the commands. like the -c on bash
	argsToKeepOpen []string // the arguments to pass to the command to keep it open. like without -c for bash
}

var (
	termFindMap = map[string]*TermFind{
		"windows": {
			matchingOs:     []string{"windows"},
			cmd:            "powershell",
			args:           []string{"-nologo", "-noprofile"},
			argsToKeepOpen: []string{"-nologo", "-noprofile", "-noexit", "-NonInteractive"},
		},
		"linux": {
			matchingOs:     []string{"linux", "darwin"},
			cmd:            "bash",
			args:           []string{"-c"},
			argsToKeepOpen: []string{},
		},
	}
	ErrNoTerminalFound = errors.New("no terminal found")
)

// GetTerminal returns the terminal finder for the current os
// if the terminal is not found by the os keyword
// it will try to find it by the os name
func GetTerminal() (*TermFind, error) {
	if termFind, ok := termFindMap[runtime.GOOS]; ok {
		return termFind, nil
	} else {
		// did not found the terminal by the keyword
		// so we will try to find it by the os
		for _, termFind := range termFindMap {
			for _, os := range termFind.matchingOs {
				if os == runtime.GOOS {
					return termFind, nil
				}
			}
		}
	}
	return nil, ErrNoTerminalFound
}

// GetCmd returns the command to run the terminal
func (t *TermFind) GetCmd() string {
	return t.cmd
}

// GetArgs returns the arguments to pass to the command
// this is ment to be used to execute the command once
func (t *TermFind) GetArgs() []string {
	return t.args
}

// GetArgsToKeepOpen returns the arguments to pass to the command
// this is ment to be used to keep the command open so we can execute multiple commands
func (t *TermFind) GetArgsToKeepOpen() []string {
	return t.argsToKeepOpen
}

// CombineArgs combines the arguments to pass to the command with the given arguments
// this is ment to be used to execute the command once together with the given arguments
func (t *TermFind) CombineArgs(args ...string) []string {
	cmds := []string{}

	cmds = append(cmds, t.args...)
	cmds = append(cmds, args...)
	return cmds
}
