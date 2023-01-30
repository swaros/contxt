package runner

import (
	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/ctxout"
)

func Init() error {
	app := NewCmdSession()
	app.Log.Logger.SetLevel(logrus.ErrorLevel)
	functions := NewCmd(app)

	ctxout.AddPostFilter(ctxout.NewTabOut())

	if err := app.Cobra.Init(functions); err != nil {
		return err
	}

	if err := app.Cobra.RootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
