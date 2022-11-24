# contxt.yml task
this document will give a plain overview how task are created and executed.
tasks are a collection of scripts they are depending to the current directory. this tasks are defined in a taskfile named **.contxt.yml**.
<!-- TOC -->

- [contxt.yml task](#contxtyml-task)
    - [task create and run](#task-create-and-run)
        - [basic use case](#basic-use-case)
            - [create  a new task](#create--a-new-task)
            - [list and run a task](#list-and-run-a-task)
        - [extended use-case](#extended-use-case)
            - [how to avoid running asynchronously](#how-to-avoid-running-asynchronously)
            - [run task from anywhere](#run-task-from-anywhere)
            - [list all task](#list-all-task)
- [structure](#structure)
    - [Task](#task)
        - [task structure](#task-structure)
            - [ID string](#id-string)
            - [Variables](#variables)
            - [Requires](#requires)
                - [system string](#system-string)
                - [exists, notExists](#exists-notexists)
                - [Variables, Environment](#variables-environment)
            - [Stopreasons](#stopreasons)
            - [Options](#options)
                - [ignorecmderror bool](#ignorecmderror-bool)
                - [format string](#format-string)
                - [stickcursor bool](#stickcursor-bool)
                - [colorcode and bgcolorcode string](#colorcode-and-bgcolorcode-string)
                - [panelsize int](#panelsize-int)
                - [displaycmd bool](#displaycmd-bool)
                - [hideout bool](#hideout-bool)
                - [invisible bool](#invisible-bool)
                - [maincmd string, mainparams list](#maincmd-string-mainparams-list)
                - [workingdir string](#workingdir-string)
    - [config](#config)
        - [sequencially bool](#sequencially-bool)
        - [coloroff bool](#coloroff-bool)
        - [loglevel](#loglevel)
        - [variables](#variables)
            - [working with variables](#working-with-variables)
            - [asynchronously behaviour](#asynchronously-behaviour)
            - [set variables from command line](#set-variables-from-command-line)
            - [import variables](#import-variables)
                - [import as short cut](#import-as-short-cut)
        - [autorun](#autorun)
            - [onenter](#onenter)
            - [onleave](#onleave)
            - [example](#example)
        - [Imports](#imports)
        - [Use. Shared Tasks 1/2](#use-shared-tasks-12)
            - [example for linux users](#example-for-linux-users)
        - [Require. Shared Task 2/2](#require-shared-task-22)
        - [allowmultiplerun bool](#allowmultiplerun-bool)

<!-- /TOC -->
## task create and run 
### basic use case
#### create  a new task
with `contxt create` you can create a simple task file.
this will look like this
````yaml
task:
  - id: script
    script:
      - echo "hello world"
````
#### list and run a task
with `contxt run` you will see all targets they can be started. 

````bash
:> contxt run
used taskfile:	/home/user/project/.contxt.yml
tasks count:  	1
existing targets:
	 script
````

>in this case it doesn't matter if you are in the workspace where the current path is assigned to or not. if a task exists in this path, it can be used.

to start the task just add the name to the run option `contxt run script`

````bash
contxt run script
[exec:async] script /home/user/project/.contxt.yml
     script :   hello world 
[done] script
````
### extended use-case 
you can also run multiple targets sequentially 
`ctx run target-1 target-2 tagets-3`
or asynchronously by separating them with comma.
`ctx run target-1,target-2,target-3`
and you can combine this
`ctx run target-1,target-2 target-3`

task can be started also by different definitions in the task-file. for now just a overview.
|definition|behaviour|
|--|--|
| needs | a list of task's they need to be executed at least ones  |
| runTargets | a list of task they have to be executed together asynchronously |
| next | a list of task's they have to be executed afterwards|
|listener|*can* also be used to execute task or a script depending on the output|
#### how to avoid running asynchronously
the regular behaviour is to run task  asynchronously to speed things up. but it is also possible to change this
behaviour by configure the task-file.
````yaml
config:
  sequencially: true
````
#### run task from anywhere
by default **contxt** will run task in the current directory. but you can also run all task in the current workspace.
`contxt run script -a`
then **contxt** iterates over all assigned paths, checks if a task file exists, and if they have a task named *script*.

if this is the case, this task will be executed in this path.
> this also means that you should name you targets with care. a task name that exists in different paths should 
> do the same. so for example it is not a good idea to make dangerous task and name them like init.
> a task name should always reflect the job he have to do.

#### list all task
especially to get a overview what tasks are defined, the `contxt dir`command is helpfully, because it shows all targets in context with the assigned tasks.
````bash
:> contxt dir
 current directory: /home/user/project/contxt/bin
 current workspace: contxt
 contains paths:
       path:  no 0 /home/user/project/contxt targets[ build clean init]
       path:  no 1 /home/user/project/wiki-contxt/contxt.wiki targets[ init]
       path:  no 2 /home/user/project/contxt/bin targets[ clean]

````
you will see all targets on the right side of any path.
so if you use `contxt run init -a` the task will be executed in path 
no **0** and **1**.
if you use `contxt run build -a` it will run on target no **0** only 
because no other path a target named **build**
and if you run `contxt run clean -a` it will run on target **0** and **2**

# structure 
## Task
the structure differs to most of all yaml based task runners or ci tools.
contxt using a list of task and not a `[string]task` list.
this results in a structure that may seem a little  more complicated.
````yaml
task:
  - id: first-task
  - id: second-task
````

but this allows us to define a task with different behaviours 
depending on some requirements. 

like this.

````yaml
task:
   - id: task
     require:
       system: linux
     script:
       - echo "hello linux"

   - id: task
     require:
       system: windows
     script:
       - echo "hello windows"
````
in this case, we still execute this task on any system `ctx run task` 
but depending on the operating system, only the task that 
matches the requirements will be executed.

you can also combine these as much if you like, for example to have a 
task that will run always, and others they will check if the can run or not.

````yaml
task:
   - id: task
     script:
       - echo "the current directory is "
       
   - id: task
     require:
       system: linux
     script:
       - pwd

   - id: task
     require:
       system: windows
     script:
       - Get-Location | Foreach-Object { $_.Path }
````

### task structure
#### ID (string)

the ID sets the identifier for the task. as described already, they 
can be present multiple times, with different requirements.
````yaml
task:
   - id: task-identifier           
````

#### Variables 
the variables' section defines variables, or update existing variables.
these variables will be existed also if the task is done.

> **IMPORTANT** variables synced but always global. be aware of raise conditions if you use the same variables in different tasks.

````yaml
task:
  - id: task-identifier
    variables:
      example-var: something               
````

these variables can be used by the placeholder key like `${example-var}`

> see variables documentation *config -> variables*. there are more details about how to use variables.

#### Requires
Require checks different cases. if one of these requirements 
are not matching, then this task section is ignored. **this is not meaning
the whole task is ignored**

for example depending on the system that is used, only one of
these section from **task** will be executed.

````yaml
task:
   - id: task
     require:
       system: linux
     script:
       - echo "hello linux"

   - id: task
     require:
       system: windows
     script:
       - echo "hello windows"
````

##### system (string)
  
this requirement checks the system. the current system have to match 
with the reported ones. here the most used (the system name is taken
from go's `runtime.GOOS`)

- windows
- linux
- darwin

````yaml
task:
  - id: run-bat
    require:
      system: windows  
````

##### exists, notExists
these two requirements checks if a file exists, or not.
you can use any variable here. (environment is also a variable)

````yaml
task:
  - id: copy-files-if-exists
    require:
      exists: 
        - ${HOME}/target/
        - ./source/copy-this-file.cpp
````

any of these files have to exists. if one of them missing, the task is
ignored.

to check if a file not exists, use `notExists` instead


````yaml
task:
  - id: create-file
    require:
      notExists:         
        - ./source/create-this-file.cpp
````
also if one of the files in the lists, exists, the whole task will be ignored.


##### Variables, Environment
these two requirements works in the same way. the different is,
`variables` checks the contxt variables, and `environment` is checking
against the environment variables.

````yaml
task:
  - id: deploy
    require:
      variables: 
        target: "*"
      environment:
        DEPLOY_TO: "stage"
````

this requirement checks the content of a context-variable/environment-variable.
the first letter tells _contxt_  how to verify the variable/environment

- the usual check is **equals**. 
````yaml
  target: "development"
````
- the example above for **equals** is the shortcut for **=**
````yaml
  target: "=development"
````
- to check **not equals** use **!**
````yaml
  target: "!prodution"
````
- to check if the variable is **not empty** use * (no matter what the value is. 
- just not "". but they must exists)
````yaml
  target: "*"
````
- to check if the variable is **greater than** use **>**
````yaml
  version: ">4.1"
````
- to check if the variable is **lower than** use **<**
````yaml
  version: "<4.3"
````

#### Stopreasons

this section defines a _trigger_ that reacts on an error.

#### Options

the option section contains the task specialized configuration.
this can be done for any task, even if they have the same **ID**.
some of them, like the visibility, have no effect. if one task with the 
same id is visible, then this task will be shown (autocomplete, task lists) 
even if any other task are invisible.

but any runtime specific setting is used for this part of the task only. 

##### ignorecmderror (bool)
if any task reports an error, contxt will exit. independent if any other
parallel running task is doing well, or this error could be 
ignored (or handled elsewhere)

so if you have a task, they can fail, set `ignorecmderror` to true.

````yaml
task:
  - id: this-can-fail
    options:
      ignorecmderror:  true
````

##### format (string)
formats the left label for the output. you can use %s as placeholder 
what would be used for the current **ID** of the task.

````yaml
task:
  - id: own-label
    options:
      format: "WATCH ME !!!111 ----  I AM IMPORTANT ---- (i am the task named %s)"
````

you can also use any existing variable.

````yaml
task:
  - id: own-label
    options:
      format: "hello ${USER} ... did you see that log-output? -->"
````

also the internal color codes can be used

````yaml
task:
  - id: own-label
    options:
      format: "<b:red><f:white>this is the<f:yellow>DANGER ZONE</>"
````

##### stickcursor (bool)
just a simple flag to use escape sequences, to stay at the last line.
there is no big magic behind, so longer lines, then the max-with o the terminal,
will push the content.

but can be usefully to prevent spamming minor information, 
while other important stuff, from parallel running task, get lost.

````yaml
task:
  - id: lot-of-text-lines-incoming
    options:
      stickcursor: true
````

##### colorcode and bgcolorcode (string)
foreground and background ANSI color code for the label.

````yaml
task:
  - id: colored-task
    options:
      colorcode: "97"
      bgcolorcode: "47"
````

##### panelsize (int)
defines the width of the left panel (count of chars) for the output screen.

````yaml
task:
  - id: example-task
    options:
      panelsize: 15      
````

##### displaycmd (bool)
if this is set to true, contxt will show additional information
on the regular output, like the composed comand, the current pid for
this task and so on. **this is only active for this section of the task**

this is usefull for fast inspecting, without enable debug
and having a lot output from any sources.

````yaml
task:
  - id: example-task
    options:
      displaycmd: true
````


##### hideout (bool)
if this is set to true, any output to screen is disabled for this task.

````yaml
task:
  - id: example-task
    options:
      hideout: true
````

##### invisible (bool)
if true, this task will not shown in autocomplete, or in the task list.
> keep in mind. if the task having mutliple sections, then all of them needs to be invisible.

````yaml
task:
  - id: example-task
    options:
      invisible: true
````

##### maincmd (string), mainparams (list)
**maincmd** defines the main command for the script section. if not set,
then the default is `bash`

for this you would often need to define startup arguments. 
this is done by **mainparams**

````yaml
task:
  - id: php-example-task
    options:
      maincmd: "php"
      mainparam: 
        - "-r"
    script:
      - phpinfo();
      - var_dump($_SERVER);
````

##### workingdir (string)

if you need to change the directory, while running a task, the following
example will **not work**.

````yaml
task:
  - id: example-task
    script:
      - cd build
      - make
````

this is because any entry in the script list, will run in his own context.
so it has by default his own `bash` running.

a quickfix would just use the **- |** yaml annotation, to get a long string, 
that can have line breaks.

so it would **work as expected**
````yaml
task:
  - id: example-task
    script:
      - | 
       cd build
       make
````

but then you have 2 task running as one task (what makes sense i  many cases). 
the alternative is to set the **workingdir**

````yaml
task:
  - id: example-task
    options:
      workingdir: build
    script:      
      - make
````



---
## config

*config* is a root element that defines the behaviour of the task runner. 
### sequencially (bool)
by default any task will be started asynchronous. 
this behavior can be disabled by this option. set it to true, 
and any task and any dependency will run one by one after the first one is done. 
````yaml
config:
  sequencially: true     
````
 
### coloroff (bool)
contxt enables ANSII colored output if possible. you can turn this off
by this setting. 
````yaml
config:
  coloroff: true     
````

runtime flag is `--coloroff` , `-c`

### loglevel
sets the logger level. 
possible values are
- panic
- fatal
- error
- warn
- info
- debug
- trace

````yaml
config:
  loglevel: debug
````

runtime flag is `--loglevel debug`

> used logger library is [logrus](https://github.com/sirupsen/logrus)

### variables

````yaml
config:
  variables:
     hostname: "localhost"
     update-command: git stash && git pull --rebase && git stash pop
````

variables are dynamic place-holder that can be set globally or in
a task. but any variable is accessible globally always, even if you define 
them in a task.
**also keep in mind:** that means variables are not bound to a scope. 

to use them just write `${hostname}`. variables can be used in values only. 
> if you like to being more flexible, contxt supports [sprig](http://masterminds.github.io/sprig/) 
> by reading external `yaml` or `json` files. 
> use `ctx create import <filename>` to create a relationship
> to the files that should be used for template

depending on the global scope of all variables you need to keep 
this in mind on more complex dependencies.
#### working with variables
to explain the variables behaviour while runtime:

````yaml
config:
  variables:
     test-output: hello
task:
   - id: testvar
     script:
       - echo "${test-output} world"
       
   - id: rewrite
     variables:
       test-output: "rescue the"
     script:
       - echo "${test-output} world"

````
because we can run multiple targets, we will do this 
now by using`ctx run testvar rewrite`. 
````bash
[exec:async] testvar /home/example4/.contxt.yml
   testvar : hello world
[done] testvar
[exec:async] rewrite /home/example4/.contxt.yml
   rewrite : rescue the world
[done] rewrite

````

so first **testvar** just prints the content of the global 
defined variable `${test-output}`.

afterwards **rewrite** is executed and redefines these variable.
and because this is a global change, it would be affect any other following tasks.

so if we execute the same in different order `ctx run rewrite testvar`,
we got a different outcome.
````bash
[exec:async] rewrite /home/tziegler/code/playground/go/ctx-examples/example4/.contxt.yml
   rewrite : rescue the world
[done] rewrite
[exec:async] testvar /home/tziegler/code/playground/go/ctx-examples/example4/.contxt.yml
   testvar : rescue the world
[done] testvar
````
now booth of the targets have the same content of the 
variable, because `rewrite` changed the value before `testvar` is executed.

#### asynchronously behaviour
important depending on this behaviour, it is only reliable if the 
tasks are running `sequencially`. so if you start booth task in his
own process, then you will get a different behaviour.

to start all task together use `ctx run rewrite,testvar`. 

````bash
[exec:async] rewrite /home/tziegler/code/playground/go/ctx-examples/example4/.contxt.yml
[exec:async] testvar /home/tziegler/code/playground/go/ctx-examples/example4/.contxt.yml
    testvar :   hello world 
    rewrite :   rescue the  world 
[done] rewrite,testvar
````

as you can see, the variable in the task *testvar* is not 
changed, because all tasks started close at the same time. 
it can be also the case the task running in different orders. 

> short explanation. the arguments of run is any **ctx run** *sequentially* *sequentially* 
> so any argument split by space will be run one be one, after the task before is done.
> but any of these arguments can contain multiple targets separated by comma they will be started asynchronously.
> you can test this with `ctx run testvar rewrite,testvar testvar`

#### set variables from command line
variables are set at first in the config root section. but these 
variables can also be set from outside with run flags `-v, --var stringToString`
this flag can be used multiple to overwrite the default value of different variables
`ctx run task -v firstvar=new-out -v secondvar="new out"`

depending on the example above:
`ctx run testvar -v test-output="this is not my"`
````bash
[exec:async] testvar /home/tziegler/code/playground/go/ctx-examples/example4/.contxt.yml
    testvar :   this is not my world 
[done] testvar
````

but still. variables redefined in tasks are not affected.
`ctx run rewrite -v test-output="this is not my"`
````bash
[exec:async] rewrite /home/tziegler/code/playground/go/ctx-examples/example4/.contxt.yml
    rewrite :   rescue the  world 
[done] rewrite
````

#### import variables
you can import `yaml` and `json` files as variables and use the content of them.
as example, we use the official **docker-compose.yml** from *postgress*.
````yaml
version: '3.1'
services:
  db:
    image: postgres
    restart: always
    environment:
      POSTGRES_PASSWORD: example

  adminer:
    image: adminer
    restart: always
    ports:
      - 8080:8080
````
for loading them, add file the path to the import section in the config.
````yaml
config:
  imports:
    - docker-compose.yml
````
this will load the whole file as variable. to access the values 
we use [gson](https://github.com/tidwall/gjson) for the path, and 
the file-name as entry point. `${`*filename*`:`*gson.path*`}`

to get the **image name** from the service named **db** the 
Placeholder would look like this
`${docker-compose.yml:services.db.image}`

so in a task you can use them as you like. for example:
````yaml
config:
  imports:
    - docker-compose.yml
task:
  - id: script
    script:      
      - echo "used image is ${docker-compose.yml:services.db.image}"
      - echo "you have to use ${docker-compose.yml:services.db.environment.POSTGRES_PASSWORD} as password"
      - echo "a adminer instance is running too on port ${docker-compose.yml:services.adminer.ports.0}"
````

````bash
[exec:async] script /home/tziegler/code/playground/go/ctx-examples/example5/.contxt.yml
     script :   used image is postgres 
     script :   you have to use example as password 
     script :   a adminer instance is running too on port 8080:8080 
[done] script

````

##### import as short cut
especially if you have to use long paths (for example if you have to use 
files in different directories) it would make sense to use a shortcut 
instead. you just need to write them behind the file name in the import list.

so now we use the name *postgres* as shortcut instead the filename docker-compose.
````yaml
config:
  imports:
    - docker-compose.yml postgres
task:
  - id: script
    script:      
      - echo "used image is ${postgres:services.db.image}"
      - echo "you have to use ${postgres:services.db.environment.POSTGRES_PASSWORD} as password"
      - echo "a adminer instance is running too on port ${postgres:services.adminer.ports.0}"
````

### autorun
the autorun option defines task they will be executed if you switch the
workspace. this task is regular contxt task.

this should help to set up automatically all needs if you have to work
on this project and needs to set up something.

````yaml
config:
  autorun:
    onleave: leave
    onenter: init
task:
  - id: init
    script:
      - echo "doing something to get this project running" 
      - git pull --rebase
      - and so on
 
  - id: leave
    script:
      - echo "doing something to cleanup this project til i come back"
      - rm -rf *.log
      - and so on
````

#### onenter
this defines the task that should be executed if you enter this workspace.
entering the workspace means 
 `ctx switch <workspace>` 

#### onleave
this defines the task that should be executed if you leave the workspace.
leave means you switch to another workspace.

#### example
for example: if you are working on project *java-project-version-11,* 
and you have to work with *java-project-version-8* execute 
`ctx switch java-project-8` then the first that will be executed, 
is any **onleave** task in the current *java-project-version-11* project. 

and then afterwards the *onenter* task in *java-project-version-8*

> this affects any autorun task definition in any assigned path to this workspace. but, of course, any of these task runs in his own path.

### Imports
defines a list of **yaml** or **json** files, they will be imported as variable map.

> **NOTE** for yaml files the only accepted extensions are *.yml* and *.yaml*. for json it is *.json' only.


accessing these variables is similar to the regular `variables`.
any imported json or yaml file will be stored by the filename
as key (full path) or by a specific key that just added in the same line 

````yaml
config:
  imports:
    - path/to/file <optional-key-name>
    
````

as example

````yaml
config:
  imports:
    - docker-compose.yml docker
    - composer.json
task:
  - id: script
    script:      
      - echo "used database image is ${docker:services.db.image}"
      - echo "php platform is {composer.json:config.platform.php}"
````

> see **imports** sub-section from the **variables** in this document, for details about working with variables

### Use. Shared Tasks (1/2)
shared tasks are **contxt** tasks defined in a seperated location in the user home dir.

`$HOME/.contxt/shared`

if you have anything that needs to be done in different projects, you can
create a folder in this path. this folder needs a sub folder 
called **source** and there we can put the `contxt.yml`


here the path as tree.
- $HOME
  - .contxt
    - shared
      - source
        - .contxt.yml

#### example for linux users

```shell
> mkdir -p $HOME/.contxt/shared/my-shard-task/source
> cd $HOME/.contxt/shared/my-shard-task/source
> ctx create
write execution template to  /home/user/.contxt/shared/my-shard-task/source/.contxt.yml
> cat .contxt.yml
task:
  - id: script
    script:
      - echo "hello world"
```

if you are working already in a workspace, then just go back to them by typing `cn`.

edit the .contxt.yml file in your current project and add the usage for the new script.


````yaml
config:
  use:
    - my-shard-task
````

now the `script` task is also shown as valid task target (even in autocomplete)

![image](https://user-images.githubusercontent.com/5437750/203767268-12d7e1b7-579e-464c-81ff-99773e6c4de9.png)

and can be started like it would be a part of the current tasks.

![image](https://user-images.githubusercontent.com/5437750/203767843-3df4d868-37e6-4010-b54b-692802cd165d.png)

these task, defined with **use**, will be run in the shared 
context before any other task from the "real" tasks are executed.


### Require. Shared Task (2/2)

read part (1/2) before, depending on **use**, because there the meaning of
shared task is explained. and how to create a shared task.

similar to `use` the `require` will be make use of the shared task.
but instead of just running them in front of the "real" task, in the 
shared context (including the shared path), the whole shared tasks 
are merged with the current tasks.

````yaml
config:
  require:
    - my-shard-task
````

from now on, anything from the shared task is part of current tasks.
even if you execute `ctx lint` you will see the external *script* task 
in the current project.

if you execute *script*, you will not see any difference to the regular execution.

![image](https://user-images.githubusercontent.com/5437750/203771904-d7ec69e1-fa07-4eb8-bfad-9d0014c6c43c.png)

### allowmultiplerun (bool)

the regular behavior is to run all dependencies once, and 
no other task twice at the same time.

it might be the case, this should be disabled, 
so anytime a task is required, it will be executed 
(for example to get the current time).

this option is then valid for any task in this task file.

````yaml
config:
  allowmultiplerun: true    
````


