package runner_test

import (
	"testing"

	"github.com/swaros/contxt/module/configure"
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

func SetupTestApp() (*runner.CmdSession, error) {
	configure.USE_SPECIAL_DIR = false
	configure.CONTEXT_DIR = "testdata/workspace0/config"
	configure.CONTXT_FILE = "ctx_test_config.yml"
	configure.MIGRATION_ENABLED = false

	app := runner.NewCmdSession()

	functions := runner.NewCmd(app)

	ctxout.AddPostFilter(ctxout.NewTabOut())

	if err := app.Cobra.Init(functions); err != nil {
		return nil, err
	}

	return app, nil
}

func TestDir(t *testing.T) {
	app, _ := SetupTestApp()

	app.Cobra.RootCmd.SetArgs([]string{"dir"})
	if err := app.Cobra.RootCmd.Execute(); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
}
