package runner_test

import (
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/runner"
	"github.com/swaros/contxt/module/systools"
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

	verifyConfigurationFile(t, func(CFG *configure.ConfigMetaV2) {
		if CFG.CurrentSet != "subproject" {
			t.Errorf("Expected the current set to be 'subproject', got '%v'", CFG.CurrentSet)
		}
		// there should no paths in the subproject at this point
		if len(CFG.Configs["subproject"].Paths) != 0 {
			t.Errorf("Expected no paths in the subproject, got '%v'", CFG.Configs["subproject"].Paths)
		}
	})

	if err := runCobraCmd(app, "dir add "+getAbsolutePath("workspace0/project1")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	verifyConfigurationFile(t, func(CFG *configure.ConfigMetaV2) {
		if len(CFG.Configs["subproject"].Paths) != 1 {
			t.Errorf("Expected one path in the subproject, got '%v'", CFG.Configs["subproject"].Paths)
		} else {
			if CFG.Configs["subproject"].Paths["0"].Path != getAbsolutePath("workspace0/project1") {
				t.Errorf("Expected the path to be '%v', got '%v'", getAbsolutePath("workspace0/project1"), CFG.Configs["subproject"].Paths["0"].Path)
			}
		}
	})

	if err := runCobraCmd(app, "dir add "+getAbsolutePath("workspace0/project2")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	verifyConfigurationFile(t, func(CFG *configure.ConfigMetaV2) {
		if len(CFG.Configs["subproject"].Paths) != 2 {
			t.Errorf("Expected two paths in the subproject, got '%v'", CFG.Configs["subproject"].Paths)
		} else {
			if CFG.Configs["subproject"].Paths["1"].Path != getAbsolutePath("workspace0/project2") {
				t.Errorf("Expected the path to be '%v', got '%v'", getAbsolutePath("workspace0/project2"), CFG.Configs["subproject"].Paths["1"].Path)
			}
		}
	})

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

	verifyConfigurationFile(t, func(CFG *configure.ConfigMetaV2) {
		if CFG.CurrentSet != "testproject" {
			t.Errorf("Expected the current set to be 'testproject', got '%v'", CFG.CurrentSet)
		}
		// there should no paths in the subproject at this point
		if len(CFG.Configs["testproject"].Paths) != 0 {
			t.Errorf("Expected no paths in the subproject, got '%v'", CFG.Configs["testproject"].Paths)
		}
	})

	dirnames = []string{"website", "backend", "testing"}
	for _, dirname := range dirnames {
		os.MkdirAll(getAbsolutePath("workspace1/testproject/"+dirname), 0755)
		if err := runCobraCmd(app, "dir add "+getAbsolutePath("workspace1/testproject/"+dirname)); err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
	}
	output.Clear()

	verifyConfigurationFile(t, func(CFG *configure.ConfigMetaV2) {
		if len(CFG.Configs["testproject"].Paths) != 3 {
			t.Errorf("Expected three paths in the testproject, got '%v'", CFG.Configs["testproject"].Paths)
		} else {
			if CFG.Configs["testproject"].Paths["2"].Path != getAbsolutePath("workspace1/testproject/testing") {
				t.Errorf("Expected the path to be '%v', got '%v'", getAbsolutePath("workspace1/testproject/testing"), CFG.Configs["testproject"].Paths["2"].Path)
			}
		}
	})

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
	assertInMessage(t, output, "docs: no such file or directory")
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

func TestRunBasic(t *testing.T) {
	app, output, appErr := SetupTestApp("tasks1", "ctx_test_basic.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	defer cleanAllFiles()
	defer output.ClearAndLog()
	// clean the output buffer
	output.Clear()
	logFileName := "basic_run_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))
	if err := runCobraCmd(app, "run test1"); err == nil {
		t.Error("Expected an error, got none")
	} else {
		expectedError := "no contxt template found in current directory"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error '%v', got '%v'", expectedError, err)
		}
	}

}

func TestRunBasic2(t *testing.T) {
	app, output, appErr := SetupTestApp("task2", "ctx_test_basic.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	defer cleanAllFiles()
	defer output.ClearAndLog()
	// clean the output buffer
	output.Clear()
	logFileName := "basic2_run_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))

	// change into the test directory
	if err := os.Chdir(getAbsolutePath("task2")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	if err := runCobraCmd(app, "run test1"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	assertInMessage(t, output, "testing-1-working")
	assertInMessage(t, output, "test1 DONE")
	assertNotInMessage(t, output, "testing-2-working")
	output.ClearAndLog()

	if err := runCobraCmd(app, "run test2"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	assertInMessage(t, output, "testing-2-working")
	assertInMessage(t, output, "test2 DONE")
	assertNotInMessage(t, output, "testing-1-working")
	output.ClearAndLog()

	if err := runCobraCmd(app, "run not-exists"); err == nil {
		t.Error("Expected an error, got none")
	} else {
		expectedError := "target not-exists not exists"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error '%v', got '%v'", expectedError, err)
		}
	}
}

// testing the ctx_pwd default variable
// and a defiened variable in the context file
func TestRunAndVariables(t *testing.T) {
	app, output, appErr := SetupTestApp("task3", "ctx_test_basic.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	defer cleanAllFiles()
	defer output.ClearAndLog()
	// clean the output buffer
	output.Clear()
	logFileName := "basic3_run_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))

	// change into the test directory
	if err := os.Chdir(getAbsolutePath("task3")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	if err := runCobraCmd(app, "run show-variables"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	expectedPath := getAbsolutePath("task3")

	assertInMessage(t, output, expectedPath)

	expectedHello := "hello john doe"
	assertInMessage(t, output, expectedHello)
	output.ClearAndLog()

}

func TestRunAndVariablesFromProjects(t *testing.T) {
	app, output, appErr := SetupTestApp("projects01", time.Now().Format(time.RFC3339)+"ctx_projects.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	defer cleanAllFiles()
	defer output.ClearAndLog()
	// clean the output buffer
	output.Clear()
	logFileName := "samira_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))

	// change into the test directory
	if err := os.Chdir(getAbsolutePath("projects01/project_samira")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	// first create a new workspace
	// we do ot care about errors here
	// because at first run the workspace might does not exists
	if err := runCobraCmd(app, "workspace new samira"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	if err := runCobraCmd(app, "dir add"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	if err := runCobraCmd(app, "workspace scan"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "found 1 projects and updated 1 projects")
	output.ClearAndLog()
	if err := runCobraCmd(app, "dir"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "projects01/project_samira")
	assertInMessage(t, output, "current workspace: samira")
	output.ClearAndLog()

	// add the website to the project
	if err := os.Chdir(getAbsolutePath("projects01/project_samira/website")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	if err := runCobraCmd(app, "dir add"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	// update workspace should find the new project and updates the workspace
	assertInMessage(t, output, "found 1 projects and updated 1 projects")
	output.ClearAndLog()

	// add the website to the project
	if err := os.Chdir(getAbsolutePath("projects01/project_samira/backend")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	if err := runCobraCmd(app, "dir add"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	if err := os.Chdir(getAbsolutePath("projects01/project_samira")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	// lets do the linting
	if err := runCobraCmd(app, "lint"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "...loading config ok")
	output.ClearAndLog()
	if err := runCobraCmd(app, "run defaults"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	assertInMessage(t, output, "CTX_PWD "+systools.PadString(getAbsolutePath("projects01/project_samira"), 40))
	assertInMessage(t, output, "CTX_PROJECT samira")
	assertInMessage(t, output, "CTX_ROLE root")
	assertInMessage(t, output, "CTX_VERSION 1.0.2")
	// just windows and linux will be checked depending OS
	if runtime.GOOS == "windows" {
		assertInMessage(t, output, "CTX_OS windows")
	} else if runtime.GOOS == "linux" {
		assertInMessage(t, output, "CTX_OS linux")
	}
	assertInMessage(t, output, "CTX_USER "+os.Getenv("USER"))
	assertRegexmatchInMessage(t, output, "CTX_HOST [a-zA-Z0-9.-]+")
	assertRegexmatchInMessage(t, output, "CTX_DATE [0-9]{4}-[0-9]{2}-[0-9]{2}")
	assertRegexmatchInMessage(t, output, "CTX_TIME [0-9]{2}:[0-9]{2}:[0-9]{2}")
	assertRegexmatchInMessage(t, output, "CTX_DATETIME [0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}")

	assertInMessage(t, output, "CTX_TARGET default")
	assertInMessage(t, output, "CTX_WS samira")
	// order of the keys is not defined
	assertRegexmatchInMessage(t, output, "CTX_WS_KEYS WS0_samira_[a-z]+ WS0_samira_[a-z]+ WS0_samira_[a-z]+")
	assertInMessage(t, output, "WS0_samira_root "+systools.PadString(getAbsolutePath("projects01/project_samira"), 40))
	assertInMessage(t, output, "WS0_samira_website "+systools.PadString(getAbsolutePath("projects01/project_samira/website"), 40))
	assertInMessage(t, output, "WS0_samira_backend "+systools.PadString(getAbsolutePath("projects01/project_samira/backend"), 40))

	ducktalePath := getAbsolutePath("projects01/dangerduck")
	// now add a new project
	if err := os.Chdir(ducktalePath); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInOsPath(t, ducktalePath)
	output.ClearAndLog()
	if err := runCobraCmd(app, "workspace new ducktale"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	assertInOsPath(t, ducktalePath)

	if err := runCobraCmd(app, "dir add"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	output.ClearAndLog()
	if err := runCobraCmd(app, "run script"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	// testing if variable map is working
	assertInMessage(t, output, "greeting raideristwix")

}
