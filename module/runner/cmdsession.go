package runner

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/ctemplate"
	"github.com/swaros/contxt/module/ctxout"
)

type CmdSession struct {
	Log          *SessionLogger
	TemplateHndl *ctemplate.Template
	Cobra        *SessionCobra
	OutPutHdnl   ctxout.PrintInterface
}

type SessionLogger struct {
	LogLevel string
	Logger   *logrus.Logger
}

func NewCmdSession() *CmdSession {
	return &CmdSession{
		OutPutHdnl:   ctxout.NewMOWrap(),
		Cobra:        NewCobraCmds(),
		TemplateHndl: ctemplate.New(),
		Log: &SessionLogger{
			LogLevel: "info",
			Logger: &logrus.Logger{
				Out:       os.Stdout,
				Formatter: new(logrus.TextFormatter),
				Hooks:     make(logrus.LevelHooks),
				Level:     logrus.ErrorLevel,
			},
		},
	}
}
