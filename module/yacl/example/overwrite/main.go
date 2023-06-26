package main

import (
	"fmt"

	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yamc"
)

type Config struct {
	AuthUrl  string `yaml:"authurl"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
}

func main() {
	// create a new yacl instance
	config := Config{}
	cfgApp := yacl.New(
		&config,
		yamc.NewYamlReader(),
	)

	// define the subdirectory where the config files are located
	cfgApp.SetSubDirs("config")

	// load all config files from the config directory
	if err := cfgApp.Load(); err != nil {
		panic(err)
	}

	// just print the overwritten values
	fmt.Println(
		" connecting: ", config.AuthUrl,
		", user: ", config.UserName,
		", password: ", config.Password,
	)

}
