package awaitgroup

import (
	"errors"
	"reflect"
)

type FluxDevice struct {
	futures   []Future
	function  func(args ...interface{}) []interface{} // the logic that should be executed
	lastError error                                   // the last error that occured
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
		fType := reflect.TypeOf(fn)
		if fType.Kind() != reflect.Func {
			return nil, errors.New("argument one must be a function")
		}

		fVal := reflect.ValueOf(fn)
		if len(args) != fVal.Type().NumIn() {
			return nil, errors.New("number of arguments does not match")

		}

		for i := 0; i < fVal.Type().NumIn(); i++ {
			if reflect.TypeOf(args[i]) != fVal.Type().In(i) {
				return nil, errors.New("argument type does not match")
			}
		}

		var arguments []reflect.Value
		for _, arg := range args {
			arguments = append(arguments, reflect.ValueOf(arg))
		}
		returnValues := fVal.Call(arguments)
		return returnValues, nil

	} else {
		return nil, errors.New("argument must be a function. got nil")
	}

}
