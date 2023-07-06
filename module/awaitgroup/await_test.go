package awaitgroup_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/swaros/contxt/module/awaitgroup"
)

func TestSingleRun(t *testing.T) {
	checkVal := 1

	future := awaitgroup.ExecFuture(nil, func() interface{} {
		time.Sleep(time.Millisecond * 500)
		checkVal++
		return 1
	})

	if checkVal != 1 {
		t.Error("checkval should be 1")
	}

	val := future.Await()
	if val != 1 {
		t.Error("val is not 1")
	}

	if checkVal != 2 {
		t.Error("checkval should be 2")
	}
}

func TestRunGroup(t *testing.T) {
	var tasks []awaitgroup.FutureStack

	tasks = append(tasks, awaitgroup.FutureStack{
		AwaitFunc: func(ctx context.Context) interface{} {
			return 2
		},
		Argument: nil,
	})

	tasks = append(tasks, awaitgroup.FutureStack{
		AwaitFunc: func(ctx context.Context) interface{} {
			return 60
		},
		Argument: nil,
	})

	tasks = append(tasks, awaitgroup.FutureStack{
		AwaitFunc: func(ctx context.Context) interface{} {
			return 100
		},
		Argument: nil,
	})

	tasks = append(tasks, awaitgroup.FutureStack{
		AwaitFunc: func(ctx context.Context) interface{} {
			return 2000
		},
		Argument: nil,
	})

	futures := awaitgroup.ExecFutureGroup(tasks)

	results := awaitgroup.WaitAtGroup(futures)

	sum := 0
	for _, v := range results {
		sum = sum + v.(int)
	}
	if sum != 2162 {
		t.Error("unexpected result:", sum)
	}

}

type Argument struct {
	WelcomeMsg string
	NumberProp int
	t          *testing.T
}

type Result struct {
	WelcomeMsg string
	NumberProp int
}

func TestArgumentUsage(t *testing.T) {
	var tasks []awaitgroup.FutureStack

	for i := 0; i < 5; i++ {
		tasks = append(tasks, awaitgroup.FutureStack{
			AwaitFunc: func(ctx context.Context) interface{} {
				// get the argument from contxt
				argument := ctx.Value(awaitgroup.CtxKey{}).(Argument)
				return Result{
					WelcomeMsg: argument.WelcomeMsg + fmt.Sprintf("[%d]", argument.NumberProp),
					NumberProp: argument.NumberProp,
				}
			},
			Argument: Argument{
				WelcomeMsg: "Hello..." + fmt.Sprintf("%d", i),
				NumberProp: i,
				t:          t,
			},
		})
	}

	futures := awaitgroup.ExecFutureGroup(tasks)
	results := awaitgroup.WaitAtGroup(futures)

	expectedLen := 5

	if len(results) != expectedLen {
		t.Error("unexpected result:", len(results))
		t.SkipNow()
	}

	expectedSlice := []string{
		"Hello...0[0]",
		"Hello...1[1]",
		"Hello...2[2]",
		"Hello...3[3]",
		"Hello...4[4]",
	}

	for _, v := range results {
		result := v.(Result)
		t.Log(result.WelcomeMsg)
		if result.WelcomeMsg != expectedSlice[result.NumberProp] {
			t.Error("unexpected result:", result.WelcomeMsg)
		}
	}

}

func TestArgumentUsageShortCuts(t *testing.T) {
	var tasks []awaitgroup.FutureStack

	for i := 0; i < 5; i++ {
		tasks = append(tasks, awaitgroup.FutureStack{
			AwaitFunc: func(ctx context.Context) interface{} {
				// get the argument from contxt
				argument := awaitgroup.GetArgument(ctx).(Argument)
				return Result{
					WelcomeMsg: argument.WelcomeMsg + fmt.Sprintf("[%d]", argument.NumberProp),
					NumberProp: argument.NumberProp,
				}
			},
			Argument: Argument{
				WelcomeMsg: "Hello..." + fmt.Sprintf("%d", i),
				NumberProp: i,
				t:          t,
			},
		})
	}

	futures := awaitgroup.ExecFutureGroup(tasks)
	results := awaitgroup.WaitAtGroup(futures)

	expectedLen := 5

	if len(results) != expectedLen {
		t.Error("unexpected result:", len(results))
		t.SkipNow()
	}

	expectedSlice := []string{
		"Hello...0[0]",
		"Hello...1[1]",
		"Hello...2[2]",
		"Hello...3[3]",
		"Hello...4[4]",
	}

	for _, v := range results {
		result := v.(Result)
		t.Log(result.WelcomeMsg)
		if result.WelcomeMsg != expectedSlice[result.NumberProp] {
			t.Error("unexpected result:", result.WelcomeMsg)
		}
	}

}
