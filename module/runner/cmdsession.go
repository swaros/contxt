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
	Log          *SessionLogger
	TemplateHndl *ctemplate.Template
	Cobra        *SessionCobra
	OutPutHdnl   ctxout.StreamInterface
	Printer      ctxout.PrintInterface
}

type SessionLogger struct {
	LogLevel string
	Logger   mimiclog.Logger
}

func NewCmdSession() *CmdSession {
	return &CmdSession{
		OutPutHdnl:   ctxout.NewFmtWrap(),
		Printer:      ctxout.NewMOWrap(),
		Cobra:        NewCobraCmds(),
		TemplateHndl: ctemplate.New(),
		Log: &SessionLogger{
			LogLevel: "error",
			Logger:   NewLogrusLogger(),
		},
	}
}
