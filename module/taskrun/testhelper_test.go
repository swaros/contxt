package taskrun_test

import (
	"os"
	"testing"

	"github.com/swaros/contxt/dirhandle"
	"github.com/swaros/contxt/taskrun"
)

// caseRunner helps to switch a testrunn in testcase directory to this
// this folder. and go back after the test is done
// it also resets all variables
// the id is just the number of the test/case folder (postfix)
func caseRunner(id string, t *testing.T, testFunc func(t *testing.T)) {
	taskrun.ClearAll()
	old, derr := dirhandle.Current()
	if derr != nil {
		t.Error(derr)
	}
	dir := "./../../docs/test/case" + id
	taskrun.GetLogger().Debug("--- [CR] TESTING FILE " + dir)
	if err := os.Chdir(dir); err == nil {
		testFunc(t)
		os.Chdir(old)
	} else {
		t.Error(err)
	}

}

func folderRunner(folder string, t *testing.T, testFunc func(t *testing.T)) {
	taskrun.ClearAll()
	taskrun.ResetAllTaskInfos()
	taskrun.InitDefaultVars()
	old, derr := dirhandle.Current()
	if derr != nil {
		t.Error(derr)
	}
	os.Chdir(folder)
	testFunc(t)
	os.Chdir(old)

}
