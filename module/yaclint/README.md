# YACLINT
yaclint is a simple linter for configurations loaded by using the yacl library.

## Example
```go
package main

import (
    "fmt"
    "github.com/yacl/yaclint"
    "github.com/yacl/yacl/api"
)

type Config struct {
    Name string `yaml:"name"`
    Age  int    `yaml:"age"`
}

func main() {
    // create a new yacl instance
    config := Config{}
    cfgApp := yacl.yacl.New(
		&config,
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
        // just print the issues
        fmt.Println(linter.PrintIssues())
    }
}
```