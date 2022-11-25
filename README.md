
# cont(e)xt

i created **contxt** as a tool for my self to keep track where i had stored releated content for different Projects, and also 
what specific techniques and configurations are required.

## installation
### Linux
check out the [releases](https://github.com/swaros/contxt/releases) to get a pre-build package.

#### shell integration

the issue with change the directory is: you can not do this. **not if you are an binaray running in a shell**. 

if you are done, you will back in the directory you started. 

to *fix* this issue, contxt will be mapped by a shell function called **ctx** (for contxt itself), and **cn** for changing directories depending the current project. 

![cn command](https://github.com/swaros/docu-asset-store/blob/main/bash_demo_01.gif)

this works because shell functions are in same context.
so for any supported shell a shell function is needed.

##### bash

for bash you just need to run `contxt install bashrc`. the functions **ctx** and **cn** will be created in the .bashrc file

##### zsh

for **zsh** use `contxt install zsh`. the functions **ctx** and **cn** will be created in the first directory in the **FPATH** that is write and readable for the current user.

##### fish 

for **fish** use `contxt install fish`.  the functions **ctx** and **cn** will be created in the default function directory `~/.config/fish/functions`

### windows
currently the windows version is *work in progess* depending the shell integration. so there is currently no *package deployment* and also no
function mapping. the windows version can be downloaded as a binary, but any integration have to be done manually.


#### powershell 7 support
the default supported shell on windows is `powershell`. 
if you have **powershell 7** installed, you can set these as default shell as environment variable. `$env:CTX_DEFAULT_CMD = "pwsh"`

> **pwsh** is one of the possible commands for powershell. they exists also others for previews etc. fit them if needed.

ansii support is disabled as long the default powershell version is lower then 7. you can force the usage of ansii codes with 
`$env:CTX_COLOR = "ON"`



## workspaces

Workspaces are a set of directories they can be on different places. Any path can have different task they can be executed automatically if you enter or leava a workspace. this allows you to setup different needs depending on the workspace.


### usage

after executing the **shell integration** steps, you should use `ctx` instead of `contxt`

different to *contxt*. *ctx* is a shell function and is able to change the directory.
so contxt can still change the workspace, but after runtime, the shell will stay on the directory as before.

![image](https://user-images.githubusercontent.com/5437750/178095430-10da7bf9-8266-45cb-aa3c-23fa0604b3e6.png)

but **ctx**, as a shell function, is able to change the current directory.

![image](https://user-images.githubusercontent.com/5437750/178095493-ee07317c-c74d-407b-9cd9-7a793ccfb458.png)


### context navigate (cn)

for navigation use the `cn` command to change paths depending on the current workspace.

`cn` will change the directory to the last used path in this workspace.

`cn website` will change to the last matching path in the workspace, that contains the word *website*

`cn 2` will change to the third stored path. *(because it starts at 0 for the first stored path)*

`cn website php frontend` similar to `cn website` but if no *website* can be found on any of the assigned paths, *pah* and afterwards *frontend* will be used to find a matching path

#### create a new workspace
`ctx dir -w mywork` will create a new workspace named *mywork*. if mywork already exists the workspace will be entered only.

##### add path to current workspace
`ctx dir add` adds the current directory to the workspace. so you have to got to the directory first they needs to
be added to the workspace.

##### remove a path from a workspace
`ctx dir rm` removes the current directory from the current workspace. current directory means the actual path (pwd). like **add** you have to change into this directory. 

##### show all paths in current workspace
`ctx dir -paths` prints all assigned paths. any path have index.

##### list existing workspaces
`ctx dir list` prints all workspaces you have.

##### entering workspace

`ctx switch my-other-work` if this workspace exists you will leave the last one. this means **if you have a task assigned to the leave trigger** this will be executed for the last workspace. And **if you have a task assigned to the enter trigger** it will be excuted afterwards.
without to much detail about task. here a example of a task that have a enter and leave action and uses `kubectx` and `kubens`.


````yaml
config:
  autorun:
    onleave: leave
    onenter: init
  variables:
    namespace: production-ns
    context: "company-aws"
    context-leave: minikube
task:
  - id: init
    script:
      - kubectx ${context}
      - kubens ${namespace}
      - kubectl get pods
 
  - id: leave
    script:
      - kubectx ${context-leave}

````

so if this is the task for a path in workspace *company* and you run `ctx switch company` then the task with id **init** will be executed and the kubernetes context and namespaces will be set to what is needed there.

but to be sure, if you switch to any other workspace, that might not have any kubernetes related action, the **leave task** is triggered.
in this case it will just set the context to the local **minikube** to make sure what ever you do, the procution environment is not affected.

###### switch shortcut

you can switch the workspace also just by typing `ctx <existing-workspace>`.

### Tasks


[more about tasks](docs/documentation/tasks.md)


### used libraries

there are of course more dependencies, but these are important depending working with placeholders and variables.
so it is usefull to know how they are working.


go/template extension https://github.com/Masterminds/sprig [template docu](http://masterminds.github.io/sprig/)

> The Go language comes with a built-in template language, but not very many template functions. Sprig is a library that provides more than 100 commonly used template functions.


for parsing json by paths https://github.com/tidwall/gjson 

> GJSON is a Go package that provides a fast and simple way to get values from a json document. It has features such as one line retrieval, dot notation paths, iteration, and parsing json lines.
