package runner_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/runner"
)

func TestOutPutHandl(t *testing.T) {
	app := runner.NewCmd(runner.NewCmdSession())

	msg := app.MessageToString("hello", "test")
	if msg != "hellotest" {
		t.Errorf("Expected 'hellotest', got '%v'", msg)
	}
	// add the tabout filter
	ctxout.AddPostFilter(ctxout.NewTabOut())

	msg = app.MessageToString(ctxout.ForeWhite, "current directory: ", ctxout.BoldTag, "/home/itsme", ctxout.CleanTag)
	if msg != "current directory: /home/itsme" {
		t.Errorf("Expected 'hellotest', got '%v'", msg)
	}

}
