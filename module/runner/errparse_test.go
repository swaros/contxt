package runner_test

import (
	"os"
	"testing"
	"time"
)

// Helper function to setup Test App that only runs in testdata/yamerror directory
// it is all about lower the code duplication
func InitErrorCheckTest(t *testing.T, relatedPath string, cobraCmd string) (*TestOutHandler, error) {
	t.Helper()
	ChangeToRuntimeDir(t)
	app, output, appErr := SetupTestApp("yamlerror", time.Now().Format(time.RFC3339)+"ctx_projects.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}

	output.SetKeepNewLines(true)
	defer cleanAllFiles()
	// clean the output buffer
	output.Clear()
	logFileName := relatedPath + "_error_" + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))

	if err := os.Chdir(getAbsolutePath("yamlerror/" + relatedPath)); err != nil {
		t.Errorf("Expected no error by getting to the test directory, got '%v'", err)
		return nil, err
	}

	return output, runCobraCmd(app, cobraCmd)
}

func TestJsonReadAsYaml(t *testing.T) {

	if output, err := InitErrorCheckTest(t, "jsonfile", "run test"); err == nil {
		t.Errorf("Expected error, got nil")
	} else {
		defer output.ClearAndLog()
		assertInMessage(t, output, "error explanation: Verify the yaml file. if you have a yaml file that is not correctly formatted, you will get this error.")

		if err.Error() != "yaml: line 2: did not find expected ',' or '}'" {
			t.Errorf("Got unexpected Error [%v]", err.Error())
		}
	}

}

func TestBrokenTpl(t *testing.T) {

	if output, err := InitErrorCheckTest(t, "brokentpl", "run maniac"); err == nil {
		t.Errorf("Expected error, got nil")
	} else {
		defer output.ClearAndLog()
		assertInMessage(t, output, "error explanation: Verify the code of the template file. this type of error can happen if we try to parse temlates")

		if err.Error() != "template: contxt-functions:4: function \"include\" not defined" {
			t.Errorf("Got unexpected Error [%v]", err.Error())
		}
	}

}
