package tasks

import (
	"context"
	"errors"
	"fmt"

	"github.com/mattn/anko/core"
	"github.com/mattn/anko/env"
	"github.com/mattn/anko/vm"
)

type AnkoDefiner struct {
	symbol string
	value  interface{}
}

type AnkoRunner struct {
	env       *env.Env
	defaults  []AnkoDefiner
	lazyInit  bool
	conTxt    context.Context
	options   vm.Options
	outBuffer []string
}

func NewAnkoRunner() *AnkoRunner {
	return &AnkoRunner{
		env:      env.NewEnv(),
		lazyInit: false,
		defaults: []AnkoDefiner{},
		conTxt:   context.Background(),
		options:  vm.Options{},
	}
}

func (ar *AnkoRunner) AddDefaultDefine(symbol string, value interface{}) error {
	if ar.lazyInit {
		return errors.New("cannot add default define after initialization")
	}
	ar.defaults = append(ar.defaults, AnkoDefiner{symbol, value})
	return nil
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
	ar.outBuffer = append(ar.outBuffer, addstr)
}

func (ar *AnkoRunner) GetBuffer() []string {
	return ar.outBuffer
}

func (ar *AnkoRunner) ClearBuffer() {
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
		ar.toBuffer(msg...)
		fmt.Print(msg...)

	})
	ar.env.Define("println", func(msg ...interface{}) {
		ar.toBuffer(msg...)
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
	return vm.ExecuteContext(ar.conTxt, ar.env, &ar.options, script)
}

func (ar *AnkoRunner) Define(symbol string, value interface{}) error {
	return ar.env.Define(symbol, value)
}
