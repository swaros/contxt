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
	"fmt"
	"math/rand"
	"time"

	"log"

	"github.com/chzyer/readline"
)

func main() {
	completer := readline.NewPrefixCompleter(
		readline.PcItem("run"),
		readline.PcItem("help"),
		readline.PcItem("exit"),
	)
	rl, err := readline.NewEx(&readline.Config{
		UniqueEditLine: true,
		AutoComplete:   completer,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	rl.SetPrompt("username: ")
	username, err := rl.Readline()
	if err != nil {
		return
	}
	rl.ResetHistory()
	log.SetOutput(rl.Stderr())

	fmt.Fprintln(rl, "Hi,", username+"! My name is Dave.")
	rl.SetPrompt(username + "> ")

	done := make(chan struct{})
	go func() {
		rand.Seed(time.Now().Unix())
	loop:
		for {
			select {
			case <-time.After(time.Duration(rand.Intn(20)) * 100 * time.Millisecond):
			case <-done:
				break loop
			}
			log.Println("Dave:", "hello")
		}
		log.Println("Dave:", "bye")
		done <- struct{}{}
	}()

	for {
		ln := rl.Line()
		if ln.CanContinue() {
			continue
		} else if ln.CanBreak() {
			break
		}
		log.Println(username+":", ln.Line)
	}
	rl.Clean()
	done <- struct{}{}
	<-done
}
