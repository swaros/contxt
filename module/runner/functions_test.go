package runner_test

import (
	"os"
	"strings"
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/runner"
	"github.com/swaros/contxt/module/systools"
)

var useLastDir = "./"
var lastExistCode = 0

// quicktesting the app messagehandler
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

// Setup the test app
// create the application. set up the config folder name, and the name of the config file.
// the testapp bevavior is afterwards different, because it uses the config
// related to the current directory.
// thats why we have some special helper functions.
// - getAbsolutePath to get the absolute path to the testdata directory
// - backToWorkDir to go back to the testdata directory
// - cleanAllFiles to remove the config file
func SetupTestApp(dir, file string) (*runner.CmdSession, *TestOutHandler, error) {
	// first we want to catch the exist codes
	systools.AddExitListener("testing_prevent_exit", func(no int) systools.ExitBehavior {
		lastExistCode = no
		return systools.Interrupt
	})

	configure.USE_SPECIAL_DIR = false   // no special directory like userHome etc.
	configure.CONTXT_FILE = file        // set the configuration file name
	configure.MIGRATION_ENABLED = false // disable the migration
	configure.CONTEXT_DIR = dir         // set the directory name
	// we need to stick to the testdata directory
	// any other directory will not work
	if err := os.Chdir("testdata"); err != nil {
		return nil, nil, err
	}
	// check if the directory exists, that we want to use in the testdata directory.
	// even if the config package is abel to create them, we want avoid this here.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil, err
	}

	// build the absolute path to the testdata directory
	// this is needed to go back to the testdata directory
	// if needed
	if pwd, derr := os.Getwd(); derr == nil {
		useLastDir = pwd
		configure.CONFIG_PATH_CALLBACK = func() string {
			return useLastDir + "/" + configure.CONTEXT_DIR + "/" + configure.CONTXT_FILE
		}
	} else {
		return nil, nil, derr
	}

	app := runner.NewCmdSession()

	functions := runner.NewCmd(app)

	ctxout.AddPostFilter(ctxout.NewTabOut())

	if err := app.Cobra.Init(functions); err != nil {
		return nil, nil, err
	}

	outputHdnl := NewTestOutHandler()
	app.OutPutHdnl = outputHdnl
	return app, outputHdnl, nil
}

// helper function to change back to the testdata directory
func backToWorkDir() {
	if err := os.Chdir(useLastDir); err != nil {
		panic(err)
	}
}

// helper function to get the absolute path to the testdata directory
func getAbsolutePath(dir string) string {
	return useLastDir + "/" + dir
}

// helper function to remove the config file
func cleanAllFiles() {
	if err := os.Remove(useLastDir + "/" + configure.CONTEXT_DIR + "/" + configure.CONTXT_FILE); err != nil {
		panic(err)
	}
}

// helper function to run a cobra command by argument line
func runCobraCmd(app *runner.CmdSession, cmd string) error {
	app.Cobra.RootCmd.SetArgs(strings.Split(cmd, " "))
	return app.Cobra.RootCmd.Execute()
}

// assert a string is part of the output buffer
func assertInMessage(t *testing.T, output *TestOutHandler, msg string) {
	if !output.Contains(msg) {
		t.Errorf("Expected '%s', got '%v'", msg, output.String())
	}
}

// assert a string is not part of the output buffer
func assertNotInMessage(t *testing.T, output *TestOutHandler, msg string) {
	if output.Contains(msg) {
		t.Errorf("Expected '%s' is not in the message, but got '%v'", msg, output.String())
	}
}

// Testing the dir command togehther with the workspace command
func TestDir(t *testing.T) {
	backToWorkDir()
	app, output, appErr := SetupTestApp("config", "ctx_test_config.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}

	defer cleanAllFiles()
	// clean the output buffer
	output.Clear()
	// just for sure, we go back to the testdata directory
	backToWorkDir()

	// we do not have any workspace, so we expect an hint to create one
	if err := runCobraCmd(app, "dir"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	expected := "no workspace found, nothing to do. create a new workspace with 'ctx workspace new <name>'"
	assertInMessage(t, output, expected)

	// create a new workspace named test
	output.Clear()
	if err := runCobraCmd(app, "workspace new test"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "workspace created test")

	// list all workspaces. we should get the test workspace
	output.Clear()
	if err := runCobraCmd(app, "workspace ls"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "test")

	output.Clear()

	// add an existing directory to the workspace without a absolute path
	if err := runCobraCmd(app, "dir add project1"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "error: path is not absolute")

	// add two directories to the workspace
	diradds := []string{"project1", "project2"}
	for _, diradd := range diradds {

		output.Clear()
		projectAbsPath := getAbsolutePath("workspace0/" + diradd)
		if err := runCobraCmd(app, "dir add "+projectAbsPath); err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
		assertInMessage(t, output, "add "+projectAbsPath)
	}
	output.Clear()

	// list all directories in the workspace
	if err := runCobraCmd(app, "dir list"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, getAbsolutePath("workspace0/project1"))
	assertInMessage(t, output, getAbsolutePath("workspace0/project2"))
	output.Clear()

	// remove the first directory
	if err := runCobraCmd(app, "dir rm "+getAbsolutePath("workspace0/project1")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "remove "+getAbsolutePath("workspace0/project1"))
	output.Clear()

	// list all directories in the workspace after removing the first one
	if err := runCobraCmd(app, "dir list"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertNotInMessage(t, output, getAbsolutePath("workspace0/project1"))
	assertInMessage(t, output, getAbsolutePath("workspace0/project2"))
	output.Clear()

	// remove the second directory
	if err := runCobraCmd(app, "dir rm "+getAbsolutePath("workspace0/project2")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "remove "+getAbsolutePath("workspace0/project2"))
	output.Clear()

	// list all directories in the workspace after removing the second one
	// so booth should be gone
	if err := runCobraCmd(app, "dir list"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertNotInMessage(t, output, getAbsolutePath("workspace0/project1"))
	assertNotInMessage(t, output, getAbsolutePath("workspace0/project2"))
	output.Clear()

	// retry removing an path that is already removed
	if err := runCobraCmd(app, "dir rm "+getAbsolutePath("workspace0/project2")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "error: could not remove path")
	output.Clear()

	// try to remove the whole workspace. that should not work, because we are in the workspace
	if err := runCobraCmd(app, "workspace rm test"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	// we should get the exitcode 10 because we tryed to remove the current workspace
	if lastExistCode != 10 {
		t.Errorf("Expected exit code 10, got '%v'", lastExistCode)
	}

}

func TestWorkSpaces(t *testing.T) {
	app, output, appErr := SetupTestApp("config", "ctx_test_workspace.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	defer cleanAllFiles()
	// clean the output buffer
	output.Clear()

	if err := runCobraCmd(app, "workspace new mainproject"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	dirnames := []string{"build", "web", "server", "client", "docs"}
	for _, dirname := range dirnames {
		os.MkdirAll(getAbsolutePath("workspace1/mainproject/"+dirname), 0755)
		runCobraCmd(app, "dir add "+getAbsolutePath("workspace1/mainproject/"+dirname))
	}

	if cdir, derr := os.Getwd(); derr == nil {
		if cdir != useLastDir {
			t.Errorf("Expected '%v', got '%v'", useLastDir, cdir)
		}
	} else {
		t.Errorf("Expected no error, got '%v'", derr)
	}

	if err := runCobraCmd(app, "workspace new subproject"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	if err := runCobraCmd(app, "dir add "+getAbsolutePath("workspace0/project1")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	if err := runCobraCmd(app, "dir add "+getAbsolutePath("workspace0/project2")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	// list all workspaces. we should get the test workspace
	output.Clear()
	if err := runCobraCmd(app, "workspace ls"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "mainproject")
	assertInMessage(t, output, "subproject")

}
