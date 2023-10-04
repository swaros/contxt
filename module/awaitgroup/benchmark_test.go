package awaitgroup_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/swaros/contxt/module/awaitgroup"
)

// benchmarking the flux engine and the flow engine

type test4Bench struct {
	datas []testStruct3Result
}

func factorial(n int) int {
	for i := 100 - 1; i > 0; i-- {
		n *= i - 1
	}
	return n
}

func (t *test4Bench) handleResult(result int, logMessage string) {
	t.datas = append(t.datas, testStruct3Result{result, logMessage})
}

func (t *test4Bench) testSomething(calcIn int, logMessage string) (int, string) {
	calcIn = factorial(calcIn)

	return calcIn, fmt.Sprintf(" result is %d. message is [%s]", calcIn, logMessage)
}

func BenchmarkFluxEngine(b *testing.B) {
	fluxCompensator := &awaitgroup.FluxDevice{}
	testBench := &test4Bench{}
	fluxCompensator.Use(testBench.testSomething)

	for i := 0; i < b.N; i++ {
		fluxCompensator.Run(i, "hello")
	}

	if err := fluxCompensator.ResultsTo(testBench.handleResult); err != nil {
		b.Error(err)
	}
}

func BenchmarkFluxEngineNoReflection(b *testing.B) {
	fluxCompensator := &awaitgroup.FluxDevice{}
	testBench := &test4Bench{}
	fluxCompensator.Fn(func(args ...interface{}) []interface{} {
		calcIn := args[0].(int)
		logMessage := args[1].(string)
		res1, res2 := testBench.testSomething(calcIn, logMessage)
		return []interface{}{res1, res2}
	})

	for i := 0; i < b.N; i++ {
		fluxCompensator.Run(i, "hello")
	}

	results := fluxCompensator.Results()
	for _, result := range results {
		result := result.([]interface{})
		testBench.handleResult(result[0].(int), result[1].(string))
	}
}

func BenchmarkNoAsync(b *testing.B) {
	testBench := &test4Bench{}
	for i := 0; i < b.N; i++ {
		calced, msg := testBench.testSomething(i, "hello")
		testBench.handleResult(calced, msg)
	}
}

func BenchmarkWithNativeGo(b *testing.B) {
	testBench := &test4Bench{}
	wg := sync.WaitGroup{}
	results := make(chan testStruct3Result, b.N)
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(i int) {
			calced, msg := testBench.testSomething(i, "hello")
			testBench.handleResult(calced, msg)
			results <- testStruct3Result{calced, msg}
			wg.Done()
		}(i)
	}
	wg.Wait()
	close(results)
	for result := range results {
		testBench.handleResult(result.result, result.message)
	}
}

func BenchmarkAwaitgroup(b *testing.B) {

	testBench := &test4Bench{}
	var tasks []awaitgroup.FutureStack
	for i := 0; i < b.N; i++ {
		tasks = append(tasks, awaitgroup.FutureStack{
			AwaitFunc: func(ctx context.Context) interface{} {
				calced, msg := testBench.testSomething(i, "hello")
				testBench.handleResult(calced, msg)
				return nil
			},
			Argument: nil,
		})
	}

	futures := awaitgroup.ExecFutureGroup(tasks)

	awaitgroup.WaitAtGroup(futures)

}
