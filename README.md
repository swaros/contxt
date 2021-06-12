
# cont(e)xt


**contxt** helps you to organize projects and the assigned tasks and techniques. for this just a shell is needed.

### installation

##### packages
check out the [releases](https://github.com/swaros/contxt/releases) to get a pre-build package.

#### shell integration
contxt itself have no control about the current working directory. this is the regular behavior for all executables.
to *fix* this issue, contxt will be mapped by a shell function called **ctx**.
this behavior is similar to **contxt**.  including code completion.

you just need to run `contxt install bash|zsh|fish`. this will update user related shell init scripts.

### workspaces

workspaces are a set of directories they can be on different places. any path can have different task they can be executed automatically if you enter or leava a workspace. there are a couple of tools they can be used together with *contxt*.

so you can use [sdkman](https://sdkman.io/) to switch automatically different versions of skd's, or use [kubens](https://github.com/ahmetb/kubectx/) to set the correct namespace and context while working with *kubernetes*


### usage

contxt is a command line tool and focused to work on any environment that have a shell. there is no other dependency (yet).
after executing the **shell integration** steps, you should use `ctx` instead of `contxt`

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

`ctx my-other-work` if this workspace exists you will leave the last one. this means **if you have a task assigned to the leave trigger** this will be executed for the last workspace. And **if you have a task assigned to the enter trigger** it will be excuted afterwards.
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

so if this is the task for a path in workspace *company* and you run `ctx company` then the task with id **init** will be executed and the kubernetes context and namespaces will be set to what is needed there.

but to be sure, if you switch to any other workspace, that might not have any kubernetes related action, the **leave task** is triggered.
in this case it will just set the context to the local **minikube** to make sure what ever you do, the procution environment is not affected.

[more about tasks](docs/documentation/tasks.md)


##### used libraries 

for parsing json by paths https://github.com/tidwall/gjson 