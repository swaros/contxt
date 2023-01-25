package main

import (
	"log"

	"github.com/swaros/contxt/module/runner"
)

func main() {

	if err := runner.Init(); err != nil {
		log.Fatal(err)
	}
}
