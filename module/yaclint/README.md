# YACLINT
yaclint is a simple linter for configurations loaded by using the yacl library.

the concept is simple. you have a struct that represents your configuration. you load the configuration from a file and then you can use the linter to check if the configuration contains all the fields that are needed, or if there are unexpected fields.

it is then up to you, how critical the issues are. if a field is missing, then the issue level is 1. if a field is unexpected, then the issue level is 2. and so on.

> **NOTE:** invalid types and values, and any other source related issues are not checked by this linter. this is still done by the yacl library. (and there, of course, by the source reader). so an tab in a yaml file is an Error and not an Linter Issue.


## Example
simple example of how to use the linter and just make sure there is no lint issue at all. what means we already blame the user on issue level 0
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

### Limitations
focus is on simplicity and ease of use. it is not meant to be a full blown linter.
it is also limited to a couple of source formats they mostly used for configuration files.

it dos not support the following data structures (_right now_):
- `map[string]string`
- `[]whatever`
- `{}`
- `*Struct`

this linter is about checking if the configuration contains all the fields that are needed, or if there are unexpected fields.
and anything is done depending go code. so no external configuration files.
