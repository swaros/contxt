# Listener, Needs, runTargets

you could define the order of task execution, and if they run simultaneously or not, just by the kind and order of the target parmeters. As described here [here](https://github.com/swaros/contxt/blob/wip/shared/docs/documentation/taskdetailed.md#simultaneously-execution)

but you can also define this behavior in the task file itself.
this provides more controll about dependencies.

## runTargets

this option defines targets they should start at the same time, together with the task they is called.

````yaml
task:
  - id: run-all
    runTargets:
       - build-backend
       - build-frontend

# the 'real' tasks
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
the task **runAll** will now also trigger the execution of 
**build-backend** and **build-frontend** afterwards.

so a usecase would be `contxt run run-all deploy`
