# YACLINT
yaclint is a simple linter for configurations loaded by using the yacl library.

the concept is simple. you have a struct that represents your configuration. you load the configuration from a file and then you can use the linter to check if the configuration contains all the fields that are needed, or if there are unexpected fields.

it is then up to you, how critical the issues are. because of different
type and value checks (depending on different config file types), it is often not possible to check if the configuration is invalid, just because the struct expects an integer, but it is a string in the config file. so this is tracked, but not ment as an error.

> **NOTE:** invalid types and values, and any other source related issues are not checked by this linter. there is also no need for, because **this would already throw an error** while reading the config file.


#### Happy Path Example
simple example of how to use the linter and just make sure there is no Bigger issue. that means it might be not all fields are set, but there are no unexpected fields.


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

	// anything that seems totally wrong is an error
	if linter.HasError() {
		// just print the issues. right now this should not be the case
		fmt.Println(`Error! 
	Sorry!
		but it seems you did something wrong
		in the config file`)
		fmt.Println(linter.PrintIssues())
	} else {
		// now we can use the config
		fmt.Println(config.Name)
		fmt.Println(config.Age)
	}
}
```

### Usage

first to understand what is happens in the background, here is a short overview of the process.

1. a configuration is loaded and mapped to a struct (*using the yacl library*)
2. the linter now reloads the configuration wthout any struct mapping
3. the linter now checks what fields are missing and what fields are unexpected, because they are not defined in the struct

it would not help to panic out if there are any differences. a difference is expected. so the linter just reports the issues and it is up to the application to decide what to do.

but there are of course some types of issues that are more critical than others. so the linter has 3 different levels of issues.

#### information level
in the **best case**, there is no different between the struct and the map. so the linter will not find any issues.
most of the time, and depending on the config file type, there are some differences "by design". so while recreate the map, some values maybe converted to a different type.

because the linter don't know about any reader based issues, it will still reported at lower level, as information.

for example:
```go
type Config struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
}
```
related yaml
```yaml
name: "John Doe"
age: 42000000
```

is reported as issue in **info** level.
```bash
[-]ValuesNotMatching: level[2] @age (Age) vs @age (Age) ['42000000' != '4.2e+07']
[+]ValuesNotMatching: level[2] @age (Age) vs @age (Age) ['4.2e+07' != '42000000']
```


#### warning level
the **usual case** is, that there are some fields they are defined in the struct, but not in the config file. this is not an issue, because this is most of the time the expected behavior, and some default values are used. **this will be an warning** and the application could decide if this is an issue or not.

for example:
```go
type Config struct {
	Name    string `yaml:"name"`
	Contact struct {
		Email string `yaml:"email"`
		Phone string `yaml:"phone"`
	} `yaml:"contact"`
	LastName string `yaml:"lastname"`
	Age      int    `yaml:"age"`
}
```
related yaml
```yaml
name: john
lastname: doe
age: 60
contact:
#  phone: 123456789 # <---- this is missing
  email: jdoe@example.com
```

is reported as issue in **warning** level.
```bash
[+]MissingEntry: level[5] @phone (Contact.Phone)
```

#### error level
the other case is, there are fields in the config file, they are not defined in the struct. this is an issue, because the configuration is not used as expected.

```yaml
name: john
lastname: doe
age: 60
phone: 123456789 # <---- expected in contact section but not here
contact:
  email: jdoe@example.com
```

is reported as issue in **error** level.
```bash
[-]UnknownEntry: level[12] @phone (phone)
[+]MissingEntry: level[5] @phone (Contact.Phone)
```

and worst, the user may have misspelled a field name. this can be a time consuming issue to find, because a typo is not always obvious. 

```yaml
name: john
lastname: doe
age: 60
contact:
  phone: 123456789
  emil: jdoe@example.com
```

is reported as issue in **error** level.
```bash
[-]UnknownEntry: level[12] @    emil (    emil)
[+]MissingEntry: level[5] @email (Contact.Email)
```

none of these issues can be seen just by using the yacl library (or by using the yaml or json libs). so the linter is used to check if the configuration is used as expected.
**so the linter will find this issues and print it out.**

### implementation

#### yaclint.NewLinter()
the linter is created by using the yaclint.NewLinter() function. it takes a yacl instance as parameter. this is used to get the configuration and the reader.

```go
func NewLinter(yacl yacl.Yacl) *Linter
```

#### linter.Verify()
the linter is used by calling the Verify() function. this will reload the configuration and check for any issues.

```go
func (l *Linter) Verify() error
```

the error that might be returned is related to the yacl library. so if there is an issue with the reader, this will be reported as error.

the issues from the Verify itself are reported in different ways.

#### linter.HasError()
checks if any of the issues at least has an error level.

```go
func (l *Linter) HasError() bool
```

#### linter.HasWarning()
checks if any of the issues at least has an warning level.

```go
func (l *Linter) HasWarning() bool
```

#### linter.HasInfo()
checks if any of the issues at least has an info level.

```go
func (l *Linter) HasInfo() bool
```

#### linter.GetIssue()
the linter has a function to get all the issues equal or higher a specific level. this is ment to react on the issues programmatically.
 
```go
func (l *Linter) GetIssue(level int, reportFn func(token *MatchToken))
```

for example:
```go
cantIgnore := linter.HasError() // error can not (or should not) be ignored at all
linter.GetIssue(yaclint.IssueLevelWarn, func(token *yaclint.MatchToken) {	 
	switch token.KeyPath {
	case "Contact":
		fmt.Println("The contact section is required.")
		cantIgnore = true
	case "Contact.Email":
		fmt.Println("The email is required. this is used for the authentication.")
		cantIgnore = true
	case "Contact.Phone":
		fmt.Println("no phone?. okay, fine...we can ignore this")
	}

})
// getting out if we found an issue that we can not ignore
if cantIgnore {
	fmt.Println("sorry, can't ignore this issue.")
	exit(1)
}
```


#### linter.PrintIssues()
the linter has a function to print out the issues. this is used to print out **all** the issues if there are any.

```go
func (l *Linter) PrintIssues() string
```

this is ment for report the issue to the user. the output is a string conversation of the issues.

```bash
[-]UnknownEntry: level[12] @    emil (    emil)
[+]MissingEntry: level[5] @email (Contact.Email)
```

for each level, we have also a function to get these outputs as string slices depending on the level.
(see below)

#### linter.Errors()
the linter has a function to get all the issues with an error level. this is used to print out the issues if there are any.
these are string conversations of the issues. it is ment for report the issue to the user.


```go
func (l *Linter) Errors() []String
```

#### linter.Warnings()
the linter has a function to get all the issues with an warning level. this is used to print out the issues if there are any.
similar to the GetErrors() function, these are string conversations of the issues. it is ment for report the issue to the user.

```go
func (l *Linter) Warnings() []String
```

#### linter.Infos()
the linter has a function to get all the issues with an info level. this is used to print out the issues if there are any.
similar to the GetErrors() function, these are string conversations of the issues. it is ment for report the issue to the user.

```go
func (l *Linter) Infos() []String
```

