
# cont(e)xt


**contxt** manage paths in workspaces while working on a shell. configure workspaces by tasks so you can use different versions of sdk's or to get the right context. for example by using `kubetctl` or `sdk`

### installation

##### packages
check out the [releases](https://github.com/swaros/contxt/releases) to get a pre-build package.

 [debian | ubuntu ](https://github.com/swaros/contxt/releases/download/v0.0.8-alpha/contxt_0.0.8-alpha_linux_amd64.deb)   -  [fedora | redhat | Suse](https://github.com/swaros/contxt/releases/download/v0.0.8-alpha/contxt_0.0.8-alpha_linux_amd64.rpm)
##### manual install
for build you need **golang 1.13** and **GNU Make**.

````shell

mkdir contxt
cd contxt
git clone git@github.com:swaros/contxt.git .
make

````
### Change Dir Function

there is no way to change a directory on console by a program/tool or whatever. of course, you can change to any directory while the program
is running, but all this changes are temporary. so if the program exists, the console is back to directory where the program
was started.

to change directories depending on paths in the workspace, you need to add a function to **.bashrc** or **.zshrc** that just use the regular
cd command.

##### bash

for easy switching paths on bash you can run this.

````bash
echo  'function cn() { cd $(contxt dir -i "$@"); }' >> ~/.bashrc

````

##### zsh

````zsh
echo  'function cn() { cd $(contxt dir -i "$@"); }' >> ~/.zshrc

````

now you can switch to any path fast by his index number like `cn 0` (after openening a new bash)

### console completion

run `contxt completion --help` to show how you can enable completion. 

for bash you need to run `source <(contxt completion bash)`. this you can also add to youre ~/.bashrc

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