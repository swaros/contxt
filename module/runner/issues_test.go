package runner_test

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/swaros/contxt/module/tasks"
)

/* issue 185
https://github.com/swaros/contxt/issues/185

there is an issue by loading the default values from the config file.
where the secrets stored on a local configuration file.
it seems that the local values are not loaded correctly.
*/

// TestIssue185_1 is testing the user local values from the config file
func TestIssue185_1(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not working on windows")
	}
	ChangeToRuntimeDir(t)
	tasks.NewGlobalWatchman().ResetAllTaskInfos()
	app, output, appErr := SetupTestApp("issues", "ctx_test_config.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	cleanAllFiles()
	defer cleanAllFiles()

	// set the log file with an timestamp
	logFileName := "issue_185_1_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))
	output.ClearAndLog()

	// change into the test directory
	if err := os.Chdir(getAbsolutePath("issues/issue_185")); err != nil {
		t.Errorf("error by changing the directory. check test: '%v'", err)
	}

	if err := runCobraCmd(app, "run build-base-image"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	expected := "--build-arg CMP_KEYID=0815 --build-arg CMP_ACID=thisisAnExample --build-arg LICENCE_KEY=andthisisalsonotanrealkey"
	assertInMessage(t, output, expected)
	output.ClearAndLog()

}

// TestIssue185_2 is testing the default values from the config file if no user local values are set (not found deping user name)
func TestIssue185_2(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not working on windows")
	}
	ChangeToRuntimeDir(t)
	tasks.NewGlobalWatchman().ResetAllTaskInfos()
	app, output, appErr := SetupTestApp("issues", "ctx_test_config.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	cleanAllFiles()
	defer cleanAllFiles()

	// set the log file with an timestamp
	logFileName := "issue_185_2_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))
	output.ClearAndLog()

	// change into the test directory
	if err := os.Chdir(getAbsolutePath("issues/issue_185")); err != nil {
		t.Errorf("error by changing the directory. check test: '%v'", err)
	}
	// testing with an variable where no default file is set, so the default value should be used
	if err := runCobraCmd(app, "run build-base-image -v CFG_USER=uni"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	expected := "build-base-image--build-arg CMP_KEYID=0 --build-arg CMP_ACID=unset_1 --build-arg LICENCE_KEY=0"
	assertInMessage(t, output, expected)
	output.ClearAndLog()

}

// TestIssue185_3 is testing the local values from another given user
func TestIssue185_3(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not working on windows")
	}
	ChangeToRuntimeDir(t)
	tasks.NewGlobalWatchman().ResetAllTaskInfos()
	app, output, appErr := SetupTestApp("issues", "ctx_test_config.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	cleanAllFiles()
	defer cleanAllFiles()
	output.Clear()
	// set the log file with an timestamp
	logFileName := "issue_185_3_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))
	output.ClearAndLog()

	// change into the test directory
	if err := os.Chdir(getAbsolutePath("issues/issue_185")); err != nil {
		t.Errorf("error by changing the directory. check test: '%v'", err)
	}
	// testing with an variable where no default file is set, so the default value should be used
	if err := runCobraCmd(app, "run build-base-image -v CFG_USER=user2 --loglevel=DEBUG"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	expected := "build-base-image--build-arg CMP_KEYID=88547 --build-arg CMP_ACID=somethingElse --build-arg LICENCE_KEY=forUserNo2"
	assertInMessage(t, output, expected)
	output.ClearAndLog()

}

// TestIssue185_4 is testing the ${USER} variable because any other case is working right now
// without any changes in the code.
func TestIssue185_4(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not working on windows")
	}
	ChangeToRuntimeDir(t)
	tasks.NewGlobalWatchman().ResetAllTaskInfos()
	app, output, appErr := SetupTestApp("issues", "ctx_test_config.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	cleanAllFiles()
	defer cleanAllFiles()

	// set the log file with an timestamp
	logFileName := "issue_185_4_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))
	output.ClearAndLog()

	// change into the test directory
	if err := os.Chdir(getAbsolutePath("issues/issue_185")); err != nil {
		t.Errorf("error by changing the directory. check test: '%v'", err)
	}
	// testing with an variable where no default file is set, so the default value should be used
	if err := runCobraCmd(app, "run test-actual-user"); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
	username := os.Getenv("USER")
	expected := "you are [" + username + "]"
	assertInMessage(t, output, expected)
	expected = "not loading local config file from user " + username + ".local.json"
	assertInMessage(t, output, expected)
	output.ClearAndLog()

}
