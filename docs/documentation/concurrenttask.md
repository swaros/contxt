# runTargets, Needs

you could define the order of task execution, and if they run simultaneously or not, just by the kind and order of the target parmeters. As described here [here](https://github.com/swaros/contxt/blob/wip/shared/docs/documentation/taskdetailed.md#simultaneously-execution)

but you can also define this behavior in the task file itself.
this provides more controll about dependencies.

## runTargets

this option defines targets they should start at the same time, together with the task they is called.

````yaml
task:
  - id: run-all
    runTargets:
       - task1
       - task2

  - id: task1
    script:
      - echo "doing task 1"
  - id: task2
    script:
      - echo "doing task 2"

````
the target task **runAll** will now also trigger the execution of 
**task1** and **task2**. they will run at the same time

so a usecase would be `contxt run run-all`

## Needs

Needs defines other Task in a task they have to be executed successfully before the task can be started.

It doesn't matter how or when this task was performed.
Different to **runtargets** the targets they defined in
**Needs** will started if they was not running already. 

````yaml
task:
  - id: prep_a
    script:
      - echo 'doing some preparation'
  - id: prep_b
    script:
      - echo 'doing some preparation too'

  - id: main
    script:
      - echo 'i need prep_a and prep_b'
    needs:
      - prep_a
      - prep_b
````

if you run the target **main** ( `contxt run main` ) the targets **prep_a** and 
**prep_b** will be run first.
