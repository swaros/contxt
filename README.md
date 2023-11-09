
# cont(e)xt

version: `0.5.3`

**contxt** is a development Tool that aims to help you to keep track of your projects and their related content. do not waist youre time by looking for the right directory, or the right command to execute, just because you have to many projects and they are all different.


### the main goals:

 - **working with mutliple projects** and having one way to build, test and deploy them. this means you can have different projects, they have different build tools, different test tools and different deployment tools. but you can use the same commands to execute them.

 - **works in the shell**. there is no dependency to any IDE, container setup or any other tool. you can use it in any shell, on any machine, with any project.

 - **shared tasks**. you can share tasks between projects. so if you have a task that is needed in different projects, you can share them. this means you can have a task that is used in different projects, but the task is only stored once. so if you change the task, it is changed for all projects.

 - **template everything**. you can use _go/template_ to template any file. also you can use any _yaml_ or _json_ file as value storage. this allows you to create files depending on the current project, or any other context. you can even template source code files, to create different source code files depending on the current project settings.

 - **based on json/yaml**. the usage of yaml and/or json based Data is a common way to store and exchange data. _contxt_ is also capable to read, manipulate and write them back without having a need to use some external tools like _jq_. So loading an _docker-compose.yml_, replacing the image name, and write them back is quite easy.

 - **running task**. the taskrunner is not build as a replacement for other build tools like make, maven, gradle, gulp, grunt, etc. it is ment as _On-Top_ Tool, to run anything by using the needs for the project. _but in a controlled way_. you can just let a build fail, or you can let it fail and execute a different task to recover, or whatever make sense. Also you make sure anything is setup by youre needs. So if the build (tests) is failing because the database service was not ready? Add a Listener to the database service, and if the service is ready, the build will be triggered. 


## installation
check out the [releases](https://github.com/swaros/contxt/releases) to get a pre-build package.

#### shell integration

any shell is restricted to the current context. this means if you change the directory "inside" the process, the shell will not change the directory "outside" of this process. this is a problem for any tool that is running in a shell and have a need to keep the directory path. If the process is done, the shell is in the same directory as before. 

to *fix* this issue, contxt will be mapped by a shell function called **ctx** (for contxt itself), and **cn** for changing directories depending the current project. 

this works because shell functions are in same context.
so for any supported shell a shell function is needed.

##### bash

for bash you just need to run `contxt install bashrc`. the functions **ctx** and **cn** will be created in the .bashrc file

##### zsh

for **zsh** use `contxt install zsh`. the functions **ctx** and **cn** will be created in the first directory in the **FPATH** that is write and readable for the current user.

##### fish 

for **fish** use `contxt install fish`.  the functions **ctx** and **cn** will be created in the default function directory `~/.config/fish/functions`

### windows
currently the windows version is *work in progess* depending the **shell integration**. but it have close the same behavior like the linux versions, **except** the contxt mapping as **ctx** function. it have currently an issue depending automatically change the dir on a switch command. so you need execute the `cn` command afterwards, to change into the working dir.


#### powershell 

run `contxt install powershell`. this will create the `cn`and `ctx`functions in the current profile. if you not have an profile for the current User in the current workspace, then you need to add the flag `--create-profile` if contxt should create them.


the default supported shell on windows is `powershell`. 
if you have **powershell 7** installed, you can set these as default shell as environment variable. `$env:CTX_DEFAULT_CMD = "pwsh"`

> **pwsh** is one of the possible commands for powershell. they exists also others for previews etc. fit them if needed.

ansii support is disabled as long the default powershell version is lower then 7. you can force the usage of ansii codes with 
`$env:CTX_COLOR = "ON"`



## workspaces

 - Workspaces are a set of directories they can be on different places. 
 - Any path can have different task they can be executed automatically if you enter or leava a workspace. this allows you to setup different needs depending on the workspace.
 - any path in a workspace can have a different project and role setting, that can be used as variables in **contxt tasks**, so this
 tasks can handle sub projects on different machines, where services, projects (and so on) exists on different paths.


### using workspaces and navigate to them

after executing the **shell integration** steps, you should use `ctx` instead of `contxt`

different to *contxt*. *ctx* is a shell function and is able to change the directory.
so contxt can still change the workspace, but after runtime, the shell will stay on the directory as before.

![image](https://user-images.githubusercontent.com/5437750/178095430-10da7bf9-8266-45cb-aa3c-23fa0604b3e6.png)

but **ctx**, as a shell function, is able to change the current directory.

![image](https://user-images.githubusercontent.com/5437750/178095493-ee07317c-c74d-407b-9cd9-7a793ccfb458.png)

> **note**: on windows the `ctx switch` is currently not change the directory. you need run `cn` afterwards.

### context navigate (cn)

for navigation use the `cn` command to change paths depending on the current workspace.

`cn` *without arguments*, will change the directory to the last used path in this workspace.

`cn website` will change to the last matching path in the workspace, that contains the word *website*. if you have multiple paths in the workspace, that contains *website*, the path is used, they have the word *webste* more on the right side as others, to get the obvious needed path.

`cn 2` will change to the path depending the ndex.
the index is created in order of the path was stored added. 

`cn website build` similar to `cn website` but it looks also if the path *build* is in the path. so you can be more specific.

#### create a new workspace
`ctx workspace new mywork` will create a new workspace named *mywork*. 

##### add path to current workspace
`ctx dir add` adds the current directory to the workspace. so you have to got to the directory first they needs to
be added to the workspace.

##### remove a path from a workspace
`ctx dir rm` removes the current directory from the current workspace. current directory means the actual path (pwd). like **add** you have to change into this directory. 

##### show all paths in current workspace
`ctx dir -paths` prints all assigned paths. any path have index.

##### list existing workspaces
`ctx workspace list` prints all workspaces you have.

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

### About this Repository
this is the main repository for the contxt project. It also contains any **package** as **module** instead of having them in different repositories. this is because i am lazy (what is by the way my motivation to write this tool) and i do not want to maintain different repositories.
Even more **go** is able to handle packages independent from the repository. so i can use the same package in different projects, without having a need to copy them.

#### modules
|code ref|module|version|description|
|---|---|---|---|
|_internal_|yacl||Yet Another Config Loader|
|_internal_|yamc||yaml/json data mapper|
|_internal_|runner||contxt task runner V2|
|_internal_|ctxtcell||experimental controll elements|
|_internal_|configure||configure contxt|
|_internal_|dirhandle||collection of directory handling functions|
|_internal_|systools||collection of system tools|
| [![Go Reference](https://pkg.go.dev/badge/github.com/swaros/contxt/module/trigger.svg)](https://pkg.go.dev/github.com/swaros/contxt/module/trigger)|trigger|v0.4.0|callback handler|
|_internal_|linehack||line based text processing. experimental|
|_internal_|ctemplate||template engine based on go/template|
|_internal_|ctxout||configurable output handler|
|_internal_|taskrun||contxt task runner V1|
| [![Go Reference](https://pkg.go.dev/badge/github.com/swaros/contxt/module/awaitgroup.svg)](https://pkg.go.dev/github.com/swaros/contxt/module/awaitgroup)|awaitgroup|v0.4.0|awaitgroup replaces sync.WaitGroup|
|_internal_|shellcmd||shell command execution for V1|
|_internal_|ctxshell||readline based shell with cobra support for V2|
|_internal_|tasks||contxt V1 task management|
|_internal_|yaclint||yacl config auto linter|
|_internal_|mimiclog||logger interface|
|_internal_|process||process management|




### used libraries

there are of course more dependencies, but these are important depending working with placeholders and variables.
so it is usefull to know how they are working.


go/template extension https://github.com/Masterminds/sprig [template docu](http://masterminds.github.io/sprig/)

> The Go language comes with a built-in template language, but not very many template functions. Sprig is a library that provides more than 100 commonly used template functions.


for parsing json by paths https://github.com/tidwall/gjson 

> GJSON is a Go package that provides a fast and simple way to get values from a json document. It has features such as one line retrieval, dot notation paths, iteration, and parsing json lines.
