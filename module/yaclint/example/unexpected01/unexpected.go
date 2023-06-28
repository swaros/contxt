package main

import (
	"fmt"

	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yaclint"
	"github.com/swaros/contxt/module/yamc"
)

type Config struct {
	Name    string `yaml:"name"`
	Contact struct {
		Email string `yaml:"email"`
		Phone string `yaml:"phone"`
	} `yaml:"contact"`
	LastName string `yaml:"lastname"`
	Age      int    `yaml:"age"`
}

func main() {
	// usual yacl stuff
	config := &Config{}
	cfgApp := yacl.New(
		config,
		yamc.NewYamlReader(),
	)
	if err := cfgApp.LoadFile("contact2.yaml"); err != nil {
		panic(err)
	}

	// now the linter
	linter := yaclint.NewLinter(*cfgApp)
	if err := linter.Verify(); err != nil {
		panic(err)
	}

	// do we have any issues?
	if linter.GetHighestIssueLevel() > 0 {
		// we can start with printing the diff as a string
		fmt.Println(linter.GetDiff())
		fmt.Println("\n\t-----------------------")

		// we can print the trace
		fmt.Println(linter.GetTrace())
		fmt.Println("\n\t-----------------------")

		// we can print the issues
		fmt.Println(linter.PrintIssues())
		fmt.Println("\n\t-----------------------")

		// we can look at any token
		linter.ReportDiffStartedAt(0, func(token *yaclint.MatchToken) {
			fmt.Println(token.ToString())
		})
	} else {
		fmt.Println("no issues found")
	}

}
