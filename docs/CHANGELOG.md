# Changelog
All notable changes to this project will be documented in this file.

## [unreleased]
### [Added]
-  **script** support for Trigger action besides **target**. this is important together with needs to not make sure to get not out of scope.
-  **needs** checks if a target was already started. this task will be started automatically if the was not started already
- **timeout option** `timeoutNeeds:` task-option for needs. defines the time in milliseconds the task is waiting for the needs. default will be 5 minutes
- **tick time** `tickTimeNeeds:` task-option for needs, defines the time in milliseconds contxt is waiting til the next check is running. default is one second. 
- **runTargets** defines a list of targets they will be started together
- new listener **now** added they are always triggered.
- changelog added

## [Changes]
- rewrite of listener Watcher to make it reusable. Listener will also be executed if no script is assigned.