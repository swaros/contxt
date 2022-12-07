package taskrun

const (
	ERRORCODE_ON_CONFIG_IMPORT = 5   // a import could not handled
	ExitOk                     = 0   // Everything is fine
	ExitByStopReason           = 101 // ExitByStopReason the process stopped because of a defined reason
	ExitNoCode                 = 102 // ExitNoCode means there was no code associated
	ExitCmdError               = 103 // ExitCmdError means the execution of the command fails. a error by the command itself
	ExitByRequirement          = 104 // ExitByRequirement means a requirement was not fulfills
	ExitAlreadyRunning         = 105 // ExitAlreadyRunning means the task is not started, because it is already created
	ExitByNoTargetExists       = 106 // none of the targets are matching in requirements
)
