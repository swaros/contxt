# YACLINT
yaclint is a simple linter for configurations loaded by using the yacl library.

## Example
```go
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

	// if we found any issues, then the issuelevel is not 0
	if linter.GetHighestIssueLevel() > 0 {
		// just print the issues. right now this should not be the case
		fmt.Println(linter.PrintIssues())
	}

	// now we can use the config
	fmt.Println(config.Name)
	fmt.Println(config.Age)

}
```