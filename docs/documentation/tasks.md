# Task
this document will give a plain overview how task are created and executed.
tasks are a collection of scripts they are depending to the current directory. this tasks are defined in a taskfile named **.contxt.yml**.  
#### create  a new task
with `contxt create` you can create a simple task file.
this will looks like this
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

## structure 

the structure differs to most of all other yaml based task runners or ci tools. contxt using a list of task and not a `[string]task` list.
this results in a structure that may seems a little bit more complicated.
but this allows us to define a task with different behaviors depending on some requirements. 

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
in this case, we still execute this task on any system `ctx run task` but depending on the operating system, only the task that matches the requirements will be executed.

you can also combine these as much if you like,for example to have a task that will runs allways, and othes they will check if the can run or not.

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
       - echo %CD%
````
### config

**config** is a root element that defines the behavior of the task runner. 

#### variables

variables are placeholder that can be set globaly or in a task. but any variable is accesible globaly allways, even if you define them in a task.
**also keep in mind:** that means variables are not bound to a scope. 

to explain:

````yaml
config:
config:
  variables:
     test-output: hello
task:
   - id: testvar
     script:
       - echo "${test-output} world"
       
   - id: rewrite
     variables:
       test-output: "rescue the "
     script:
       - echo "${test-output} world"

````

because we can run mutlipe targets at ones, we will do this now `ctx run testvar rewrite`. 

````bash
[exec:async] testvar /home/example4/.contxt.yml
   testvar : hello world
[done] testvar
[exec:async] rewrite /home/example4/.contxt.yml
   rewrite : rescue the  world
[done] rewrite

````

so first **testvar** just prints the content of the global defined variable `${test-output}`.

afterwards **rewrite** is executed and redefines these variable and prints them.
and because this is a global change, it would be affect any other following tasks.

so if we exectue the same in different order `ctx run rewrite testvar`, we got a different outcome.
````bash
[exec:async] rewrite /home/tziegler/code/playground/go/ctx-examples/example4/.contxt.yml
   rewrite : rescue the  world
[done] rewrite
[exec:async] testvar /home/tziegler/code/playground/go/ctx-examples/example4/.contxt.yml
   testvar : rescue the  world
[done] testvar
````
now booth of the targets have the same content of the variable, because `rewrite` changed the value before `testvar` is executed.

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
there is additional output what is not interesting yet. you will see all targets on the right side of any path.
so if you use `contxt run init -a` the task will be executed in path no **0** and **1**.
if you use `contxt run build -a` it will run on target no **0** only because no other path a target named **build**
and if you run `contxt run clean -a` it will run on target **0** and **2**

###


