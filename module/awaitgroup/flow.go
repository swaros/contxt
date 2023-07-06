package awaitgroup

import (
	"context"

	"github.com/google/uuid"
)

func GetArgument(ctx context.Context) interface{} {
	return ctx.Value(CtxKey{})
}

type Flow struct {
	function func(args ...interface{}) []interface{} // the logic that should be executed
	handler  func(args ...interface{})               // the logic that handles the return values
	args     map[string]ArgContext                   // the arguments they are passed to the function
	argOrder []string                                // the order of the arguments and also the function execution order and return order
}

type ArgContext struct {
	Arguments []interface{}
	id        string
}

type ReturnContexr struct {
	ReturnValue []interface{}
	id          string
}

func NewFlow() *Flow {
	return &Flow{
		args: make(map[string]ArgContext),
	}
}

func (f *Flow) Func(fn func(args ...interface{}) []interface{}) *Flow {
	f.function = fn
	return f
}

func (f *Flow) Handler(fn func(args ...interface{})) *Flow {
	f.handler = fn
	return f
}

func (f *Flow) Each(arg ...interface{}) *Flow {
	// create and return a new uuid
	// that will be used as a key
	uuid := uuid.New().String()
	f.args[uuid] = ArgContext{
		Arguments: arg,
		id:        uuid,
	}
	f.argOrder = append(f.argOrder, uuid)
	return f
}

func (f *Flow) Run() {
	var allTasks []FutureStack
	for _, arg := range f.argOrder {
		allTasks = append(allTasks, FutureStack{
			AwaitFunc: func(ctx context.Context) interface{} {
				arg := GetArgument(ctx).(ArgContext)
				resturnFromFunc := f.function(arg.Arguments...)
				return ReturnContexr{
					ReturnValue: resturnFromFunc,
					id:          arg.id,
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
}
