package runner

import (
	"github.com/sirupsen/logrus"
)

func Init() error {
	app := NewCmdSession()
	app.Log.Logger.SetLevel(logrus.ErrorLevel)
	functions := NewCmd(app)

	if err := app.Cobra.Init(functions); err != nil {
		return err
	}

	if err := app.Cobra.RootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
