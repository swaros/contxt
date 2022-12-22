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

**goal** load a existing configuration that is stored as json file named _config.json_

```go
package main

import (

    "github.com/swaros/contxt/module/yacl"
    "github.com/swaros/contxt/module/yamc"
)

type MyConfig struct {
    Name string `json: "name"`
}

func main() {
    var cfg MyConfig
    err := yacl.New(&cfg,yamc.NewJsonReader()).
       LoadFile("config.json")
    
    if err != nil {
        // handle the error
    }
}
```
