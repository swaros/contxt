package runner

import (
	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
)

type CmdExecutor interface {
	PrintPaths(plain bool)
	GetLogger() *logrus.Logger
	GetOuputHandler() ctxout.PrintInterface
	SetLogLevel(level string) error
	ResetVariables()
	MainInit()
	doMagicParamOne(string)
	RunTargets(string, bool)
	CallBackNewWs(string)
	CallBackOldWs(string) bool
	FindWorkspaceInfoByTemplate(updateFn func(workspace string, cnt int, update bool, info configure.WorkspaceInfoV2)) (allCount int, updatedCount int)
}
