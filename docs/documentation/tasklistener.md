# listener

Listener contains a List of Triggers they watch for the output of the script and starts other task depending if
some keywords are found in the output.

````yaml
task:
  - id: main
    script:
      - echo 'start A'
      - echo 'start B'
    listener:
      - trigger:
          onoutContains:
            - start A
        action:
          target: task_a

  - id: task_a
    script:
      - echo 'task a is running'
````

with this you can react to something that happens while
execution. 


## Usecase Example
A example would be a local development that uses minikube
to work with kubernetes. 

````yaml
task:
  - id: show_pods
    script:
      - kubectl get pods
````

the task **show_pods** requires minikube is already started. If not you will get a error message.

you have to react to this message and start minikube.
**contxt** can do this too. in different ways.

### the "human reaction" solution

we will add a task, that runs minikube and if we run in a error from *kubectl* that tells us " *Unable to connect* to the server ..." we will trigger the task.

> we use the option "ignoreCmdError: true" so the execution will not aborted due to the error reported from kubectl

afterwards the show pods command is executed again.

this is how we would react in this case. 

````yaml
task:
  # just to start minikube
  - id: run-minikube
    script:    
      - minikube start
  # if we done...rerun the task    
    next:
      - show-pods
  
  # show pods
  - id: show-pods
    options:
      # ignore if the command is failing
      # and continue execution
      ignoreCmdError: true
    script:
      - kubectl get pods
    # here we define what happens if
    # we found a error message
    listener:
      - trigger:
          onoutContains:
            - "Unable to connect"
        action:
          target: run-minikube
````
this is just a example how trigger can be used to did some "fixes" on the fly.
this example will work. with a couple of downsides.

 - the execution of kubectl without running minikube takes seconds to realize they is no service running
 - **run-minkube** is *hard coded* to launch **show-pods**. the whole construct is about one task and not reusable for other tasks

### the solution

to fix that downsides we used the **needs** feature and change the behavior of the script.

instead acting like a human, that run in a error and have to react to them ... and then rerun the task again, we add
a task that checks if minikube is running, and if not, it starts.

then we define that **show-pods** depends on **check-minikube**

````yaml
task:
  # this task is responsible to make sure
  # minikube is running
  - id: check-minikube
    options:
      # important or we just stopping at all
      # because minikube status will returning
      # a error if it not runs
      ignoreCmdError: true
    script:
      - minikube status

    # we just looking if we get the 
    # status "host: Stopped"
    listener:
      - trigger:
          onoutContains:
            - "host: Stopped"
        action:
          # use script instead of target.
          # by using target we would break          
          script:
            - minikube start

  - id: show-pods
    # before we run the script, lets check
    # if minikube is started
    needs:
      - check-minikube
    script:
      - kubectl get pods
    
````

> for task they are used as need it is recommended to use a script action instead of target. A script action is assigned to the current task and target will be get out of scope instead.
