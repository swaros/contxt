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
	"errors"
	"reflect"
)

type FluxDevice struct {
	futures              []Future
	function             func(args ...interface{}) []interface{} // the logic that should be executed
	lastError            error                                   // the last error that occured
	isSetByReflect       bool                                    // flag if we use a method, that is set by reflection
	reflectFn            *ReflectFunc                            // the reflect function
	resultIsSEtByREflect bool                                    // flag if the result is set by reflection
	reflectResult        *ReflectFunc                            // the reflect result
}

type ReflectFunc struct {
	fn          interface{}
	argCount    int
	returnCount int
	argValue    reflect.Value
}

func NewFluxDevice(fn interface{}) (*FluxDevice, error) {
	fd := &FluxDevice{}
	if err := fd.Use(fn); err != nil {
		return nil, err
	}
	return fd, nil
}

func (f *FluxDevice) Fn(fn func(args ...interface{}) []interface{}) *FluxDevice {
	f.function = fn
	return f
}

func (f *FluxDevice) Use(fn interface{}) error {
	// fn should be a function and we we have to figure out
	// the arguments and the return values
	// and then we have to create a new function that will
	// be used as the function

	if fn != nil {
		fType := reflect.TypeOf(fn)
		if fType.Kind() != reflect.Func {
			return errors.New("argument must be a function")
		}

		f.isSetByReflect = true
		f.reflectFn = &ReflectFunc{
			fn:          fn,
			argCount:    fType.NumIn(),
			returnCount: fType.NumOut(),
			argValue:    reflect.ValueOf(fn),
		}

		f.function = func(args ...interface{}) []interface{} {
			if len(args) != f.reflectFn.argCount {
				f.lastError = errors.New("number of arguments does not match")
				return nil
			}

			for i := 0; i < f.reflectFn.argCount; i++ {
				if reflect.TypeOf(args[i]) != f.reflectFn.argValue.Type().In(i) {
					f.lastError = errors.New("argument type does not match")
					return nil
				}
			}

			var arguments []reflect.Value
			for _, arg := range args {
				arguments = append(arguments, reflect.ValueOf(arg))
			}
			returnValues := f.reflectFn.argValue.Call(arguments)
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

func (f *FluxDevice) Run(args ...interface{}) {
	future := ExecFuture(args, func() interface{} {
		return f.function(args...)
	})
	f.futures = append(f.futures, future)

}

func (f *FluxDevice) Results() []interface{} {
	var results []interface{}
	for _, future := range f.futures {
		results = append(results, future.Await())
	}
	return results
}

func (f *FluxDevice) ResultsTo(funcRef interface{}) error {
	if !f.isSetByReflect {
		return errors.New("you need to set the main function by reflection first. use Use()")
	}
	fType := reflect.TypeOf(funcRef)
	if fType.Kind() != reflect.Func {
		return errors.New("argument must be a function")
	}
	f.resultIsSEtByREflect = true
	f.reflectResult = &ReflectFunc{
		fn:          funcRef,
		argCount:    fType.NumIn(),
		returnCount: fType.NumOut(),
		argValue:    reflect.ValueOf(funcRef),
	}

	// check if the count of arguments matches to the count of results
	if f.reflectResult.argCount != f.reflectFn.returnCount {
		return errors.New("number of arguments does not match the number of results")
	}

	iDidSomething := false
	for _, future := range f.futures {
		iDidSomething = true
		res := future.Await()
		switch args := res.(type) {
		case []interface{}:
			_, err := f.funcCall(funcRef, args)
			if err != nil {
				return err
			}
		default:
			return errors.New("invalid type of result")
		}

	}
	if iDidSomething {
		return nil
	}
	return errors.New("nothing to handle. did you forget to call Run()?")
}

func (f *FluxDevice) funcCall(fn interface{}, args []interface{}) ([]reflect.Value, error) {

	// fn should be a function and we we have to figure out
	// the arguments and the return values
	// and then we have to create a new function that will
	// be used as the function

	if fn != nil {

		var arguments []reflect.Value
		for _, arg := range args {
			arguments = append(arguments, reflect.ValueOf(arg))
		}
		returnValues := f.reflectResult.argValue.Call(arguments)
		return returnValues, nil

	} else {
		return nil, errors.New("argument must be a function. got nil")
	}

}
