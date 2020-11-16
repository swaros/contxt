# cont(e)xt

**contxt** manage paths in workspaces while working on a shell. configure workspaces by tasks so you can use different versions of sdk's or to get the right context. for example by using `kubetctl` or `sdk`

### installation

for build you need **golang 1.13** and **GNU Make**.

````shell
mkdir contxt
cd contxt
git clone git@github.com:swaros/contxt.git .
make
````

without make just compile by using go.
````shell
go build -i -o ./bin/contxt cmd/cmd-contxt/main.go
````
> if you do so, there will no version and buildnumber created.

for bash user install.
````bash
make install-local
````

for easy switching paths on bash you can run this.
````bash
echo 'function cn() { cd $(contxt dir -i "$@"); }' >> ~/.bashrc
````
now you can switch to any path fast by his index number like `cn 0` (after openening a new bash)

### workspaces
workspaces are a set of directories they can be on different places. any path can have different task they can be executed automatically if you enter or leava a workspace. there are a couple of tools they can be used together with *contxt*. 
so you can use [sdkman](https://sdkman.io/) to switch automatically different versions of skd's, or use [kubens](https://github.com/ahmetb/kubectx/) to set the correct namespace and context while working with *kubernetes*

### usage
contxt is a command line tool and focused to work on any environment that have a shell. there is no other dependency.

##### create a new workspace
`contxt dir -w mywork` will create a new workspace named *mywork*. if mywork already exists the workspace will be entered only.

##### list existing workspaces
`contxt dir list` prints all workspaces you have.

##### entering workspace
`contxt my-other-work` if this workspace exists you will leave the last one. this means **if you have a task assigned to the leave trigger** this will be executed for the last workspace. And **if you have a task assigned to the enter trigger** it will be excuted afterwards.

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
so if this is the task for a path in workspace *company* and you run `contxt company` then the task with id **init** will be executed and the kubernetes context and namespaces will be set to what is needed there.

but to be sure, if you switch to any other workspace, that might not have any kubernetes related action, the **leave task** is triggered. 
in this case it will just set the context to the local **minikube** to make sure what ever you do, the procution environment is not affected.

[more about tasks](docs/documentation/tasks.md)

##### add path to current workspace
`contxt dir add` adds the current directory to the workspace. so you have to got to the directory first they needs to
be added to the workspace.

##### show all paths in current workspace
`contxt dir -paths` prints all assigned paths. any path have index. 


