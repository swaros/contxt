# Changelog
All notable changes to this project will be documented in this file.
## [0.1.3] - 2021-12-19
### [Added]
- **os-version-switch** added support for different os versions. tested on linux, windows. `CTX_OS` contains the os like *linux* or *windows* depending on the system. (runtime.GOOS is used, so no hardcoded enum). 
-  **#@if-os** added a contxt if statetment that will accept following lines in a script section only, if the operation system matches. til `#@end`
-  **#@if-equals** added a if statetement for scrit sections, that will check if 2 variables ar equals. til `#@end`
-  **require variables/environments** added checkfor greater, lower equals and not equals. `VARNAME : ">500"` 
-  **require system** added require check for the operation system
-  **CTX_TARGET** variable added for the current executed Target.
### [changes]
-  **https://github.com/imdario/mergo** used instead of own implementation for merging maps.
## [0.1.2] - 2021-07-08
### [Added]
-  **shared task** storing task in the users contxt folder they can be used by the new **use** statement.it can be used in any .contxt.yml taskfile. afterwards any targets from this shared context can used in the run command.
-  **gitlab based shared task** shared task can be fetched from gitlab by the nameing of the task. the content will be fetched by using git. 
-  **shared command** the *shared* command is added to list all shared task and update gitlab based task.
-  **install bash|zsh|fish** *install* command, to create all needed shell related needs. includes the completion setup.
-  **var command** variables can be set by -v name=value. this will overwrite existing variables in the main *variables* section. redefines of variables in targets is not affected.
-  **dir find** the find command, assigned to the dir command, will be look at the path and search for a matching part, that is set by the find param. the last hit will be used. also a index number can be used like the index commmand. 

## [0.1.1] - 2021-08-07
### [Added]
-  **lint** lint errors will be shown at the button of the view, so the user do not have to find them in the diff.
-  **autocompletion support** autocompletion support added for bash, zsh, fish
-  **switch** switching workspaces. supports autocompletion and is used by the ctx shell function to switch also the last used path
### [changes]
-  **runall flag** renamed misleading flag *all-workspaces* into *all-paths*.
-  **variables for require** all require options can be used together with variables
-  **variables** main variables will be set only if they not already defined
### [fixes]
-  **need timeouts** fix behavior of timeouts for needs. timeout check was still running even if all needs was executed successfully.
  
## [0.0.9] - 2020-12-18
### [Added]
-  **script** support for Trigger action besides **target**. this is important together with needs to not make sure to get not out of scope.
-  **needs** checks if a target was already started. this task will be started automatically if the was not started already
- **timeout option** `timeoutNeeds:` task-option for needs. defines the time in milliseconds the task is waiting for the needs. default will be 5 minutes
- **tick time** `tickTimeNeeds:` task-option for needs, defines the time in milliseconds contxt is waiting til the next check is running. default is one second. 
- **runTargets** defines a list of targets they will be started together
- new listener **now** added they are always triggered.
- changelog added

### [Changes]
- rewrite of listener Watcher to make it reusable. Listener will also be executed if no script is assigned.