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

	"github.com/swaros/contxt/module/yamc"
)

// this shows how to use the yamc module to read data from a json source
// independent from the source structure.
// the source can be a map[string]interface{} or a []interface{}
// it shows how to get the data from the source by using a simple path
func main() {
	// the soruce data
	jsonSource := []byte(`{
		"users": [
			{
				"id": 1,
				"name": "John Doe",
				"email": "jdoe@somewhere.org",
				"phone": "555-555-5555"
			},
			{
				"id": 2,
				"name": "Jane Doe",
				"email": "jdoe@somewhere.org",
				"phone": "555-555-5521"
			},
			{
				"id": 3,
				"name": "John Smith",
				"email": "jsm@prirate.org",
				"phone": "888-555-5555"
			}
		]
	}`)
	// create a new yamc instance
	data := yamc.New()
	// parse the data by using the json reader
	if err := data.Parse(yamc.NewJsonReader(), jsonSource); err != nil {
		panic(err)
	}

	// get the third users name.
	// the path is "users.2.name" what means keyname.index.keyname
	if user, err := data.FindValue("users.2.name"); err != nil {
		panic(err)
	} else {
		out := fmt.Sprintf("user: %v", user)
		fmt.Println(out)
	}

	// get the third users email
	if user, err := data.FindValue("users.2.email"); err != nil {
		panic(err)
	} else {
		out := fmt.Sprintf("email: %v", user)
		fmt.Println(out)
	}

	// using the data as map[string]interface{}
	if user, err := data.FindValue("users.2"); err != nil {
		panic(err)
	} else {
		out := fmt.Sprintf("user: %v", user)
		fmt.Println(out)
	}

	// using the gson path to get the whole users array
	// gson is a
	// https://github.com/tidwall/gjson
	if user, err := data.GetGjsonString("users.#.name"); err != nil {
		panic(err)
	} else {
		out := fmt.Sprintf("[gson] users.#.name: %v", user)
		fmt.Println(out)
	}

	// ---------- now the same as []interface{} ------------
	fmt.Println("\n------------- case 2 ------------")
	jsonSource = []byte(`[
		{
			"id": 1,
			"name": "John Doe",
			"email": "sfjhg@sdfkfg.net",
			"phone": "555-555-5555"
		},
		{
			"id": 2,
			"name": "Jane Doe",
			"email": "lllll@jjjjj.net",
			"phone": "555-555-5521"
		},
		{
			"id": 3,
			"name": "John Smith",
			"email": "kjsahd@fjjgg.net",
			"phone": "888-555-5555"
		}
	]`)
	// parse the data by using the json reader
	if err := data.Parse(yamc.NewJsonReader(), jsonSource); err != nil {
		panic(err)
	}

	// get the third users name in an array
	// here we have to use the index as string followoed by the keyname
	// so "2.name" means index.keyname
	if user, err := data.FindValue("2.name"); err != nil {
		panic(err)
	} else {
		out := fmt.Sprintf("user: %v", user)
		fmt.Println(out)
	}

	// get the third users email
	if user, err := data.FindValue("2.email"); err != nil {
		panic(err)
	} else {
		out := fmt.Sprintf("email: %v", user)
		fmt.Println(out)
	}

	// get the whole user
	if user, err := data.FindValue("2"); err != nil {
		panic(err)
	} else {
		out := fmt.Sprintf("user: %v", user)
		fmt.Println(out)
	}

}
