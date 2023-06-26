package main

import (
	"fmt"

	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yamc"
)

type Config struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
}

func main() {
	// create a new yacl instance
	config := Config{}
	cfgApp := yacl.New(
		&config,
		yamc.NewYamlReader(),
	)
	// load the config file. must be done before the linter can be used
	// in this case, the config is loaded from the current directory
	if err := cfgApp.LoadFile("config.yaml"); err != nil {
		panic(err)
	}
	// now any entry in the config file is mapped to the config struct
	if config.Age < 6 {
		panic("age must be greater than 5")
	}
	fmt.Println(" hello ", config.Name, ", age of  ", config.Age, " is correct?")
}
