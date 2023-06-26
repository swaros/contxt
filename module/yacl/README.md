# YACL

Yet another config loader

## features

- overrides for multiple configurations (for example local, dev, deployment) by ordered names
- configurable basedir (relative, absolute, homedir, config dir) and sub dirs
- config migration support (you need to change the config structure?)
- supports yaml and json by default. also together.
- configurable behavior
  - parse a single config file
  - parse a couple files one after another (value override)
  - parse all files (they matching readers files extension) in the config folder and any subfolder
  - ... but exclude a list of folders ...
  - ... or exclude any foldername that s not matching a path-pattern...
  - parse all files in the config dir, but ignore subfolder
  - exclude files by regex match
  - case depending Init callback
    - config folder not exists
    - no config files
- simple usage like an json.Marshall

## Example

### simple example
**goal** load a existing configuration that is stored as _yaml_ file named _config.yaml_ in the same directory as the executable.


#### config file _config.yaml_:
```yaml
name: "bugs bunny"
age: 102
```

##### example code:

```go
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
    // and we can use them
	if config.Age < 6 {
		panic("age must be greater than 5")
	}
	fmt.Println(" hello ", config.Name, ", age of  ", config.Age, " is correct?")
}
```
### overwriting values

**goal** instead of loading the configuration from a single file, we want to load the configuration from a couple of files. the order of the files is important. the last file in the list will overwrite any value from the previous files.

#### main config file _01-main.yml_:

the first configuraion file is named _01-main.yml_ and is located config folder. the config folder is located in the same directory as the executable.

this file contains the main configuration. it is the base for all other configurations.

for oure example, we assume that if the user is guest, then the user has no password. if the user is not guest, then the user has a password.
so we do not set the password in the main config file.

```yaml
authurl: company.com/auth
username: guest
```

now we want to overwrite the username and the password. we create a second file named _02-local.yml_ in the same folder as the _01-main.yml_.

because it is in order after the _01-main.yml_, the values in this file will overwrite the values from the _01-main.yml_.

#### local config file _02-local.yml_:

```yaml
username: testuser
password: "ghhnj44582#%$^"
```

##### example code:

```go
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
```

### short example
 for any example, we assume that the config is created already like
````go
config := Config{}
cfgApp := yacl.New(
   &config,
   yamc.NewYamlReader(), // yamc.NewJsonReader() for json
)
````

#### how to use a absolute path for the config file

```go
cfgApp.SetFileAndPathsByFullFilePath("/etc/myapp/configuraion.json")
```

this is an shortcut for a couple of settings that ends up reading the file from the given path.

#### how to allow not existing config files

```go
cfgApp.SetExpectNoConfigFiles()
```
now the loader will not complain if the config file does not exists. it will just do nothing.

#### how to use the home directory as base for the config file

```go
cfgApp.UseHomeDir()
```
now the users home directory is used as base for the config file. the config file must be located in the home directory. if you like to have your config file in a subdirectory, you can use the SetSubDirs() function.

```go
cfgApp.UseHomeDir().SetSubDirs(".config","myapp")
```

#### how to exclude a folder from the config file search

##### blacklist
```go
cfgApp.SetSubDirs("data", "v2").
  SetFolderBlackList([]string{"data/v2/deploy", "data/v2/backup"})
```
here we use a blacklist to exclude the folders _data/v2/deploy_ and _data/v2/backup_ from the config file search.

##### regex

depends on lists, only blacklists supported. but we can use a regex to limit the folders that are searched for config files.
like if we want to exclude all folders that are not matching the pattern ` _data/v2/.*_.`

```go
cfgApp.SetSubDirs("data", "v2").
  AllowSubdirsByRegex("_data/v2/.*_.)
```






