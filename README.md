
# cont(e)xt


**contxt** helps you to organize projects and the assigned tasks and techniques. 

### installation

##### packages
check out the [releases](https://github.com/swaros/contxt/releases) to get a pre-build package.

#### shell integration
contxt itself have no control about the current working directory in the current shell. this is the regular behavior for all executables.
to *fix* this issue, contxt will be mapped by a shell function called **ctx**.

for bash you just need to run `contxt install bash`. this will update user related shell init scripts. 
for **zsh** use `contxt install zsh` and for **fish** use `contxt install fish` instead.

### workspaces

Workspaces are a set of directories they can be on different places. Any path can have different task they can be executed automatically if you enter or leava a workspace. this allows you to setup different needs depending on the workspace.


### usage

contxt is a command line tool and focused to work on any environment that have a shell. there is no other dependency (yet).
after executing the **shell integration** steps, you should use `ctx` instead of `contxt`

different to *contxt*. *ctx* is a shell function and is able to change the directory.
so contxt can change the workspace, but after runtime, the shell will stay on the current directory
````shell
user@localhost:~/project/alpha$ contxt switch omega
current workspace is now:omega
	paths stored in omega
	 0  /home/project/omega/frontend
	[1] /home/project/omega/backend
user@localhost:~/project/alpha$ pwd
/home/project/alpha
````

but **ctx**, as a shell function, is able to change the current directory.
````shell
user@localhost:~/project/alpha$ ctx switch omega
current workspace is now:omega
	paths stored in omega
	 0  /home/project/omega/frontend
	[1] /home/project/omega/backend
user@localhost:~/project/omega/backend$ pwd
/home/project/omega/backend
````
#### context navigate (cn)

for navigation use the `cn` command to change paths depending on the current workspace.

`cn` will change the directory to the last used path in this workspace.

`cn website` will change to the last matching path in the workspace, that contains the word *website*

`cn 2` will change to the third stored path. *(because it starts at 0 for the first stored path)*

`cn website php frontend` similar to `cn website` but if no *website* can be found on any of the assigned paths, *pah* and afterwards *frontend* will be used to find a matching path

##### create a new workspace
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

[more about tasks](docs/documentation/tasks.md)


##### used libraries 

go/template extension https://github.com/Masterminds/sprig [template docu](http://masterminds.github.io/sprig/)
for parsing json by paths https://github.com/tidwall/gjson 