// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Licensed under the MIT License
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// this solution is currently experimental.
// the current implemenation via taskwatcher works, but is not a nice implementation
// and is hard to keep track of the running concurrents. there is no real state we should wait.
// the second try taskrunner.go on the other hand is just overengeniered, and much more out of
// control, even it looks "overwatched".
// to get rid of all this "half-baked" solutions i remembered an articel where an await like
// replacement was descriped.
//
// https://hackernoon.com/asyncawait-in-golang-an-introductory-guide-ol1e34sg
//
// based on that we have now a group of futures they we can use to have at some point
// a clear wait state for all this tasks. and all of them based on channels. so no waitgroup is required.
//
// so this can replace any concurrent tasks. for this reason i decided to have this as experimental feature.
// there is not test that can check any side-effect. and yes we have side effects.

package awaitgroup

import (
	"context"
)

// CtxKey is just the global key for the arguments
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

// Await creates the context including arfgument, and blocks til
// any execution is done.
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
	for _, funcTr := range fg {
		future := ExecFuture(funcTr.Argument, funcTr.Await)
		futures = append(futures, future)
	}
	return futures
}

// WaitAtGroup wait until all Futures are executes
func WaitAtGroup(futures []Future) []interface{} {
	var results []interface{}
	for _, f := range futures {
		val := f.Await()
		results = append(results, val)
	}
	return results
}
