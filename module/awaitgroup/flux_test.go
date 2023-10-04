package awaitgroup_test

import (
	"fmt"
	"testing"

	"github.com/swaros/contxt/module/awaitgroup"
)

func TestFluxEngine(t *testing.T) {
	fluxCompensator := &awaitgroup.FluxDevice{}
	fluxCompensator.Use(func() interface{} {
		return 1
	})

	fluxCompensator.Run()

	fluxCompensator.ResultsTo(func(result interface{}) {
		if result != 1 {
			t.Error("result is not expected")
		}
	})

}

func testSomething(calcIn int) int {
	fmt.Println(" handle number ", calcIn)
	return calcIn + 100
}

var results []int

func handleResult(result int) {
	results = append(results, result)
	fmt.Println(" handle result", result)
}

func TestFluxEngineWithArgs(t *testing.T) {
	fluxCompensator := &awaitgroup.FluxDevice{}
	fluxCompensator.Use(testSomething)

	fluxCompensator.Run(1)
	fluxCompensator.Run(2)
	fluxCompensator.Run(3)

	if err := fluxCompensator.ResultsTo(handleResult); err != nil {
		t.Error(err)
	}

	expected := []int{101, 102, 103}
	for i, result := range results {
		if result != expected[i] {
			t.Error("result is not expected")
		}
	}

}

// testing flux together with function in a struct

type testStruct struct {
	datas []int
}

func (t *testStruct) handleResult(result int) {
	t.datas = append(t.datas, result)
	fmt.Println(" handle result", result)
}

func (t *testStruct) testSomething(calcIn int) int {
	fmt.Println(" handle number ", calcIn)
	return calcIn + 100
}

func TestFluxEngineWithStruct(t *testing.T) {
	fluxCompensator := &awaitgroup.FluxDevice{}
	test := &testStruct{}
	fluxCompensator.Use(test.testSomething)

	fluxCompensator.Run(1)
	fluxCompensator.Run(2)
	fluxCompensator.Run(3)

	if err := fluxCompensator.ResultsTo(test.handleResult); err != nil {
		t.Error(err)
	}

	expected := []int{101, 102, 103}
	for i, result := range test.datas {
		if result != expected[i] {
			t.Error("result is not expected")
		}
	}

}

// now again, testing flux together with function in a struct and more complex
// datas, arguments and results

type testStruct2Result struct {
	result  int
	message string
}

type testStruct2 struct {
	datas []testStruct2Result
}

func (t *testStruct2) handleResult(result int, logMessage string) {
	t.datas = append(t.datas, testStruct2Result{result, logMessage})
	fmt.Println(" handle result", result)
}

func (t *testStruct2) testSomething(calcIn int) (int, string) {
	fmt.Println(" handle number ", calcIn)
	return calcIn + 100, fmt.Sprintf("result is %d", calcIn)
}

func TestFluxEngineWithStruct2(t *testing.T) {
	fluxCompensator := &awaitgroup.FluxDevice{}
	test := &testStruct2{}
	fluxCompensator.Use(test.testSomething)

	fluxCompensator.Run(1)
	fluxCompensator.Run(2)
	fluxCompensator.Run(3)

	if err := fluxCompensator.ResultsTo(test.handleResult); err != nil {
		t.Error(err)
	}

	expected := []testStruct2Result{
		{101, "result is 1"},
		{102, "result is 2"},
		{103, "result is 3"},
	}
	for i, result := range test.datas {
		if result != expected[i] {
			t.Error("result is not expected")
		}
	}

}

// same as above, but this time with an need to pass multiple arguments to the
// function

type testStruct3Result struct {
	result  int
	message string
}

type testStruct3 struct {
	datas []testStruct3Result
}

func (t *testStruct3) handleResult(result int, logMessage string) {
	t.datas = append(t.datas, testStruct3Result{result, logMessage})
	fmt.Println(" handle result", result)
}

func (t *testStruct3) testSomething(calcIn int, logMessage string) (int, string) {
	fmt.Println(" handle number ", calcIn)
	return calcIn + 100, fmt.Sprintf(" result is %d. message is [%s]", calcIn, logMessage)
}

func TestFluxEngineWithStruct3(t *testing.T) {
	fluxCompensator := &awaitgroup.FluxDevice{}
	test := &testStruct3{}
	fluxCompensator.Use(test.testSomething)

	fluxCompensator.Run(1, "hello")
	fluxCompensator.Run(2, "world")
	fluxCompensator.Run(3, "again")

	if err := fluxCompensator.ResultsTo(test.handleResult); err != nil {
		t.Error(err)
	}

	expected := []testStruct3Result{
		{101, " result is 1. message is [hello]"},
		{102, " result is 2. message is [world]"},
		{103, " result is 3. message is [again]"},
	}
	for i, result := range test.datas {
		if result != expected[i] {
			t.Error("result is not expected")
		}
	}

}

func TestNewFlux(t *testing.T) {
	test := &testStruct3{}
	flux, err := awaitgroup.NewFluxDevice(test.testSomething)
	if err != nil {
		t.Error(err)
	}

	flux.Run(1, "hello")
	flux.Run(2, "world")
	flux.Run(3, "again")

	if err := flux.ResultsTo(test.handleResult); err != nil {
		t.Error(err)
	}

	expected := []testStruct3Result{
		{101, " result is 1. message is [hello]"},
		{102, " result is 2. message is [world]"},
		{103, " result is 3. message is [again]"},
	}

	for i, result := range test.datas {
		if result != expected[i] {
			t.Error("result is not expected")
		}
	}
}
