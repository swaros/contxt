package runner_test

import (
	"os"
	"testing"
	"time"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/runner"
)

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

// Testing the dir command togehther with the workspace command
func TestDir(t *testing.T) {
	popdTestDir()
	app, output, appErr := SetupTestApp("config", "ctx_test_config.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	cleanAllFiles()
	defer cleanAllFiles()
	// clean the output buffer
	output.Clear()
	// just for sure, we go back to the testdata directory
	backToWorkDir()

	// set the log file with an timestamp
	logFileName := "TestDir_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))

	// we do not have any workspace, so we expect an hint to create one
	if err := runCobraCmd(app, "dir"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	expected := "no workspace found, nothing to do. create a new workspace with 'ctx workspace new <name>'"
	assertInMessage(t, output, expected)

	// create a new workspace named test
	output.ClearAndLog()
	if err := runCobraCmd(app, "workspace new test"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "workspace created test")

	// list all workspaces. we should get the test workspace
	output.ClearAndLog()
	if err := runCobraCmd(app, "workspace ls"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "test")

	output.ClearAndLog()

	// add an existing directory to the workspace without a absolute path
	if err := runCobraCmd(app, "dir add project1"); err == nil {
		t.Errorf("Expected an error, got '%v'", err)
	}
	assertInMessage(t, output, "error: path is not absolute")

	// add two directories to the workspace
	diradds := []string{"project1", "project2"}
	for _, diradd := range diradds {

		output.ClearAndLog()
		projectAbsPath := getAbsolutePath("workspace0/" + diradd)
		if err := runCobraCmd(app, "dir add "+projectAbsPath); err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
		assertInMessage(t, output, "add "+projectAbsPath)
	}
	output.ClearAndLog()

	// list all directories in the workspace
	if err := runCobraCmd(app, "dir list"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, getAbsolutePath("workspace0/project1"))
	assertInMessage(t, output, getAbsolutePath("workspace0/project2"))
	output.ClearAndLog()

	// remove the first directory
	if err := runCobraCmd(app, "dir rm "+getAbsolutePath("workspace0/project1")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "remove "+getAbsolutePath("workspace0/project1"))
	output.ClearAndLog()

	// list all directories in the workspace after removing the first one
	if err := runCobraCmd(app, "dir list"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertNotInMessage(t, output, getAbsolutePath("workspace0/project1"))
	assertInMessage(t, output, getAbsolutePath("workspace0/project2"))
	output.ClearAndLog()

	// remove the second directory
	if err := runCobraCmd(app, "dir rm "+getAbsolutePath("workspace0/project2")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "remove "+getAbsolutePath("workspace0/project2"))
	output.ClearAndLog()

	// list all directories in the workspace after removing the second one
	// so booth should be gone
	if err := runCobraCmd(app, "dir list"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertNotInMessage(t, output, getAbsolutePath("workspace0/project1"))
	assertNotInMessage(t, output, getAbsolutePath("workspace0/project2"))
	output.ClearAndLog()

	// retry removing an path that is already removed
	if err := runCobraCmd(app, "dir rm "+getAbsolutePath("workspace0/project2")); err == nil {
		t.Errorf("Expected an error, because the path is already removed, got none ")
	}
	assertInMessage(t, output, "error: could not remove path")
	output.ClearAndLog()

	// try to remove the whole workspace. that should not work, because we are in the workspace
	if err := runCobraCmd(app, "workspace rm test"); err == nil {
		t.Errorf("we expected an error, but got none")
	}
}

func TestWorkSpaces(t *testing.T) {
	app, output, appErr := SetupTestApp("config", "ctx_test_workspace.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	defer cleanAllFiles()
	defer output.ClearAndLog()
	// clean the output buffer
	output.Clear()
	logFileName := "TestWorkSpaces_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))
	if err := runCobraCmd(app, "workspace new mainproject"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	// create the folders and add them to the workspace
	dirnames := []string{"build", "web", "server", "client", "docs"}
	for _, dirname := range dirnames {
		os.MkdirAll(getAbsolutePath("workspace1/mainproject/"+dirname), 0755)
		runCobraCmd(app, "dir add "+getAbsolutePath("workspace1/mainproject/"+dirname))
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

	if err := runCobraCmd(app, "workspace new testproject"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	dirnames = []string{"website", "backend", "testing"}
	for _, dirname := range dirnames {
		os.MkdirAll(getAbsolutePath("workspace1/testproject/"+dirname), 0755)
		if err := runCobraCmd(app, "dir add "+getAbsolutePath("workspace1/testproject/"+dirname)); err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
	}
	output.Clear()
	// try to add a directory that is already added
	if err := runCobraCmd(app, "dir add "+getAbsolutePath("workspace1/testproject/testing")); err == nil {
		t.Error("Expected an error, got none")
	}

	output.Clear()
	// try to add a directory that is not exists
	if err := runCobraCmd(app, "dir add "+getAbsolutePath("workspace1/testproject/abc")); err == nil {
		t.Error("Expected an error, got none")
	}

	output.ClearAndLog()
	if err := runCobraCmd(app, "workspace ls"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "mainproject")
	assertInMessage(t, output, "subproject")
	assertInMessage(t, output, "testproject")

	output.ClearAndLog()
	if err := runCobraCmd(app, "dir -a"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "mainproject: index (0)")
	assertInMessage(t, output, "subproject: index (0)")
	assertInMessage(t, output, "testproject: index (0)")
	output.ClearAndLog()

	if err := runCobraCmd(app, "switch lalaland"); err == nil {
		t.Error("Expected an error, got none")
	}

	output.ClearAndLog()

	if err := runCobraCmd(app, "switch testproject"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	output.ClearAndLog()
	if err := runCobraCmd(app, "workspace current"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "testproject")
	assertNotInMessage(t, output, "mainproject")
	assertNotInMessage(t, output, "subproject")
	output.ClearAndLog()

	if err := runCobraCmd(app, "workspace rm subproject"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	output.ClearAndLog()
	if err := runCobraCmd(app, "dir -a"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "mainproject: index (0)")
	assertNotInMessage(t, output, "subproject")
	assertInMessage(t, output, "testproject: index (0)")
	output.ClearAndLog()

	// behavior if paths get removed
	os.RemoveAll(getAbsolutePath("workspace1/mainproject/docs"))
	output.ClearAndLog()
	if err := runCobraCmd(app, "dir -a"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertNotInMessage(t, output, "docs")
}

func TestWorkSpacesInvalidNames(t *testing.T) {

	app, output, appErr := SetupTestApp("config", "ctx_test_ws_naming.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	defer cleanAllFiles()
	defer output.ClearAndLog()
	// clean the output buffer
	output.Clear()
	logFileName := "ws_names_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))
	assertCobraError(t, app, "workspace new", "no workspace name given")
	assertCobraError(t, app, "workspace new 1", "the workspace name [1] too short")

	toLongString := "asbfdufkif"
	for i := 0; i < 13; i++ {
		toLongString += toLongString
	}
	assertCobraError(t, app, "workspace new "+toLongString, " too long")
	output.ClearAndLog()
	assertCobraError(t, app, "workspace new ^^..", "string contains not accepted chars")
	output.ClearAndLog()
	assertCobraError(t, app, "workspace new \"hello\"", "string contains not accepted chars")
	output.ClearAndLog()
	assertCobraError(t, app, "workspace new 'hello'", "string contains not accepted chars")
	output.ClearAndLog()
	assertCobraError(t, app, "workspace new hello/", "the workspace name [hello/] is invalid")
	output.ClearAndLog()
	assertCobraError(t, app, "workspace new hello\\", "the workspace name [hello\\] is invalid")
	output.ClearAndLog()
	assertCobraError(t, app, "workspace new hello:", "the workspace name [hello:] is invalid")
	output.ClearAndLog()
	assertCobraError(t, app, "workspace new hello?", "string contains not accepted chars")
	output.ClearAndLog()
	assertCobraError(t, app, "workspace new hello*", "string contains not accepted chars")
	output.ClearAndLog()
	assertCobraError(t, app, "workspace new hello<", "string contains not accepted chars")
	output.ClearAndLog()
	assertCobraError(t, app, "workspace new hello>", "string contains not accepted chars")
	output.ClearAndLog()
	assertCobraError(t, app, "workspace new hello world", "to many arguments")
}
