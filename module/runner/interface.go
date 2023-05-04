package runner

import (
	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
)

type CmdExecutor interface {
	Print(msg ...interface{})
	Println(msg ...interface{})
	PrintPaths(plain bool, showFulltask bool)                         // print out all paths
	GetLogger() *logrus.Logger                                        // get logger
	GetOuputHandler() (ctxout.StreamInterface, ctxout.PrintInterface) // get output handlers
	SetLogLevel(level string) error                                   // set log level
	ResetVariables()                                                  // reset old variables while change the workspace. (req for shell mode)
	MainInit()                                                        // initialize the workspace
	doMagicParamOne(string)
	RunTargets(string, bool)   // run targets
	CallBackNewWs(string)      // callback for new workspace
	CallBackOldWs(string) bool // callback for old workspace
	FindWorkspaceInfoByTemplate(updateFn func(workspace string, cnt int, update bool, info configure.WorkspaceInfoV2)) (allCount int, updatedCount int)
	PrintWorkspaces()                                  // print out all workspaces
	GetWorkspaces() []string                           // print out all workspaces as a list
	DirFindApplyAndSave(args []string) (string, error) // find pathbay arguments,save the current path print the path
	InteractiveScreen()                                // interactive screen
}
