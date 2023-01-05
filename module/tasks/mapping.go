package tasks

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
)

func (t *targetExecuter) out(msg ...interface{}) {
	if t.outputHandler != nil {
		t.outputHandler(msg...)
	}
}

func (t *targetExecuter) getLogger() *logrus.Logger {
	if t.Logger == nil {
		t.Logger = logrus.New()
		t.Logger.SetFormatter(&logrus.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
		t.Logger.SetOutput(os.Stdout)
	}
	return t.Logger
}

func (t *targetExecuter) SetArgs(args map[string]string) {
	t.arguments = args
}

func (t *targetExecuter) setPh(name, value string) {
	if t.phHandler != nil {
		t.phHandler.SetPH(name, value)
	}
}

func (t *targetExecuter) getPh(input string) string {
	if t.phHandler != nil {
		return t.phHandler.GetPH(input)
	}
	return input
}

func (t *targetExecuter) checkReason(reason configure.Trigger, output string, e error) (bool, string) {
	if t.reasonCheck != nil {
		return t.reasonCheck(reason, output, e)
	}
	return false, ""
}

func (t *targetExecuter) checkRequirements(require configure.Require) (bool, string) {
	if t.checkReqs != nil {
		return t.checkReqs(require)
	}
	return false, ""
}

func (t *targetExecuter) GetWatch() *Watchman {
	return t.watch
}
