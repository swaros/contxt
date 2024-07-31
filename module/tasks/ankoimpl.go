package tasks

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mattn/anko/core"
	"github.com/mattn/anko/env"
	"github.com/mattn/anko/vm"
)

type AnkoDefiner struct {
	symbol string
	value  interface{}
}

type BufferHook func(msg string)

type AnkoRunner struct {
	env           *env.Env
	defaults      []AnkoDefiner
	lazyInit      bool
	conTxt        context.Context
	options       vm.Options
	outBuffer     []string
	bufferPrepStr string
	hook          BufferHook
}

func NewAnkoRunner() *AnkoRunner {
	return &AnkoRunner{
		env:           env.NewEnv(),
		lazyInit:      false,
		defaults:      []AnkoDefiner{},
		conTxt:        context.Background(),
		options:       vm.Options{},
		bufferPrepStr: "",
	}
}

func (ar *AnkoRunner) AddDefaultDefine(symbol string, value interface{}) error {
	if ar.lazyInit {
		return errors.New("cannot add default define after initialization")
	}
	ar.defaults = append(ar.defaults, AnkoDefiner{symbol, value})
	return nil
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
		err := ar.AddDefaultDefine(def.symbol, def.value)
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
	// set output handler to capture output
	ar.env.Define("print", func(msg ...interface{}) {
		ar.handleBufferAdd(fmt.Sprint(msg...))
		fmt.Print(msg...)

	})
	ar.env.Define("println", func(msg ...interface{}) {
		ar.handleBufferAdd(fmt.Sprintln(msg...))
		fmt.Println(msg...)
	})
	ar.lazyInit = true
	return nil
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

func (ar *AnkoRunner) RunAnko(script string) (interface{}, error) {
	if !ar.lazyInit {
		if err := ar.DefaultDefine(); err != nil {
			return nil, err
		}
		ar.lazyInit = true
	}
	defer ar.ClearBuffer()
	return vm.ExecuteContext(ar.conTxt, ar.env, &ar.options, script)
}

func (ar *AnkoRunner) Define(symbol string, value interface{}) error {
	return ar.env.Define(symbol, value)
}
