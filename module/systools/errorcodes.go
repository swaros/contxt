package systools

const (
	ExitOk               = 0   // Everything is fine
	ErrorExitDefault     = 1   // any exit depending an error that do not need being specific
	ErrorInitApp         = 2   // any application error while setting up
	ErrorWhileLoadCfg    = 3   // any error while loading configuration
	ErrorOnConfigImport  = 5   // a import could not handled
	ErrorTemplate        = 6   // template related error. depending reported issues about version, lint, yaml structure
	ErrorTemplateReading = 7   // errors processing the template
	ErrorBySystem        = 10  // errors related to the system, like while change dir or reading a file
	ExitByStopReason     = 101 // ExitByStopReason the process stopped because of a defined reason
	ExitNoCode           = 102 // ExitNoCode means there was no code associated
	ExitCmdError         = 103 // ExitCmdError means the execution of the command fails. a error by the command itself
	ExitByRequirement    = 104 // ExitByRequirement means a requirement was not fulfills
	ExitAlreadyRunning   = 105 // ExitAlreadyRunning means the task is not started, because it is already created
	ExitByNoTargetExists = 106 // none of the targets are matching in requirements
)
