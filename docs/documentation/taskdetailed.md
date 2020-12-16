# Tasks definition

a contxt task is defined in yaml and always named **.contxt.yml**. a exception of this rule is the *personal taskfile* that will be used for the current user only. you can name the task file like **root.contxt.yml** and this will only used, if the current user is **root**.
this will work with any user, so you could have different taskfiles dependig on the user.

## structure

**task** is the root for a list of named tasks (id)

````yaml
task:
  - id: example-1
  - id: example-2
````

if you would execute `contxt run` it shows these targets, even if they
have nothing to do

````
contxt run
used taskfile:	ctx-examples/example1/.contxt.yml
tasks count:  	2
existing targets:
	 example-1
	 example-2
````

you could also *run* this target by `contxt run example-1` and you would just see
````
contxt run example-1
[exec:async] example-1 ctx-examples/example1/.contxt.yml
[done] example-1
````
## script
any of these task can have a **script** section they contains a list
of console commands. *bash is used by default*

````yaml
task:
  - id: example-1
    script:
      - echo "this is example 1"
  - id: example-2
    script:
      - echo "this is example 2"
````

> **important** any single line is a new process. you need to keep this in mind because if you need to change directories, or set variables, you have to put this in one line. use **- |** annotation to write mutliple lines in one script.

if you run the *example-1* target again, it would looks like so

````
contxt run example-1
[exec:async] example-1 ctx-examples/example1/.contxt.yml
  example-1 :   this is example 1 
[done] example-1
````

but you can also run boot targets `contxt run target1 target2`
````
contxt run example-1 example-2
[exec:async] example-1 ctx-examples/example1/.contxt.yml
  example-1 :   this is example 1 
[done] example-1
[exec:async] example-2 ctx-examples/example1/.contxt.yml
  example-2 :   this is example 2 
[done] example-2

````

in this case the first target *example-1* will be executed and afterwards the second target *example-2*

## simultaneously execution
you could also run this targets paralell. then you need to combine them
as one parameter seperated by comma.
`contxt run example-1,example-2`

to show this is working, a slepp in the example would help.

````yaml
task:
  - id: example-1
    script:
      - echo "start example 1"
      - sleep 5
      - echo "done example 1"
  - id: example-2
    script:
      - echo "start example 2"
      - sleep 2
      - echo "done example 2"
````

now you start booth task at the same time
````
contxt run example-1,example-2
[exec:async] example-1 ctx-examples/example1/.contxt.yml
[exec:async] example-2 ctx-examples/example1/.contxt.yml
  example-1 :   start example 1 
  example-2 :   start example 2 
  example-2 :   done example 2 
  example-1 :   done example 1 
[done] example-1,example-2
````

you can combine these behavior to define a kind of stages. 
as a example here a simulated buildjob.

we have a backend build job, a frontend build job and a deploy job they depends booth and have to run afterwards.
````yaml
task:
  - id: build-backend
    script:
      - echo "enter Backend"
      - |
        for job in git-update run-maven run-test build-artifacts get-coffee update
        do
          echo "backend doing .... [$job] "
          sleep 1
        done 
      - echo "leave Backend"
  - id: build-frontend
    script:
      - echo "enter Frontend"
      - |
        for job in create wait finalize
        do
          echo "frontend ----> $job <---"
          sleep 1
        done 
      - echo "leave Frontend"
  - id: deploy
    script:
      - echo "deploy frontend"
      - echo "deploy backend"

````
> as mentioned before, here is a example of multiple lines usage what is needed for the loop.


now you can run this by `contxt run build-backend,build-frontend deploy`

````
contxt run build-backend,build-frontend deploy
[exec:async] ctx-examples/example1/.contxt.yml
[exec:async] ctx-examples/example1/.contxt.yml
 build-fronte   enter Frontend 
 build-backen   enter Backend 
 build-fronte   frontend ----> create <--- 
 build-backen   backend doing .... [git-update]  
 build-fronte   frontend ----> wait <--- 
 build-backen   backend doing .... [run-maven]  
 build-backen   backend doing .... [run-test]  
 build-fronte   frontend ----> finalize <--- 
 build-backen   backend doing .... [build-artifacts]  
 build-fronte   leave Frontend 
 build-backen   backend doing .... [get-coffee]  
 build-backen   backend doing .... [update]  
 build-backen   leave Backend 
[done] build-backend,build-frontend
[exec:async] deploy /home/tziegler/code/playground/go/ctx-examples/example1/.contxt.yml
     deploy :   deploy frontend 
     deploy :   deploy backend 
[done] deploy

````


