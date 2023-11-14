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
