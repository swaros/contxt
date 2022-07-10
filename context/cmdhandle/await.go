// based on https://hackernoon.com/asyncawait-in-golang-an-introductory-guide-ol1e34sg
package cmdhandle

import (
	"context"

	"github.com/sirupsen/logrus"
)

type CtxKey struct{}

// Future interface has the method signature for await
type Future interface {
	Await() interface{}
}

// FutureStack struct contains the AwaitFunc
// and argurments
//
// note: this mght not the best way to handle the argument
// delivery see: https://go.dev/blog/context#TOC_3.2.
type FutureStack struct {
	AwaitFunc func(ctx context.Context) interface{}
	Argument  interface{}
}

func (f FutureStack) Await() interface{} {
	ctx := context.Background()
	ctxUsed := context.WithValue(ctx, CtxKey{}, f.Argument)
	return f.AwaitFunc(ctxUsed)
}

// ExecFuture executes the async function and set the the argument
func ExecFuture(arg interface{}, f func() interface{}) Future {
	var result interface{}
	c := make(chan struct{})
	go func() {
		defer close(c)
		result = f()
	}()
	return FutureStack{
		Argument: arg,
		AwaitFunc: func(ctx context.Context) interface{} {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-c:
				return result
			}
		},
	}
}

// ExecFutureGroup executes an group of Futures and returns
// assotiated future handler
func ExecFutureGroup(fg []FutureStack) []Future {
	var futures []Future
	GetLogger().WithField("taskCount", len(fg)).Debug("Task added")
	for _, funcTr := range fg {
		future := ExecFuture(funcTr.Argument, funcTr.Await)
		futures = append(futures, future)
	}
	GetLogger().WithField("futureCount", len(futures)).Debug("futures created")
	return futures
}

// WaitAtGroup wait until all Futures are executes
func WaitAtGroup(futures []Future) []interface{} {
	var results []interface{}
	GetLogger().WithField("futureCount", len(futures)).Debug("waiting of futures being executed")
	for i, f := range futures {
		GetLogger().WithFields(logrus.Fields{"cur": i, "of": len(futures)}).Debug("wating of...")
		val := f.Await()
		GetLogger().WithFields(logrus.Fields{"cur": i, "val": val, "of": len(futures)}).Debug("await result ...")
		results = append(results, val)
	}
	return results
}
