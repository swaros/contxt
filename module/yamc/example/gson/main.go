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

	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/yamc"
)

// simple usage of the yamc moduletogehter with the gjson module
// we ignore any error handling here just for easy reading
func main() {
	source := []byte(`{
	"name": {"first": "Tom", "last": "Anderson"},
	"age":37,
	"children": ["Sara","Alex","Jack"],
	"fav.movie": "Deer Hunter",
	"friends": [
	  {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
	  {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
	  {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
	]
  }`)

	data := yamc.New()
	data.Parse(yamc.NewJsonReader(), source)

	paths := []string{
		"name.last",
		"age",
		"children",
		"children.#",
		"children.1",
		"child*.2",
		"c?ildren.0",
		"fav\\.movie",
		"friends.#.first",
		"friends.1.last",
		"friends.#(last==\"Murphy\").first",
		"friends.#(last==\"Murphy\")#.first",
		"friends.#(age>45)#.last",
		"friends.#(first%\"D*\").last",
		"friends.#(first!%\"D*\").last",
		"friends.#(nets.#(==\"fb\"))#.first",
	}
	// just use any of the examples from gson README.md (https://github.com/tidwall/gjson)
	for _, gsonPath := range paths {
		name, _ := data.GetGjsonString(gsonPath)
		pforOut := systools.PadStringToR(gsonPath, 40) // just to format the output a bit. not needed for the example
		fmt.Println(pforOut, " => ", name)
	}

}
