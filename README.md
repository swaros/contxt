# cont(e)xt

**contxt** manage paths in workspaces while working on a shell.

*example session* switch to workspace and change paths by shortcuts. 
> shortcut is in this example a function in ~/.bashrc `function cn() { cd $(contxt dir -i "$@"); }` 
````shell
user@showcase:~$ contxt projectX
current workspace is now:projectX
user@showcase:~$ context dir --paths
paths stored in deltadna
	 0 	/home/user/devsource/project_x
	 1 	/home/user/devsource/project_x/server
	 2 	/home/user/devsource/project_x/client

	to change directory depending stored path you can write cd $(contxt -i 2) in bash
	this will be the same as cd /home/user/devsource/project_x/client
user@showcase:~$ cn 1
user@showcase:~/devsource/project_x/server$ pwd
/home/user/devsource/project_x/server
user@showcase:~/devsource/project_x/server$ cn 2
user@showcase:~/devsource/project_x/client$ pwd
/home/user/devsource/project_x/client

````

## workspaces
#### create a new workspace
`contxt dir -w mywork` will create a new workspace named *mywork*
#### list existing workspaces
````shell
user@showcase:~$  contxt dir -list
mywork
````
## 	manage paths
### add path to current workspace
`contxt dir -add` adds the current directory to the workspace

example:
````shell
user@showcase:~$ cd /home/user/devsource/project_x/server
user@showcase:~/devsource/project_x/server$ contxt dir -add
add /home/user/devsource/project_x/server
````

### show all paths in current workspace
`contxt dir -paths` 
example:
````shell
user@showcase:~$ context dir --paths
paths stored in deltadna
	 0 	/home/user/devsource/project_x/server
	 
	to change directory depending stored path you can write cd $(contxt -i 0) in bash
	this will be the same as /home/user/devsource/project_x/server
````
