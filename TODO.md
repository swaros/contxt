### Todo List

---

#### contxt <some-param>
parsing param. if they are integer use it like 'conxtxt dir -i <param>'. if they are a word look for the matching parts in stored paths to get the dir

##### IMPLEMENTATION
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

mirror paths in /home/<user>/.contxt/paths/ and look for .contxt.yml first. make an optional merge (configurable in .contxt.yml). add option -mirror to copy .contxt.yml into mirror path 
   
##### IMPLEMENTATION
 not yet
 
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
 not yet
 
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

 implemented as described. also variable values can use existing placeholders

---
