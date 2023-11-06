package main

import (
	"fmt"
	"time"
)

// just to start a process and keep it running
// we will print some stuff to stdout so we can see it

func main() {
	for {
		fmt.Println("Hello World:", time.Now().String())
		time.Sleep(5 * time.Second)
	}
}
