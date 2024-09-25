package runner_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/swaros/contxt/module/configure"
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
	path := RuntimeFileInfo(t)
	t.Log(path)
	ChangeToRuntimeDir(t)
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
	ChangeToRuntimeDir(t)
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
	if runtime.GOOS == "windows" {
		assertInMessage(t, output, "The system cannot find the file specified.")
	} else {
		assertInMessage(t, output, "docs: no such file or directory")
	}
}

func TestWorkSpacesInvalidNames(t *testing.T) {
	ChangeToRuntimeDir(t)
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
	assertInMessage(t, output, "test1")
	assertInMessage(t, output, "DONE")
	assertNotInMessage(t, output, "testing-2-working")
	output.ClearAndLog()

	if err := runCobraCmd(app, "run test2"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	assertInMessage(t, output, "testing-2-working")
	assertInMessage(t, output, "test2")
	assertInMessage(t, output, "DONE")
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

// this test is simualting working with different projects they have different context files
// and different tasks
// so we have to change the directorys, add the projects and different subprojects (paths)
// register an different project and checking again if they are added and handled correctly.
// also we check if the tasks are loaded correctly and can be executed as expected.
// even if they are in a place that is still not added to the project.
// here we are in a special behavior, because the usal way is to run any of these commands
// in his own run. so while executing these commands. the application is initialized fresh.
// here we work with the same instance of the instance, so we can make sure, that the
// application keeps the state of the projects and tasks correctly, even it is not re-initialized
func TestRunAndVariablesFromProjects(t *testing.T) {
	ChangeToRuntimeDir(t)
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
	// especially here to the first projectwe have there.
	if err := os.Chdir(getAbsolutePath("projects01/project_samira")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	// first create a new workspace named samira
	if err := runCobraCmd(app, "workspace new samira"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	// add the current directory to the workspace
	if err := runCobraCmd(app, "dir add"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	// scan the workspace for projects
	if err := runCobraCmd(app, "workspace scan"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "found 1 projects and updated 1 projects")
	output.ClearAndLog()

	// show all projects in the workspace
	if err := runCobraCmd(app, "dir"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, filepath.Clean("projects01/project_samira"))
	assertInMessage(t, output, "current workspace: samira")
	output.ClearAndLog()

	// add the website path to the project.
	//so we change to the website directory first
	if err := os.Chdir(getAbsolutePath("projects01/project_samira/website")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	// and then add them to the samira project
	if err := runCobraCmd(app, "dir add"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	// update workspace should find the new project and updates the workspace
	assertInMessage(t, output, "found 1 projects and updated 1 projects")
	output.ClearAndLog()

	// now we are doing the same for the backend path
	// again by changing first into the directory
	if err := os.Chdir(getAbsolutePath("projects01/project_samira/backend")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	// ...and adding them to the project
	if err := runCobraCmd(app, "dir add"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	// going back into the project directory
	if err := os.Chdir(getAbsolutePath("projects01/project_samira")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	// lets do the linting
	if err := runCobraCmd(app, "lint"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "...loading config ok")
	output.ClearAndLog()
	// now we run the task "run defaults" from the root project directory.
	// this task can be found in projects01/project_samira/.contxt.yml
	// and here is how it looks like:
	/*
		workspace:
			project: samira
			role: root
			version: "1.0.2"

		task:
		- id: defaults
			script:
			- echo "a) CTX_PWD ${CTX_PWD}"
			- echo "b) CTX_PROJECT ${CTX_PROJECT}"
			- echo "c) CTX_ROLE ${CTX_ROLE}"
			- echo "d) CTX_VERSION ${CTX_VERSION}"
			- echo "e) CTX_OS ${CTX_OS}"
			- echo "f) CTX_ARCH ${CTX_ARCH}"
			- echo "g) CTX_USER ${CTX_USER}"
			- echo "h) CTX_HOST ${CTX_HOST}"
			- echo "i) CTX_HOME ${CTX_HOME}"
			- echo "j) CTX_DATE ${CTX_DATE}"
			- echo "k) CTX_TIME ${CTX_TIME}"
			- echo "l) CTX_DATETIME ${CTX_DATETIME}"
			- echo "m) CTX_BUILD_NO ${CTX_BUILD_NO}"
			- echo "n) CTX_TARGET ${CTX_TARGET}"
			- echo "o) CTX_FORCE ${CTX_FORCE}"
			- echo "p) CTX_WS ${CTX_WS}"
			- echo "q) CTX_WS_KEYS ${CTX_WS_KEYS}"
			- echo "r) WS0_samira_root ${WS0_samira_root}"
			- echo "s) WS0_samira_website ${WS0_samira_website}"
			- echo "t) WS0_samira_backend ${WS0_samira_backend}"
	*/
	// the goal is to check if all the variables are set correctly.
	// and can be used as placeholders in the script.
	if err := runCobraCmd(app, "run defaults"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	// check the outpout if we got the variables used as expected.
	assertInMessage(t, output, "CTX_PWD "+getAbsolutePath("projects01/project_samira"))
	assertInMessage(t, output, "CTX_PROJECT samira")
	assertInMessage(t, output, "CTX_ROLE root")
	assertInMessage(t, output, "CTX_VERSION 1.0.2")
	// just windows and linux will be checked depending OS
	if runtime.GOOS == "windows" {
		assertInMessage(t, output, "CTX_OS windows")
	} else if runtime.GOOS == "linux" {
		assertInMessage(t, output, "CTX_OS linux")
	}
	assertInMessage(t, output, "CTX_USER "+os.Getenv("USER"))                                                  // user should be the current OS user
	assertRegexmatchInMessage(t, output, "CTX_HOST [a-zA-Z0-9.-]+")                                            // hostname differs depending on the OS
	assertRegexmatchInMessage(t, output, "CTX_DATE [0-9]{4}-[0-9]{2}-[0-9]{2}")                                // also the date is different
	assertRegexmatchInMessage(t, output, "CTX_TIME [0-9]{2}:[0-9]{2}:[0-9]{2}")                                // also the time is different
	assertRegexmatchInMessage(t, output, "CTX_DATETIME [0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}") // also the datetime is different

	assertInMessage(t, output, "CTX_TARGET default")
	assertInMessage(t, output, "CTX_WS samira")
	// order of the keys is not defined
	assertRegexmatchInMessage(t, output, "CTX_WS_KEYS WS0_samira_[a-z]+ WS0_samira_[a-z]+ WS0_samira_[a-z]+") // here we don't know the order of the keys
	assertInMessage(t, output, "WS0_samira_root "+getAbsolutePath("projects01/project_samira"))
	assertInMessage(t, output, "WS0_samira_website "+getAbsolutePath("projects01/project_samira/website"))
	assertInMessage(t, output, "WS0_samira_backend "+getAbsolutePath("projects01/project_samira/backend"))

	// the durcktale project what is independent from the project_samira
	ducktalePath := getAbsolutePath("projects01/dangerduck")
	// doing the same as usual.
	// - creating a new workspace
	// - adding the directory to the workspace
	// - running the script
	// - checking if the scripts doing the expected stuff

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

	// this script is using the variable map.
	// this time we import a values.yml file, namig it as "values".
	// the script is accessing the variable "ducktale.nightrider" from these variables.
	// so we just need to check if the variable is set correctly.
	// this is how the values.yml looks like:
	/*
		ducktale:
		  nightrider: raideristwix
	*/
	// and this is how the script looks like:
	/*
		config:
		  imports:
		    - values.yml values
		task:
		  - id: script
		    script:
		      - echo "greeting ${values:ducktale.nightrider}"
	*/
	if err := runCobraCmd(app, "run script"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	// testing if variable map is working
	assertInMessage(t, output, "greeting raideristwix")

	// now we are going in to another directory and running the same script again.
	// note: we do not need to add this folder to the workspace, because scripts can run from any directory.
	if err := os.Chdir(ducktalePath + "/varia"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	output.ClearAndLog()

	// this time the script making use of the whole templating features.
	// so there is a additional configuration file ".inc.contxt.yml" in the directory, that defines
	// what directories or files should be read before as values.
	// these files are then used as values for go/template.
	// in this case the relative folder data/ is read and any yaml or json file is used as values.
	// right now this is the users.json file:
	/*
		{
		    "user" : [
		        {
		            "name" : "John",
		            "age" : 30,
		            "cars" : [
		                {
		                    "name" : "Ford",
		                    "models" : ["Fiesta", "Focus", "Mustang"]
		                },
		                {
		                    "name" : "BMW",
		                    "models" : ["320", "X3", "X5"]
		                },
		                {
		                    "name" : "Fiat",
		                    "models" : ["500", "Panda"]
		                }
		            ]
		        },
		        {
		            "name" : "Peter",
		            "age" : 46,
		            "cars" : [
		                {
		                    "name" : "Hundai",
		                    "models" : ["i10", "i20", "i30"]
		                },
		                {
		                    "name" : "Rover",
		                    "models" : ["25", "45", "75"]
		                }
		            ]
		        }
		    ]
		}
	*/
	// this will used as values and then the script is read as template and the values are used to fill the template.
	// this is the origin script:
	/*
		config:
		  imports:
		    - imp/hello.txt letter

		task:
		  - id: testimports
		    script:
		      - echo "start template"
		  {{ range $key, $User := .user }}
		      - echo "User  {{ $User.name }} {{ $User.age }}"
		    {{ range $kn, $Cars := $User.cars }}
		      - echo "  --> {{ $User.name }}'s Car no {{ $kn }} {{ $Cars.name }}"
		      {{ range $km, $Car := $Cars.models }}
		      - echo "       {{ $Cars.name }} {{ $Car }}"
		      {{ end }}
		    {{ end }}
		  {{ end }}

		  - id: letter
		    script:
		      - |
		        cat << EOF
		        ${letter}
		        EOF
	*/
	// after executing the templating step, the "real" script looks like this:
	/*
		config:
		  imports:
		    - imp/hello.txt letter

		task:
		  - id: testimports
		    script:
		      - echo "start template"
		      - echo "start template"
		      - echo "User  John 30"
		      - echo "  --> John's Car no 0 Ford"
		      - echo "       Ford Fiesta"
		      - echo "       Ford Focus"
		      - echo "       Ford Mustang"
		      - echo "  --> John's Car no 1 BMW"
		      - echo "       BMW 320"
		      - echo "       BMW X3"
		      - echo "       BMW X5"
		      - echo "  --> John's Car no 2 Fiat"
		      - echo "       Fiat 500"
		      - echo "       Fiat Panda"
		      - echo "User  Peter 46"
		      - echo "  --> Peter's Car no 0 Hundai"
		      - echo "       Hundai i10"
		      - echo "       Hundai i20"
		      - echo "       Hundai i30"
		      - echo "  --> Peter's Car no 1 Rover"
		      - echo "       Rover 25"
		      - echo "       Rover 45"
		      - echo "       Rover 75"

		  - id: letter
		    script:
		      - |
		        cat << EOF
		        ${letter}
		        EOF
	*/
	// so lets see if this is working as expected.
	if err := runCobraCmd(app, "run testimports"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	expectedStrings := `start template
User  John 30
  --> John's Car no 0 Ford
   Ford Fiesta
   Ford Focus
   Ford Mustang
 --> John's Car no 1 BMW
   BMW 320
   BMW X3
   BMW X5
 --> John's Car no 2 Fiat
   Fiat 500
   Fiat Panda
User  Peter 46
 --> Peter's Car no 0 Hundai
   Hundai i10
   Hundai i20
   Hundai i30
 --> Peter's Car no 1 Rover
   Rover 25
   Rover 45
   Rover 75`
	assertSplitTestInMessage(t, output, expectedStrings)

	// the next task is letter, where we have an import of an file, that is NOT readable as yaml or json.
	// what means this file is not used for values, but is just read as string and used as simple value.
	/*
		config:
			imports:
			 - imp/hello.txt letter
	*/
	// BUT the hello.txt itself have go/template syntax, so the values from the previous task are used as values for the template.
	/*
			{{ range $key, $User := .user }}
		      Hello {{ $User.name }} !

		      we have to talk about your age {{ $User.age }}.

		      as you know you have {{ len $User.cars }} different car models.
		      so we have to talk about them.

		    {{ range $kn, $Cars := $User.cars }}
		      Number: {{ $kn }} is {{ $Cars.name }}" and from them you have
		      {{- range $km, $Car := $Cars.models }}
		            {{ $Cars.name }} {{ $Car }}
		      {{- end }}
		    {{ end }}
		    -------------------------------------------------------------
		{{ end }}
	*/
	// so this file is also used as template file.
	// this way we can use any file as template file, even if it is not readable as yaml or json.
	// for the test we just print the parsed content of the file.
	/*
			  - id: letter
		    script:
		      - |
		        cat << EOF
		        ${letter}
		        EOF
	*/
	output.ClearAndLog()
	if err := runCobraCmd(app, "run letter --loglevel debug"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	expected := `
Hello John ! 

we have to talk about your age 30. 

as you know you have 3 different car models. 
so we have to talk about them. 

Number: 0 is Ford" and from them you have 
  Ford Fiesta 
  Ford Focus 
  Ford Mustang 

Number: 1 is BMW" and from them you have 
  BMW 320 
  BMW X3 
  BMW X5 
   
Number: 2 is Fiat" and from them you have 
  Fiat 500 
  Fiat Panda 
   
  ------------------------------------------------------------- 

Hello Peter ! 

we have to talk about your age 46. 
 
as you know you have 2 different car models. 
so we have to talk about them. 

Number: 0 is Hundai" and from them you have 
  Hundai i10 
  Hundai i20 
  Hundai i30 

Number: 1 is Rover" and from them you have 
  Rover 25 
  Rover 45 
  Rover 75 
`
	assertSplitTestInMessage(t, output, expected)
	// finally we just test the lint command.
	// and inspect also any issues.
	output.ClearAndLog()
	if err := runCobraCmd(app, "lint --show-issues"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	expected = `
Config.AllowMutliRun false MissingEntry: level[5] @allowmultiplerun (Config.AllowMutliRun)bool
Config.Autorun  MissingEntry: level[5] @autorun (Config.Autorun)undefined
Config.Autorun.Onenter  MissingEntry: level[5] @onenter (Config.Autorun.Onenter)string
Config.Autorun.Onleave  MissingEntry: level[5] @onleave (Config.Autorun.Onleave)string
Config.Coloroff false MissingEntry: level[5] @coloroff (Config.Coloroff)bool
Config.Loglevel  MissingEntry: level[5] @loglevel (Config.Loglevel)string
Config.MergeTasks false MissingEntry: level[5] @mergetasks (Config.MergeTasks)bool
Config.Require nil MissingEntry: level[5] @require (Config.Require)string
Config.Sequencially false MissingEntry: level[5] @sequencially (Config.Sequencially)bool
Config.Use nil MissingEntry: level[5] @use (Config.Use)string
Config.Variables nil MissingEntry: level[5] @variables (Config.Variables)string
Task.Listener nil MissingEntry: level[5] @listener (Task.Listener)string
Task.Needs nil MissingEntry: level[5] @needs (Task.Needs)string
Task.Next nil MissingEntry: level[5] @next (Task.Next)string
Task.Options  MissingEntry: level[5] @options (Task.Options)undefined
Task.Options.Bgcolorcode  MissingEntry: level[5] @bgcolorcode (Task.Options.Bgcolorcode)string
Task.Options.Colorcode  MissingEntry: level[5] @colorcode (Task.Options.Colorcode)string
Task.Options.Displaycmd false MissingEntry: level[5] @displaycmd (Task.Options.Displaycmd)bool
Task.Options.Format  MissingEntry: level[5] @format (Task.Options.Format)string
Task.Options.Hideout false MissingEntry: level[5] @hideout (Task.Options.Hideout)bool
Task.Options.IgnoreCmdError false MissingEntry: level[5] @ignoreCmdError (Task.Options.IgnoreCmdError)bool
Task.Options.Invisible false MissingEntry: level[5] @invisible (Task.Options.Invisible)bool
Task.Options.Maincmd  MissingEntry: level[5] @maincmd (Task.Options.Maincmd)string
Task.Options.Mainparams nil MissingEntry: level[5] @mainparams (Task.Options.Mainparams)string
Task.Options.NoAutoRunNeeds false MissingEntry: level[5] @noAutoRunNeeds (Task.Options.NoAutoRunNeeds)bool
Task.Options.Panelsize 0 MissingEntry: level[5] @panelsize (Task.Options.Panelsize)int
Task.Options.Stickcursor false MissingEntry: level[5] @stickcursor (Task.Options.Stickcursor)bool
Task.Options.TickTimeNeeds 0 MissingEntry: level[5] @tickTimeNeeds (Task.Options.TickTimeNeeds)int
Task.Options.TimeoutNeeds 0 MissingEntry: level[5] @timeoutNeeds (Task.Options.TimeoutNeeds)int
Task.Options.WorkingDir  MissingEntry: level[5] @workingdir (Task.Options.WorkingDir)string
Task.Requires.Environment nil MissingEntry: level[5] @environment (Task.Requires.Environment)string
Task.Requires.Exists nil MissingEntry: level[5] @exists (Task.Requires.Exists)string
Task.Requires.NotExists nil MissingEntry: level[5] @notExists (Task.Requires.NotExists)string
Task.Requires.Variables nil MissingEntry: level[5] @variables (Task.Requires.Variables)string
Task.RunTargets nil MissingEntry: level[5] @runTargets (Task.RunTargets)string
Task.Stopreasons  MissingEntry: level[5] @stopreasons (Task.Stopreasons)undefined
Task.Stopreasons.Now false MissingEntry: level[5] @now (Task.Stopreasons.Now)bool
Task.Stopreasons.Onerror false MissingEntry: level[5] @onerror (Task.Stopreasons.Onerror)bool
Task.Stopreasons.OnoutContains nil MissingEntry: level[5] @onoutContains (Task.Stopreasons.OnoutContains)string
Task.Stopreasons.OnoutcountLess 0 MissingEntry: level[5] @onoutcountLess (Task.Stopreasons.OnoutcountLess)int
Task.Stopreasons.OnoutcountMore 0 MissingEntry: level[5] @onoutcountMore (Task.Stopreasons.OnoutcountMore)int
Task.Variables nil MissingEntry: level[5] @variables (Task.Variables)string
Version  MissingEntry: level[5] @version (Version)string
Workspace  MissingEntry: level[5] @workspace (Workspace)undefined
Workspace.Project  MissingEntry: level[5] @project (Workspace.Project)string
Workspace.Role  MissingEntry: level[5] @role (Workspace.Role)string
Workspace.Version  MissingEntry: level[5] @version (Workspace.Version)string`
	assertSplitTestInMessage(t, output, expected)

	// now we add thos path to the project
	output.ClearAndLog()
	if err := runCobraCmd(app, "dir add"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	fullPath := getAbsolutePath("projects01/dangerduck/varia")
	assertInMessage(t, output, "add "+fullPath)

	output.ClearAndLog()
	// simle checking if we see any project, and get the mark for the current one
	if err := runCobraCmd(app, "workspace show"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	expected = `
samira
[ ducktale ]
`

	assertSplitTestInMessage(t, output, expected)

}

func TestImportRequired(t *testing.T) {
	ChangeToRuntimeDir(t)
	app, output, appErr := SetupTestApp("projects01", time.Now().Format(time.RFC3339)+"ctx_projects.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	output.SetKeepNewLines(true)
	defer cleanAllFiles()
	defer output.ClearAndLog()
	// clean the output buffer
	output.Clear()
	logFileName := "testRequired_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))

	// change into the test directory
	// especially here to the first projectwe have there.
	if err := os.Chdir(getAbsolutePath("projects01/testrequire")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	// print the targets
	if err := runCobraCmd(app, "run"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "first")
	assertInMessage(t, output, "docker-stop-all")
	assertInMessage(t, output, "docker-show-ip")

	output.ClearAndLog()
	// lint yaml to get the merged context
	if err := runCobraCmd(app, "lint yaml"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "- id: first")
	assertInMessage(t, output, "- id: docker-stop-all")
	assertInMessage(t, output, "- id: docker-show-ip")
	assertInMessage(t, output, "- swaros/ctx-docker")
}

// testing if the prevalues are working as expected.
// this is the used context file:
/*
config:
  variables:
     username: "master"
     password: "check12345"

task:
  - id: values
    script:
      - echo "props [${username}] [${password}]"

  - id: rewrite
    variables:
      username: "jon-doe"
      password: "mysecret"
    script:
      - echo "reused [${username}] [${password}]"

*/
func TestSetPrevalues(t *testing.T) {
	ChangeToRuntimeDir(t)
	app, output, appErr := SetupTestApp("projects01", time.Now().Format(time.RFC3339)+"ctx_projects.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	output.SetKeepNewLines(false)
	defer cleanAllFiles()
	defer output.ClearAndLog()
	// clean the output buffer
	output.Clear()
	logFileName := "testPrevalues_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))

	// change into the test directory
	// especially here to the first projectwe have there.
	if err := os.Chdir(getAbsolutePath("projects01/prevalues")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	// just lint
	if err := runCobraCmd(app, "lint"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	// first without any prevalues
	if err := runCobraCmd(app, "run values"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "props [master] [check12345]")

	output.ClearAndLog()
	// next just replace the username
	if err := runCobraCmd(app, "run values -v username=root"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "props [root] [check12345]")
	output.ClearAndLog()

	// next just replace the password and again the username
	if err := runCobraCmd(app, "run values -v password=889977 -v username=mimi"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "props [mimi] [889977]")
	output.ClearAndLog()

	// here we check if variables are still be set, if the defined in the task definition
	if err := runCobraCmd(app, "run rewrite -v password=889977 -v username=mimi"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "reused [jon-doe] [mysecret]")
	output.ClearAndLog()

	// next we check the chain usage of prevalues and rewrite after launch them one after the other
	if err := runCobraCmd(app, "run values rewrite -v password=kmmgt -v username=rumble"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	assertInMessage(t, output, "props [rumble] [kmmgt]")
	assertInMessage(t, output, "reused [jon-doe] [mysecret]")
	output.ClearAndLog()

	// next in different order
	// TODO: commented, because this is the behavior from V1, but now we have a new behavior. and i am not sure what is more useful

	/*
		if err := runCobraCmd(app, "run rewrite values -v password=ppoker -v username=lucker"); err != nil {
			t.Errorf("Expected no error, got '%v'", err)
		}
		assertInMessage(t, output, "props [jon-doe] [mysecret]")
		assertInMessage(t, output, "reused [jon-doe] [mysecret]")
		output.ClearAndLog()
	*/

}

func TestCreate(t *testing.T) {
	ChangeToRuntimeDir(t)
	app, output, appErr := SetupTestApp("workspace0", time.Now().Format(time.RFC3339)+"ctx_projects.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	// removing the test file, if it exists, at start and end
	removeFilePath := getAbsolutePath("workspace0/testcreate/.contxt.yml")
	removeFilePath2 := getAbsolutePath("workspace0/testcreate/.inc.contxt.yml")
	removeFile(removeFilePath)
	removeFile(removeFilePath2)
	defer removeFile(removeFilePath)
	defer removeFile(removeFilePath2)

	output.SetKeepNewLines(true)
	defer cleanAllFiles()
	defer output.ClearAndLog()
	// clean the output buffer
	output.Clear()
	logFileName := "testCreate_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))

	// change into the test directory
	// especially here to the first projectwe have there.
	if err := os.Chdir(getAbsolutePath("workspace0/testcreate")); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	// first create new context file
	if err := runCobraCmd(app, "create"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

	// print the targets
	if err := runCobraCmd(app, "run"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	// we should have the default task "my_task" in the output
	assertInMessage(t, output, "my_task")

	output.ClearAndLog()

	// testing create import without any path, what should fail
	if err := runCobraCmd(app, "create import"); err == nil {
		t.Error("Expected an error, got none")
	} else {
		expectedError := "no path given"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error '%v', got '%v'", expectedError, err)
		}
	}

	// testing create import with a path, what should work
	if err := runCobraCmd(app, "create import imports"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}

}

func TestAnkoRunner(t *testing.T) {
	ChangeToRuntimeDir(t)
	app, output, appErr := SetupTestApp("workspace0", time.Now().Format(time.RFC3339)+"ctx_projects.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}

	output.SetKeepNewLines(true)
	defer cleanAllFiles()
	defer output.ClearAndLog()
	// clean the output buffer
	output.Clear()
	logFileName := "testAnko_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))

	if err := runCobraCmd(app, "anko println('hello')"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
}

func TestAnkoFileExecute(t *testing.T) {
	ChangeToRuntimeDir(t)
	app, output, appErr := SetupTestApp("workspace0", time.Now().Format(time.RFC3339)+"ctx_projects.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}

	output.SetKeepNewLines(true)
	defer cleanAllFiles()
	defer output.ClearAndLog()
	// clean the output buffer
	output.Clear()
	logFileName := "testAnkoFile_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))

	if err := runCobraCmd(app, "anko -f test1.anko"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
}

func TestAnkoFileExecuteWrongFile(t *testing.T) {
	ChangeToRuntimeDir(t)
	app, output, appErr := SetupTestApp("workspace0", time.Now().Format(time.RFC3339)+"ctx_projects.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}

	output.SetKeepNewLines(true)
	defer cleanAllFiles()
	defer output.ClearAndLog()
	// clean the output buffer
	output.Clear()
	logFileName := "testAnkoFileNotExists_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))

	if err := runCobraCmd(app, "anko -f wrongname.anko"); err == nil {
		t.Error("Expected an error, got none")
	} else {
		// expected error message is: file wrongname.anko not found
		expectedError := "file wrongname.anko not found"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected error '%v', got '%v'", expectedError, err)
		}
	}
}
