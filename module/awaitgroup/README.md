# Await-Group

[![Go Reference](https://pkg.go.dev/badge/github.com/swaros/contxt/module/awaitgroup.svg)](https://pkg.go.dev/github.com/swaros/contxt/module/awaitgroup)
this package is part of the [contxt mono-repo](https://github.com/swaros/contxt).

## Usage

install package

`go get -u github.com/swaros/contxt/module/awaitgroup`

here a example implementation

````go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/swaros/contxt/module/awaitgroup"
)

func main() {
	// create a simple task slice
	var tasksLayer1 []awaitgroup.FutureStack // first group of tasks
	var tasksLayer2 []awaitgroup.FutureStack // second group of tasks

	// adding a task to the first lask list. we call them layer1
	tasksLayer1 = append(tasksLayer1, awaitgroup.FutureStack{
		AwaitFunc: func(ctx context.Context) interface{} {
			fmt.Println("L1-1 :: i am task No. 1 ...and i run in the fist layer")
			fakeDoing(5, "L1-1 :: hard working")
			fmt.Println("L1-1 :: done doing something")
			return 2
		},
		Argument: nil,
	})
	// a second taks in the first task group named layer 1
	tasksLayer1 = append(tasksLayer1, awaitgroup.FutureStack{
		AwaitFunc: func(ctx context.Context) interface{} {
			fmt.Println("L1-2 :: i am task No. 2 ...and i run in the fist layer too")
			fakeDoing(5, "L1-2 :: counting stars")
			fmt.Println("L1-2 :: done doing something")
			return 50
		},
		Argument: nil,
	})

	// here we add the first task to the second task list
	tasksLayer2 = append(tasksLayer2, awaitgroup.FutureStack{
		AwaitFunc: func(ctx context.Context) interface{} {
			fmt.Println("L2-1 :: i am in layer 2 ... ")
			fakeDoing(15, "L2-1 :: way more to do")
			return "check"
		},
		Argument: nil,
	})
	// and another task to the second task list. this task will be done quit fast
	tasksLayer2 = append(tasksLayer2, awaitgroup.FutureStack{
		AwaitFunc: func(ctx context.Context) interface{} {
			fmt.Println("L2-2 :: i am in layer 2 ... ")
			fakeDoing(2, "L2-2 :: the lazy ones")
			fmt.Println("L2-2 :: already done")
			return "lazy-is-first"
		},
		Argument: nil,
	})
	// starts all task lists and recive the future handler
	futuresL1 := awaitgroup.ExecFutureGroup(tasksLayer1)
	futuresL2 := awaitgroup.ExecFutureGroup(tasksLayer2)

	fmt.Println("starting...")
	// til now we wait until all tasks from layer1 (tasklist 1) are done
	results := awaitgroup.WaitAtGroup(futuresL1)
	fmt.Println(" =====  so layer 1 IS DONE ...wait for the second group ==== ")
	// working with the results of the tasklist
	sum := 0
	for _, v := range results {
		sum = sum + v.(int)
	}
	// now we wait for the second tasklist
	results2 := awaitgroup.WaitAtGroup(futuresL2)
	fmt.Println(" =====  so layer 2 is also DONE ... thats it ==== ")

	fmt.Println("reported from all layer 1 results. sum of layer 1:", sum, " data from layer 2:", results2)
}

func fakeDoing(times int, message string) {
	for i := 0; i < times; i++ {
		fmt.Println("\t", message, " -- ", i, "/", times)
		time.Sleep(1 * time.Second)
	}
}
````