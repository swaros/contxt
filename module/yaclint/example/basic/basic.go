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

	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yaclint"
	"github.com/swaros/contxt/module/yamc"
)

type Config struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
}

func main() {
	// create a new yacl instance
	config := &Config{}
	cfgApp := yacl.New(
		config,
		yamc.NewYamlReader(),
	)
	// load the config file. must be done before the linter can be used
	if err := cfgApp.LoadFile("config.yaml"); err != nil {
		panic(err)
	}

	// create a new linter instance
	linter := yaclint.NewLinter(*cfgApp)
	// error if remapping is not possible. so no linting error
	if err := linter.Verify(); err != nil {
		panic(err)
	}

	// anything that seems totally wrong is an error
	if linter.HasError() {
		// just print the issues. right now this should not be the case
		fmt.Println(`Error! 
		Sorry!
			but it seems you did something wrong
			in the config file`)
		fmt.Println(linter.PrintIssues())
	} else if linter.HasWarning() {
		// print the issues. right now this should not be the case
		fmt.Println(`Warning!
		You should check the config file
		`)
		fmt.Println(linter.PrintIssues())
	} else {
		// now we can use the config
		fmt.Println("hello ", config.Name)
		fmt.Println("you are ", config.Age, " years old")

		// the output is:
		/*
			hello  john
			you are  60  years old
		*/
	}
	fmt.Println(linter.PrintIssues())
}
