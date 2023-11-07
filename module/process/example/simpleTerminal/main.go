// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

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
		// on windows we need to remove the \r
		if len(text) > 0 && text[len(text)-1] == '\r' {
			text = text[:len(text)-1]
		}

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
