package main

import (
	"fmt"

	"github.com/swaros/contxt/module/yamc"
)

func main() {
	data := []byte(`{"age": 45, "hobbies": ["golf", "reading", "swimming"]}`)
	conv := yamc.New()
	if err := conv.Parse(yamc.NewJsonReader(), data); err != nil {
		panic(err)
	} else {
		if str, err2 := conv.ToString(yamc.NewYamlReader()); err2 != nil {
			panic(err2)
		} else {
			fmt.Println(str)
		}
	}
}
