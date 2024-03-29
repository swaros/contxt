# script cheats
<!-- TOC -->

- [script cheats](#script-cheats)
    - [overview of cheats](#overview-of-cheats)
        - [import-json](#import-json)
        - [import-json-exec](#import-json-exec)
        - [var](#var)
        - [set](#set)
        - [set-in-map](#set-in-map)
        - [export-to-yaml](#export-to-yaml)
        - [export-to-json](#export-to-json)
        - [add](#add)
        - [if-equals](#if-equals)
        - [if-not-equals](#if-not-equals)
        - [if-os](#if-os)
        - [var-to-file](#var-to-file)

<!-- /TOC -->
script cheats are special lines in the script section, they are not submitted to the target command.
they calling some internal functions instead.


for example assigning a value to an variable.
````yaml
task:
  - id: script
    script:
      - "#@set MY-VAR hello world"
      - echo ${MY-VAR}
````

these cheats having also the possibility to manipulate the script section.

````yaml
task:
  - id: script
    script:
      - "#@if-os linux"
      - echo "hello linux user"
      - "#@end"
      - "#@if-os windows"
      - echo "hello windows user"
      - "#@end"
````

this will work at one level only. this cheats should not being a new
intepreter language. if you like doing more complex logic, then this is the wrong place.

so this example is **not** working.

````yaml
task:
  - id: script-not-working
    script:
      - "#@if-os linux"
      - "#@if-equals ${USER} root "
      - echo "oh ..hello root"
      - "#@end"
      - "#@end"
````
## overview of cheats

| aviable           |
|-------------------|
|#@import-json      |
|#@import-json-exec |
|#@var              |
|#@set              |
|#@set-in-map       |
|#@export-to-yaml   |
|#@export-to-json   |
|#@add              |
|#@if-equals        |
|#@if-not-equals    |
|#@if-os            |
|#@end              |
|#@var-to-file      |



### import-json

imports a json content what then can be used as variable.


**arguments**: `map-key-name` `json content`

example
````yaml
task:
  - id: example
    script:
      - "#@import-json CONFIG {\"setup\":{\"testmsg\" : \"hello world\"}}"
      - echo "${CONFIG:setup.testmsg}"

````


### import-json-exec

imports a json by an bash command, what then can be used as variable

**arguments**: `map-key-name` `command to execute`

we asume a file named `user-config.json` exists in the same path.
````json
{
  "setup":{
    "testmsg" : "hello world"
  }
}
````

example
````yaml
task:
  - id: example
    script:
      - "#@import-json CONFIG cat user-config.json"
      - echo "${CONFIG:setup.testmsg}"

````


### var

sets a simple variable from the result of an bash command

**arguments**: `key-name` `command to execute`

example
````yaml
task:
  - id: example
    script:
      - "#@var DATE date"
      - echo "${DATE}"

````

### set

sets a simple variable just by the argument

**arguments**: `key-name` `content of the key`

example
````yaml
task:
  - id: example
    script:
      - "#@set HELLO hello world"
      - echo "${HELLO}"

````

### set-in-map

re-set an value of a loaded key-value map.
> uses [sjson](https://github.com/tidwall/sjson) annotation

**arguments**: `map-key-name` `sjson annotation` `content of the key`

example
````yaml
task:
  - id: example
    script:
      - "#@import-json CONFIG cat user-config.json" # just to load something
      - echo "${CONFIG:setup.testmsg}" # origin
      - "#@set-in-map CONFIG setup.testmsg the new content"
      - echo "${CONFIG:setup.testmsg}" # should be 'the new content'

````

### export-to-yaml

export the content of an existing key-map as yaml in a variable
> uses [sjson](https://github.com/tidwall/sjson) annotation

**arguments**: `map-key-name` `variable-name`

example
````yaml
task:
  - id: example
    script:
      - "#@import-json CONFIG cat user-config.json" # just to load something
      - "#@export-to-yaml CONFIG CONFIG_YAML"
      - echo "${CONFIG_YAML}"

````

### export-to-json

export the content of an existing key-map as json in a variable
> uses [sjson](https://github.com/tidwall/sjson) annotation

**arguments**: `map-key-name` `variable-name`

example
````yaml
task:
  - id: example
    script:
      - "#@import-json CONFIG cat user-config.json" # just to load something
      - "#@export-to-json CONFIG CONFIG_YAML"
      - echo "${CONFIG_YAML}"

````

### add

adds a string at the end to an existing string.
this is the same result as by using **set**
(`#@set HELLO ${HELLO} plus this`)
but without raise conditions.

it works directly on the variable and blocks any other rewrite of them.

**arguments**: `key-name` `content of the key`

example
````yaml
task:
  - id: example
    script:
      - "#@set HELLO hello world"
      - echo "${HELLO}"

````

### if-equals
condition to ignore the script lines til we reach the next `#@end` annotation,
when the check fails.

````yaml
task:
  - id: script
    script:
      - "#@if-equals ${USER} root"
      - echo "root!? okay then lets do the danger things"
      - "#@end"
````

### if-not-equals
condition to ignore the script lines til we reach the next `#@end` annotation,
when the check fails.

````yaml
task:
  - id: script
    script:
      - "#@if-not-equals ${USER} root"
      - echo "you have no power!"
      - "#@end"
````

### if-os
condition to ignore the script lines til we reach the next `#@end` annotation,
when the check fails.

````yaml
task:
  - id: script
    script:
      - "#@if-os linux"
      - echo "hello linux user"
      - "#@end"
      - "#@if-os windows"
      - echo "hello windows user"
      - "#@end"
````
### var-to-file
writes a variable into a file. any placeholder will be parsed first.
this is usefull in combination of importing a text file into a
variable.

**arguments**: `variable` `filename`

simple info file
````md
# current version
current version is ${version}
````

````yaml
config:
  imports:
    - version.tpl V-INFO
  variables:
    version: "1.0.4"
task:
  - id: write-makefile
    script:
      - "#@var-to-file V-INFO version.md"
````

content of *version.md* would then be
````md
# current version
current version is 1.0.4
````
