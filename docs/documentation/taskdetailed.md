# Tasks definition

a contxt task is defined in yaml and always named **.contxt.yml**.

## structure

**task** is the root for a list of named tasks (id)

````yaml
task:
  - id: example-1
  - id: example-2
````

if you would execute `contxt run` it shows these targets, even if they
have nothing to do

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
if you run the *example-1* target again, it would looks like so

````
contxt run example-1
  example-1 :   this is example 1 
````

but you can also run boot targets `contxt run target1 target2`
````
contxt run example-1 example-2
  example-1 :   this is example 1 
  example-2 :   this is example 2 
````

in this case the first target *example-1* will be executed and afterwards the second target *example-2*

## Mutliline script
> **important** any single line is a new process. you need to keep this in mind because if you need to change directories, or set variables, you have to put this in one line. use **- |** annotation to write mutliple lines in one script.

even if we use in this documentation single line scripts, you should always prefer multiline scripts.
only there you can use all bash features like loops, conditions and so on.

````yaml
task:
  - id: clean_tmp
    script:
      - |
        pushd /tmp
        rm -rf *
        popd 
````


## simultaneously execution
you could also run this targets paralell. then you need to combine them
as one parameter seperated by comma.
`contxt run example-1,example-2`

to show this is working, a slepp in the example would help.

````yaml
task:
  - id: example-1
    script:
      - | 
       echo "start example 1"
       sleep 5
       echo "done example 1"
  - id: example-2
    script:
      - |
       echo "start example 2"
       sleep 2
       echo "done example 2"
````

now you start booth task at the same time
````
contxt run example-1,example-2
  example-1 :   start example 1 
  example-2 :   start example 2 
  example-2 :   done example 2 
  example-1 :   done example 1 
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




