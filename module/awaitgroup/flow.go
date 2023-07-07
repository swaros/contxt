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

 package awaitgroup

import (
	"context"
	"errors"
	"reflect"

	"github.com/google/uuid"
)

func GetArgument(ctx context.Context) interface{} {
	return ctx.Value(CtxKey{})
}

type Flow struct {
	function  func(args ...interface{}) []interface{} // the logic that should be executed
	handler   func(args ...interface{})               // the logic that handles the return values
	args      map[string]ArgContext                   // the arguments they are passed to the function
	argOrder  []string                                // the order of the arguments and also the function execution order and return order
	lastError error                                   // the last error that occured
}

type ArgContext struct {
	Arguments []interface{}
}

type ReturnContexr struct {
	ReturnValue []interface{}
}

func NewFlow() *Flow {
	return &Flow{
		args: make(map[string]ArgContext),
	}
}

func (f *Flow) Use(fn interface{}) error {
	// fn should be a function and we we have to figure out
	// the arguments and the return values
	// and then we have to create a new function that will
	// be used as the function

	if fn != nil {
		fType := reflect.TypeOf(fn)
		if fType.Kind() != reflect.Func {
			return errors.New("argument must be a function")
		}
		f.function = func(args ...interface{}) []interface{} {
			fVal := reflect.ValueOf(fn)
			if len(args) != fVal.Type().NumIn() {
				f.lastError = errors.New("number of arguments does not match")
				return nil
			}

			for i := 0; i < fVal.Type().NumIn(); i++ {
				if reflect.TypeOf(args[i]) != fVal.Type().In(i) {
					f.lastError = errors.New("argument type does not match")
					return nil
				}
			}

			var arguments []reflect.Value
			for _, arg := range args {
				arguments = append(arguments, reflect.ValueOf(arg))
			}
			returnValues := fVal.Call(arguments)
			var returnValuesInterface []interface{}
			for _, returnValue := range returnValues {
				returnValuesInterface = append(returnValuesInterface, returnValue.Interface())
			}
			return returnValuesInterface
		}
	} else {
		return errors.New("argument must be a function. got nil")
	}

	return nil
}

func (f *Flow) Func(fn func(args ...interface{}) []interface{}) *Flow {
	f.function = fn
	return f
}

func (f *Flow) Handler(fn func(args ...interface{})) *Flow {
	f.handler = fn
	return f
}

func (f *Flow) Go(arg ...interface{}) *Flow {
	// create and return a new uuid
	// that will be used as a key
	uuid := uuid.New().String()
	f.args[uuid] = ArgContext{
		Arguments: arg,
	}
	f.argOrder = append(f.argOrder, uuid)
	return f
}

func (f *Flow) Run() error {
	f.lastError = nil
	if f.function == nil {
		return errors.New("function is not defined")
	}
	var allTasks []FutureStack
	for _, arg := range f.argOrder {
		allTasks = append(allTasks, FutureStack{
			AwaitFunc: func(ctx context.Context) interface{} {
				arg := GetArgument(ctx).(ArgContext)
				resturnFromFunc := f.function(arg.Arguments...)
				return ReturnContexr{
					ReturnValue: resturnFromFunc,
				}
			},
			Argument: f.args[arg],
		})
	}
	futures := ExecFutureGroup(allTasks)
	results := WaitAtGroup(futures)
	for _, result := range results {
		if f.handler != nil {
			f.handler(result.(ReturnContexr).ReturnValue...)
		}
	}
	return f.lastError
}
