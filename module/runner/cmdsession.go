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
package runner

import (
	"github.com/swaros/contxt/module/ctemplate"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/mimiclog"
)

type CmdSession struct {
	Log              *SessionLogger         // the whole logging stuff
	TemplateHndl     *ctemplate.Template    // template handler they execute the template
	Cobra            *SessionCobra          // the cobra command handler
	OutPutHdnl       ctxout.StreamInterface // used for output stream
	Printer          ctxout.PrintInterface  // used formated printing to the console
	DefaultVariables map[string]string      // DefaultVariables are variables which are set for every task. they are predefines. not the used variables itself
	SharedHelper     *SharedHelper          // SharedHelper is responsible for the shared tasks.
}

type SessionLogger struct {
	LogLevel string
	Logger   mimiclog.Logger
}

func NewCmdSession() *CmdSession {
	logger := NewLogrusLogger()
	templateHndl := ctemplate.New()
	templateHndl.SetIgnoreHndl(true)
	session := &CmdSession{
		OutPutHdnl:       ctxout.NewFmtWrap(),
		Printer:          ctxout.NewMOWrap(),
		Cobra:            NewCobraCmds(),
		TemplateHndl:     templateHndl,
		SharedHelper:     NewSharedHelper(),
		DefaultVariables: make(map[string]string),
		Log: &SessionLogger{
			LogLevel: "error",
			Logger:   logger,
		},
	}
	mimiclog.ApplyLogger(logger, session.TemplateHndl)
	mimiclog.ApplyIfPossible(logger, session.TemplateHndl)
	mimiclog.ApplyIfPossible(logger, session.SharedHelper)
	return session
}
