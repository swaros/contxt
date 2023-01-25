package runner

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/ctemplate"
)

type CmdSession struct {
	Log          *SessionLogger
	TemplateHndl *ctemplate.Template
	Cobra        *SessionCobra
}

type SessionLogger struct {
	LogLevel string
	Logger   *logrus.Logger
}

func NewCmdSession() *CmdSession {
	return &CmdSession{
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
