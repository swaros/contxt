# Task
tasks are a collection of scripts they are depending to the current directory. this tasks are defined in a taskfile named **.contxt.yml**.  
#### create  a new task
with `contxt create` you can create a simple task file.
this will looks like this
````yaml
task:
  - id: script
    script:
      - echo 'hallo welt'
      - ls -ga
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
     script :   hallo welt 
     script :   insgesamt 112 
     script :   drwxrwxr-x. 10 tziegler  4096 16. Nov 08:06 . 
     script :   drwxrwxr-x.  6 tziegler  4096 23. Okt 07:58 .. 
     script :   drwxrwxr-x.  2 tziegler  4096 16. Nov 08:01 bin 
     script :   drwxrwxr-x.  3 tziegler  4096  8. Okt 09:39 cmd 
     script :   -rw-rw-r--.  1 tziegler    90  8. Okt 09:33 config.go 
     script :   drwxrwxr-x.  7 tziegler  4096 14. Nov 12:48 context 
     script :   -rw-rw-r--.  1 tziegler   142 21. Okt 12:16 .contxt.yml 
     script :   drwxrwxr-x.  5 tziegler  4096 16. Nov 08:06 docs 
     script :   drwxrwxr-x.  8 tziegler  4096 14. Nov 12:49 .git 
     script :   drwxrwxr-x.  3 tziegler  4096  8. Okt 10:33 .github 
     script :   -rw-rw-r--.  1 tziegler   273  8. Okt 10:21 .gitignore 
     script :   -rw-rw-r--.  1 tziegler   310 14. Nov 12:49 go.mod 
     script :   -rw-rw-r--.  1 tziegler 33032 14. Nov 12:49 go.sum 
     script :   drwxrwxr-x.  2 tziegler  4096  8. Okt 09:29 internal 
     script :   -rw-rw-r--.  1 tziegler  1071  8. Okt 09:16 LICENSE 
     script :   -rw-rw-r--.  1 tziegler   747 14. Nov 12:48 Makefile 
     script :   -rw-rw-r--.  1 tziegler  3246 14. Nov 12:49 README.md 
     script :   -rw-rw-r--.  1 tziegler  4841 14. Nov 12:48 TODO.md 
     script :   drwxrwxr-x.  2 tziegler  4096 14. Nov 12:49 .vscode 
[done] script
````