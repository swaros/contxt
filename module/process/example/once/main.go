package main

import (
	"fmt"

	"github.com/swaros/contxt/module/process"
)

// run a command once and print the output
func main() {
	command := process.NewTerminal(`echo "Hello World"`)   // create a new terminal process
	command.SetOnOutput(func(msg string, err error) bool { // set the output handler so we can see the output
		fmt.Println(msg) // print the output (we ignore the error case)
		return true      // keep the process running
	})
	command.Exec() // start the process
}
