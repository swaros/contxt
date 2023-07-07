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

var (
	calcsA = []int{1, 2, 3, 4, 5, 6, 7, 8}
	calcsB = []int{9, 10, 11, 12, 13, 14, 15, 16}
)

// this function is called all the time
// it takes two numbers and returns the sum of it
// and the execution number
func callSomething(excutionNumber int, numA int, numB int) (int, int) {
	// wait random milliseconds beween 100 and 1000
	// to bring chaos into the flow
	fmt.Println(" <--> the function that calculates ...  No:", excutionNumber, "numA", numA, "numB", numB)
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000-100)+100))
	return numA + numB, excutionNumber
}

// this function is using a waitgroup to wait for all
func doneWithWaitGroup() {

	fmt.Println("start waitgroup")
	waitGroup := sync.WaitGroup{}
	for i := 0; i < len(calcsA); i++ {
		waitGroup.Add(1)
		go func(execNr int, numA int, numB int) {
			result, _ := callSomething(execNr, numA, numB)
			fmt.Println("No:", execNr, "result : ", result)
			waitGroup.Done()
		}(i, calcsA[i], calcsB[i])
	}
	fmt.Println("it runs concurrently. wait for all to be done by waitgroup.Wait()")
	waitGroup.Wait()
	fmt.Println("done")
}

// this function is using a channel to wait for all
func doneWithChannel() {
	type cResult struct {
		execNr int
		result int
	}

	fmt.Println("channel start")
	channel := make(chan cResult)
	for i := 0; i < len(calcsA); i++ {
		go func(en int, numA int, numB int) {
			result, _ := callSomething(en, numA, numB)
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

// this function is using a flow to wait for all
func doneWithFlow() {
	flow := awaitgroup.NewFlow()
	flow.Use(callSomething)

	for i := 0; i < len(calcsA); i++ {
		flow.Go(i, calcsA[i], calcsB[i])
	}
	flow.Handler(func(args ...interface{}) {
		if len(args) != 2 {
			fmt.Println("wrong number of arguments")
			return
		}
		fmt.Println("No : ", args[1], "result : ", args[0])
	})
	fmt.Println("flow start")
	flow.Run()
	fmt.Println("flow done")

}
