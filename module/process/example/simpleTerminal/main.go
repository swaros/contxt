package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/swaros/contxt/module/process"
)

func main() {
	// create a new terminal process
	process := process.NewTerminal()

	// flag to keep the process running
	process.SetKeepRunning(true)

	// set the output handler so we can see the output
	// and handle errors
	process.SetOnOutput(func(msg string, err error) bool {
		if err != nil {
			fmt.Println("[Terminal] Error: ", err) // print the error
			return true                            // keep it running also if the command failed
		} else {
			fmt.Println("[Terminal] Output: ", msg) // print the output
			return true                             // keep the process running
		}
	})

	process.Exec()       // start the process
	defer process.Stop() // stop the process when we exit

	// the man loop til we enter "exit"
	for {
		// get the input from regular stdin
		// do not try to use the process stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Println(" <<<demo>>>enter command: (exit to stop) >>")
		text, terr := reader.ReadString('\n')
		fmt.Println()
		if terr != nil {
			fmt.Println(" <<<demo>>> Readline Error: ", terr)
			continue
		}
		text = text[:len(text)-1] // remove the newline

		if text == "exit" {
			break
		} else {
			process.Command(text)
		}
		// print the input
		fmt.Println(" <<<demo>>> command (" + text + ")")
	}
	fmt.Println(" <<<demo>>> Bye")
}
