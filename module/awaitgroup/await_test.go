package awaitgroup_test

import (
	"context"
	"testing"
	"time"

	"github.com/swaros/contxt/awaitgroup"
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
