// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

 package main

import (
	"fmt"
	"strings"

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
	// first the fixed version
	demo("contact2.yaml")

	// now the broken version
	demo("contact.yaml")
}

func demo(file string) {
	// print headline to split the output for each demo

	fmt.Println(strings.Repeat("-", 80))
	fmt.Println("Demo for file: ", file)
	fmt.Println(strings.Repeat("-", 80))

	// usual yacl stuff
	config := &Config{}
	cfgApp := yacl.New(
		config,
		yamc.NewYamlReader(),
	)
	if err := cfgApp.LoadFile(file); err != nil {
		panic(err)
	}

	// so we need to use the linter to find this issue
	// now the linter
	linter := yaclint.NewLinter(*cfgApp)
	if err := linter.Verify(); err != nil {
		panic(err)
	}

	// do we have any issues?
	// yes...we do! because WE know already about the wrong email property
	// but we would also like to handle any other issues at warning level
	if linter.HasWarning() {
		fmt.Println("Issues found by using file: ", file)

		// so we loaded the contact.yaml file (the other demofile have no issues)
		/*
			name: john
			lastname: doe
			age: 60
			email: jdoe@example.com
		*/
		// as you can se, there is no contact section. so we expect an error
		// the email is there, but in the wrong section.
		// the config is did not complain about the missing contact section, because
		// there is no invalid type or something else critical that would trigger an error.

		// first, let us show on screen, what issues we have
		// this is just for demonstration. but also a lazy way to show the issues to the user
		fmt.Println(strings.Join(linter.Warnings(), "\n"))

		// the output is:
		/*
			[-]UnknownEntry: level[12] @email (email)
			[+]MissingEntry: level[5] @contact (Contact)
			[+]MissingEntry: level[5] @email (Contact.Email)
			[+]MissingEntry: level[5] @phone (Contact.Phone)
		*/

		// to explain the output:

		// have a look at the [-] and [+] signs in front of the issues label.
		// [-] means that the source of this issue is the data, that was read without using the struct.
		// by comparing the data with the struct, we found the property email, that is not defined in the struct.
		// this way we asume that this is an UnknownEntry. so this should not be there.
		// [+] means that the source of this issue is the struct. we found a property in the struct, that is not in the data.
		// in opposite to the UnknownEntry, this is also an issue, but the linter can not say if this is a real issue.

		// to explain the issue label, like UnknownEntry or MissingEntry.
		// this is the text description of the issue level(5 and 12).
		// anything between 5 and 9 is a warning. anything between 10 and higher is an error.

		// last we see the struct information. first the property name, followed by the Path.
		// the path is the full path to the property (parent.parent.prop ...and so on).
		// so if we have a struct like this:
		/*
			type Config struct {
				Contact struct {
					Email string `yaml:"email"`
					Phone string `yaml:"phone"`
				} `yaml:"contact"`
			}
		*/
		// then the path for the email property is: Contact.Email
		// but the property name depends on the tag info. `yaml:"email"`
		// so to compare entries, we need to know the property name and the path.

		// so now we know, that we have an issue with the email property.
		// but we also have an more critical issue with the contact section.

		//			[-]UnknownEntry: level[12] @email (email)

		// so, of course, we need to fix the contact section first.

		// but even then, we still may haveing issues with the other properties in the contact section.
		// thats missed.
		/*
			[+]MissingEntry: level[5] @contact (Contact)
			[+]MissingEntry: level[5] @email (Contact.Email)
			[+]MissingEntry: level[5] @phone (Contact.Phone)
		*/

		// this is also the reason, why we are dealing with warnings and not only errors.
		// just because it depends on the application if these are also critical or not.
		// for doing this, we handle any issues at warning level and decide case by case if we can ignore it or not.

		cantIgnore := false
		linter.GetIssue(yaclint.IssueLevelWarn, func(token *yaclint.MatchToken) {
			switch token.KeyPath {
			case "Contact":
				fmt.Println("The contact section is required.")
				cantIgnore = true
			case "Contact.Email":
				fmt.Println("The email is required. this is used for the authentication.")
				cantIgnore = true
			case "Contact.Phone":
				fmt.Println("do you have no phone?. okay, fine...we can ignore this")
			}

		})
		// getting out if we found an issue that we can not ignore
		if cantIgnore {
			fmt.Println("sorry, can't ignore this issue.")
			return
		}
	}

	fmt.Println("no issues found by using file: ", file)
	// proceed with the application
}
