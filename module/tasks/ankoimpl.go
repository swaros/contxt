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

package tasks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mattn/anko/core"
	"github.com/mattn/anko/env"
	_ "github.com/mattn/anko/packages"
	"github.com/mattn/anko/parser"
	"github.com/mattn/anko/vm"
	"github.com/swaros/contxt/module/mimiclog"
)

type RiskLevel int

const (
	RISK_LEVEL_LOW RiskLevel = iota
	RISK_LEVEL_MEDIUM
	RISK_LEVEL_HIGH
)

type AnkoDefiner struct {
	symbol string      // the name of the symbol
	value  interface{} // the value of the symbol. this is a function or a variable
	risk   RiskLevel   // the risk level of the symbol
	help   string      // a help text for the symbol
}

type BufferHook func(msg string)

type AnkoException struct {
	Script string
	Err    error
}

type AnkoRunner struct {
	env           *env.Env
	defaults      []AnkoDefiner
	lazyInit      bool
	conTxt        context.Context
	options       vm.Options
	outBuffer     []string
	bufferPrepStr string
	hook          BufferHook
	lastScript    string
	lastError     error
	supressOutput bool
	cancelationFn context.CancelFunc
	timeoutCancel context.CancelFunc
	cancelation   bool
	timeOut       time.Duration
	exceptions    []AnkoException
	riskLevel     RiskLevel
	logger        mimiclog.Logger
}

func NewAnkoRunner() *AnkoRunner {
	return &AnkoRunner{
		env:           env.NewEnv(),
		lazyInit:      false,
		defaults:      []AnkoDefiner{},
		conTxt:        context.Background(),
		options:       vm.Options{},
		bufferPrepStr: "",
		outBuffer:     []string{},
		hook:          nil,
		lastScript:    "",
		lastError:     nil,
		supressOutput: false,
		cancelation:   false,
		timeOut:       0,
		exceptions:    []AnkoException{},
		riskLevel:     RISK_LEVEL_LOW,
		logger:        mimiclog.NewNullLogger(),
	}
}

func (ar *AnkoRunner) SetLogger(logger mimiclog.Logger) {
	ar.logger = logger
}

func (ar *AnkoRunner) AddDefaultDefine(symbol string, value interface{}, risk RiskLevel, help string) error {
	if ar.lazyInit {
		return errors.New("cannot add default define after initialization")
	}
	if risk > ar.riskLevel {
		return errors.New("command " + symbol + " has a higher risk level than the AnkoRunner:" + fmt.Sprintf("%d > %d", risk, ar.riskLevel))
	}
	ar.defaults = append(ar.defaults, AnkoDefiner{symbol, value, risk, help})
	return nil
}

// SetRiskLevel sets the risk level for the AnkoRunner.
// so any DefaultDefine that is added must be at the same or lower risk level
func (ar *AnkoRunner) SetRiskLevel(risk RiskLevel) {
	ar.riskLevel = risk
}

func (ar *AnkoRunner) SetTimeOut(to time.Duration) {
	ar.timeOut = to
	ar.conTxt, ar.timeoutCancel = context.WithTimeout(ar.conTxt, ar.timeOut)
}

func (ar *AnkoRunner) EnableCancelation() context.CancelFunc {
	ar.cancelation = true
	ar.conTxt, ar.cancelationFn = context.WithCancel(ar.conTxt)
	return ar.cancelationFn
}

func (ar *AnkoRunner) SetOutputSupression(sup bool) {
	ar.supressOutput = sup
}

func (ar *AnkoRunner) SetBufferHook(hook BufferHook) {
	ar.hook = hook
}

func (ar *AnkoRunner) SetOptions(opts vm.Options) {
	ar.options = opts
}

func (ar *AnkoRunner) SetContext(ctx context.Context) {
	ar.conTxt = ctx
}

func (ar *AnkoRunner) GetContext() context.Context {
	return ar.conTxt
}

func (ar *AnkoRunner) GetEnv() *env.Env {
	return ar.env
}

func (ar *AnkoRunner) toBuffer(msg ...interface{}) {
	addstr := ""
	for _, m := range msg {
		switch s := m.(type) {
		case string:
			addstr += s
		default:
			addstr += s.(string)
		}
	}
	if ar.hook != nil {
		ar.hook(addstr)
	}
	ar.outBuffer = append(ar.outBuffer, addstr)
}

func (ar *AnkoRunner) GetBuffer() []string {
	if ar.bufferPrepStr != "" {
		ar.toBuffer(ar.bufferPrepStr)
		ar.bufferPrepStr = ""
	}
	return ar.outBuffer
}

func (ar *AnkoRunner) ClearBuffer() {
	// calling GetBuffer will trigger the hook
	// for any remaining buffer that have still not been
	// closed with a newline
	ar.GetBuffer()
	ar.outBuffer = []string{}
}

func (ar *AnkoRunner) AddDefaultDefines(defs []AnkoDefiner) error {
	for _, def := range defs {
		err := ar.AddDefaultDefine(def.symbol, def.value, def.risk, def.help)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ar *AnkoRunner) handleBufferAdd(str string) {
	ar.bufferPrepStr += str
	// now put anything in the buffer, what ends with a newline
	slices := strings.Split(ar.bufferPrepStr, "\n")
	if len(slices) > 1 {
		for i := 0; i < len(slices)-1; i++ {
			ar.toBuffer(slices[i])
		}
		ar.bufferPrepStr = slices[len(slices)-1]
	}
}

func (ar *AnkoRunner) DefaultDefine() error {
	ar.env = core.Import(ar.env)
	for _, def := range ar.defaults {
		err := ar.env.Define(def.symbol, def.value)
		if err != nil {
			return err
		}
	}
	ar.defaultCmds() // anything we want to have as base functions (or redefined like print/println)
	ar.lazyInit = true
	return nil
}

func (ar *AnkoRunner) defaultCmds() {
	// set output handler to capture output
	ar.env.Define("print", func(msg ...interface{}) {
		ar.handleBufferAdd(fmt.Sprint(msg...))
		if !ar.supressOutput {
			fmt.Print(msg...)
		}

	})
	ar.env.Define("println", func(msg ...interface{}) {
		ar.handleBufferAdd(fmt.Sprintln(msg...))
		if !ar.supressOutput {
			fmt.Println(msg...)
		}
	})

	// a sleep function. always good to have
	ar.env.Define("sleep", func(milliSeconds int) {
		time.Sleep(time.Duration(milliSeconds) * time.Millisecond)
	})
}

func (ar *AnkoRunner) Defines(defs []AnkoDefiner) error {
	for _, def := range defs {
		err := ar.env.Define(def.symbol, def.value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ar *AnkoRunner) InitEnv() (*env.Env, error) {
	if !ar.lazyInit {
		if err := ar.DefaultDefine(); err != nil {
			return ar.env, err
		}
		ar.lazyInit = true
	}
	return ar.env, nil
}

func (ar *AnkoRunner) ThrowException(err error, script string) {
	ar.exceptions = append(ar.exceptions, AnkoException{script, err})
}

func (ar *AnkoRunner) ErrorPos2Message(err error) error {
	return errors.New(ar.ErrorExplain(err))
}

func (ar *AnkoRunner) ErrorExplain(err error) string {
	switch tErr := err.(type) {
	case *vm.Error:
		return fmt.Sprintf("Runtime Error: %s; line %d pos %d ", tErr.Message, tErr.Pos.Line, tErr.Pos.Column)
	case *parser.Error:
		return fmt.Sprintf("Parsing Error: %s; line %d pos %d fatal %v", tErr.Message, tErr.Pos.Line, tErr.Pos.Column, tErr.Fatal)
	default:
		return fmt.Sprintf("Error in script: %s errType: %T ", err.Error(), err)
	}
}

func (ar *AnkoRunner) Error2MsgErrorDebug(target string, err error) (MsgErrDebug, bool) {
	switch tErr := err.(type) {
	case *vm.Error:
		return MsgErrDebug{Target: target, Line: tErr.Pos.Line, Column: tErr.Pos.Column, Script: ar.lastScript, Err: tErr}, true
	case *parser.Error:
		return MsgErrDebug{Target: target, Line: tErr.Pos.Line, Column: tErr.Pos.Column, Script: ar.lastScript, Err: tErr}, true
	default:
		return MsgErrDebug{}, false
	}
}

func (ar *AnkoRunner) RunAnko(script string) (interface{}, error) {
	if _, err := ar.InitEnv(); err != nil {
		return nil, err
	}
	var res interface{}
	res, ar.lastError = vm.ExecuteContext(ar.conTxt, ar.env, &ar.options, script)

	// no error, but we have exceptions?
	if ar.lastError == nil {
		if len(ar.exceptions) > 0 {
			for _, ex := range ar.exceptions {
				ar.toBuffer(ar.ErrorExplain(ex.Err))
				ar.lastError = ex.Err
			}
			ar.exceptions = []AnkoException{}

		}
	}

	ar.lastScript = script
	return res, ar.lastError
}

func (ar *AnkoRunner) LastError() error {
	return ar.lastError
}

func (ar *AnkoRunner) LastScript() string {
	return ar.lastScript
}

func (ar *AnkoRunner) Define(symbol string, value interface{}) error {
	return ar.env.Define(symbol, value)
}
