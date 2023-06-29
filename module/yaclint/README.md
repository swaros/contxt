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
the other case is, that there are fields in the config file, they are not defined in the struct. this is an issue, because the configuration is not used as expected.

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


