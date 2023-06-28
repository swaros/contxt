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
	if err := cfgApp.LoadFile("contact.yaml"); err != nil {
		panic(err)
	}

	// so we loaded the contact.yaml file
	/*
		name: john
		lastname: doe
		age: 60
		email: jdoe@example.com
	*/
	// as you can se, there is no contact section. so we expect an error
	// the email is there, but in the wrong section.
	// the config is did not complain about the missing contact section, because
	// there is no invalid type or something that would trigger an error.

	// so we need to use the linter to find this issue
	// now the linter
	linter := yaclint.NewLinter(*cfgApp)
	if err := linter.Verify(); err != nil {
		panic(err)
	}

	// do we have any issues?
	// we do! beaucse WE know already about the wrong email property
	if linter.HasError() {
		// first, let us show what issues we have
		fmt.Println(linter.PrintIssues())

		// the output is:
		/*
			[-]ValuesNotMatching: level[2] @email ['jdoe@example.com' != '']
			[+]MissingEntry: level[10] @contact
			[+]ValuesNotMatching: level[2] @email ['' != 'jdoe@example.com']
			[+]MissingEntry: level[10] @phone
		*/

		// the most important issues have the highest level. anything above 9 is a real issue
		// we got 2 MissingEntry issues. one for the contact section and one for the phone property.
		// booth are level 10. so this is something that can lead to errors in the application.
		// so here we should tell the user that he needs to fix the configuration file.

	} else {
		fmt.Println("no issues found. (but you shold not see this message)")
	}

}
