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

package runner

import (
	"fmt"
	"strings"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/tasks"
)

var (
	// random color helper
	randColors = NewRandColorStore()
	// color for the state label
	stateColorPreDef = ctxout.ToString(ctxout.NewMOWrap(), ctxout.ForeLightCyan+ctxout.BoldTag+ctxout.BackBlue)
	processPreDef    = ctxout.ToString(ctxout.NewMOWrap(), ctxout.ForeLightCyan+ctxout.BackBlack)
	pidPreDef        = ctxout.ToString(ctxout.NewMOWrap(), ctxout.ForeLightYellow+ctxout.BackBlack)
	commentPreDef    = ctxout.ToString(ctxout.NewMOWrap(), ctxout.ForeLightBlue+ctxout.BackBlack)
)

// handles all the incomming messages from the tasks
// depending on the message type it will print the message.
// for this we parsing at the fist level the message type for each message.

func (c *CmdExecutorImpl) setOutHandler(name string) (func(msg ...interface{}), error) {
	if hndl, ok := c.getHandlerByName(name); ok {
		c.usedHandler = hndl.GetName()
		return hndl.GetOutHandler(c), nil
	}
	return nil, fmt.Errorf("no handler found for %s", name)
}

func (c *CmdExecutorImpl) addOutHandler(outHndl ...OutputHandler) {
	for _, hndl := range outHndl {
		name := hndl.GetName()
		c.outHandlers[name] = &hndl
	}
}

func (c *CmdExecutorImpl) getHandlerByName(name string) (OutputHandler, bool) {
	if hndl, ok := c.outHandlers[name]; ok {
		return *hndl, true
	}
	return nil, false
}

func (c *CmdExecutorImpl) GetAllOutputHandlerNames() []string {
	names := make([]string, 0, len(c.outHandlers))
	for name := range c.outHandlers {
		names = append(names, name)
	}
	return names
}

func (c *CmdExecutorImpl) formatDebugError(err tasks.MsgErrDebug) string {
	lines := strings.Split(err.Script, "\n")
	msg := "can not format error"
	lineNr := err.Line - 1
	wordPos := err.Column - 1
	if len(lines) >= lineNr {
		lineCode := lines[lineNr]
		if len(lineCode) > wordPos {
			left := lineCode[:wordPos]
			right := lineCode[wordPos:]
			msg = fmt.Sprintf("Command Error: %s\n%s%s --- %s [%d]", err.Err.Error(), ctxout.ForeDarkGrey+left, ctxout.ForeRed+right+ctxout.CleanTag, err.Err.Error(), err.Column)
		} else {
			msg += " (column out of range)"
			msg += fmt.Sprintf(" ... Command Error: %s [%d:%d]", err.Err.Error(), err.Column, err.Line)
		}

	} else {
		msg += " (line out of range)"
	}
	return msg
}
