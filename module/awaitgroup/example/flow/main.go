package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/swaros/contxt/module/awaitgroup"
)

func main() {
	doneWithWaitGroup() // running a group of functions with a waitgroup
	doneWithChannel()   // running a group of functions with a channel
	doneWithFlow()      // running a group of functions with a flow
}

// this function is called all the time
func callSomething(excutionNumber int, numA int, numB int) (int, error) {
	// wait random milliseconds beween 100 and 1000
	// to bring chaos into the flow
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000-100)+100))
	return numA + numB, nil
}

func doneWithWaitGroup() {
	calcsA := []int{1, 2, 3, 4, 5, 6, 7, 8}
	calcsB := []int{9, 10, 11, 12, 13, 14, 15, 16}

	fmt.Println("start waitgroup")
	waitGroup := sync.WaitGroup{}
	for i := 0; i < len(calcsA); i++ {
		waitGroup.Add(1)
		go func(execNr int, numA int, numB int) {
			result, err := callSomething(execNr, numA, numB)
			if err != nil {
				panic(err)
			}
			fmt.Println("No:", execNr, "result : ", result)
			waitGroup.Done()
		}(i, calcsA[i], calcsB[i])
	}
	fmt.Println("it runs concurrently. wait for all to be done by waitgroup.Wait()")
	waitGroup.Wait()
	fmt.Println("done")
}

func doneWithChannel() {
	calcsA := []int{1, 2, 3, 4, 5, 6, 7, 8}
	calcsB := []int{9, 10, 11, 12, 13, 14, 15, 16}

	type cResult struct {
		execNr int
		result int
	}

	fmt.Println("channel start")
	channel := make(chan cResult)
	for i := 0; i < len(calcsA); i++ {
		go func(en int, numA int, numB int) {
			result, err := callSomething(en, numA, numB)
			if err != nil {
				panic(err)
			}
			channel <- cResult{en, result}
		}(i, calcsA[i], calcsB[i])
	}
	fmt.Println("it runs concurrently. wait for all to be done by <-channel")
	for i := 0; i < len(calcsA); i++ {
		resFromChannel := <-channel
		fmt.Println("No : ", resFromChannel.execNr, " result", resFromChannel.result)
	}
	fmt.Println("channel done")
}

func doneWithFlow() {
	calcsA := []int{1, 2, 3, 4, 5, 6, 7, 8}
	calcsB := []int{9, 10, 11, 12, 13, 14, 15, 16}

	flow := awaitgroup.NewFlow()
	flow.Func(func(args ...interface{}) []interface{} {
		execNr := args[0].(int)
		arg1 := args[1].(int)
		arg2 := args[2].(int)
		ret, err := callSomething(execNr, arg1, arg2)
		if err != nil {
			panic(err)
		}
		return []interface{}{execNr, ret, err}
	})

	for i := 0; i < len(calcsA); i++ {
		flow.Each(i, calcsA[i], calcsB[i])
	}
	flow.Handler(func(args ...interface{}) {
		fmt.Println("No : ", args[0], "result : ", args[1])
	})
	fmt.Println("flow start")
	flow.Run()
	fmt.Println("flow done")

}
