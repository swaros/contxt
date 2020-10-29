### Todo List

*this is not a complete task list. it is more then a overview for things they are planned, done and comes in mind.
it is just better to have something to keep ideas tracked. and not somewhere on the lokal pc*

---

#### contxt <some-param>
  
parsing param. if they are integer use it like 'contxt dir -i param'. if they are a word look for the matching parts in stored paths to get the dir

##### IMPLEMENTATION

![task Status](https://img.shields.io/badge/status-done-green)
![task Status](https://img.shields.io/badge/implemented-v0.0.2-blue)

`contxt bla` will check the workspaces. if there a workspaces exists, then it will be the same as `contxt dir -w bla`
without the danger to create a new workspace by accident.
if the workspace does not exists, a error will shown that this can not be used as workspace, and also that the param can't be used differently
````shell
contxt bla
bla is not a workspace
unexpected command  bla
````
`contxt run bla` will check for the target **bla** and executes them if exists. it is similiar to `contxt run -target bla` 
but the short form dos not accept multiple targets. so `contxt run bla,bla2` will not work for nonof the targets, even if they exists booth.

---

#### mirror paths

mirror paths in /home/username/.contxt/paths/ and look for .contxt.yml first. make an optional merge (configurable in .contxt.yml). add option -mirror to copy .contxt.yml into mirror path 
   
##### IMPLEMENTATION

![task Status](https://img.shields.io/badge/status-open-red)


 
----

#### Requirements

requirements should be checked for task. requirements should check
 - Operation System level 1 (linux, windows, os) *even if its not supported yet*
 - Operation System level 2 (linux-fedora, linux-debian and so on)
 - Operation System versions number
 - username
 - variables set
 - variables values
 - file exists
 - file not exists
 - result of a script (just exist code)

requirements have to be defined for multiple usage.

##### IMPLEMENTATION

![task Status](https://img.shields.io/badge/status-open-red)
  
---

#### varibales for placeholder block

define a **variables** key where variables can be set. they can be later used
variables always a map of string:string

variables can be set at different places and they will overwrite others, if the same key
is used.

variables can be set like:
````yaml
config:
  variables:
     var1: check1
     var2: test
task:
  - id: task1
    variables:
      var1: check2
    script:
      - echo '${var1} ${var2}'
````

should print `check2 test`

base template 
````go
config:
  variables: []
````

##### IMPLEMENTATION

![task Status](https://img.shields.io/badge/status-done-green)
![task Status](https://img.shields.io/badge/implemented-v0.0.2-blue)

 implemented as described. also variable values can use existing placeholders

---

#### implementing text/template

implementing parsing golang template engine for yaml files.

##### IMPLEMENTATION

![task Status](https://img.shields.io/badge/status-done-green)
![task Status](https://img.shields.io/badge/implemented-v0.0.2-blue)

because a yaml file that includes placeholders like `{{ .setup.options.sequence }}` is not parsable a different config file was needed.
so a new file comes in that adds **.inc** as prefix to the used template filename and loads values from them.
this file have currently the following structure

````yaml
include:
    basedir: true
    folders:
        - "../subdir/"
        - "../../readthistoo/"
````

if basedir is set, the current directory, including subdirectories, will be parsed and all **.json, *.yaml, .yml** files will  be parsed
and merged.
folders can have different directories they will be parsed too.

any content will be merged, so it can happen that a entrie from file A can overwrite file B

a template file can use this template like so:

````yaml
config:    
    sequencially: {{ .setup.options.sequence }}
    coloroff: {{ .setup.options.color }}
    variables: 
        checkApi: {{ .apiVersion }}
        checkName: {{ .name }}
task:
  - id: script
    script:
      - echo 'hallo welt'
      - ls -ga
      {{ range $key, $value := .mapcheck }}
      - echo ' tag {{ $key }} value {{ $value }}'
      {{ end }}
````

this will only work if any of these placeholders is set. otherwise a parsing error will happen.
see *docs/test/case10*

---

#### implementing blacklist for text/template 

because all file from any subdirectory is parsed, we need to exclude files from them
so it would make sense to add a blacklist that contains a list of regex.
if one of them is matching, the file will be ignored.

````yaml
include:
    basedir: true
    folders:
      - "../subdir/"
      - "../../readthistoo/"
    blacklist:
      - "test[234].yml"
````

##### IMPLEMENTATION

![task Status](https://img.shields.io/badge/status-open-red)