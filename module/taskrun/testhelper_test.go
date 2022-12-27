package taskrun_test

import (
	"os"
	"testing"

	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/taskrun"
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

// folderRunner takes a path relative to the current module path (most starts with "./../docs/")
// and executes the the function.
// if the folder is not valid, a error will be returned instead
func folderRunner(folder string, t *testing.T, testFunc func(t *testing.T)) error {
	taskrun.ClearAll()
	taskrun.ResetAllTaskInfos()
	taskrun.InitDefaultVars()
	old, derr := dirhandle.Current()
	if derr != nil {
		t.Error(derr)
		return derr
	}
	if err := os.Chdir(folder); err != nil {
		t.Error(err)
		return err
	}
	testFunc(t)
	if err := os.Chdir(old); err != nil {
		t.Error(err)
		return err
	}
	return nil

}

func eqCheckVarMapValue(keyname, jsonPath, expected string, success, fail, notfound func(result string)) {
	if result, ok := taskrun.GetJSONPathResult(keyname, jsonPath); ok {
		if result.Str == expected {
			success(result.Str)
		} else {
			fail(result.Str)
		}
	} else {
		notfound(result.Str)
	}
}

// assertVarStrEquals is testing a contxt variable content against the expected value
func assertVarStrEquals(t *testing.T, keyname, expected string) bool {
	check := clearStrings(taskrun.GetPH(keyname))

	if check != clearStrings(expected) {
		t.Error("expected " + expected + " as variable. but got <" + check + ">")
		return false
	}
	return true
}

// assertVarStrEquals is testing a contxt variable content against the expected value
func assertVarStrNotEquals(t *testing.T, keyname, unExpected string) bool {
	check := clearStrings(taskrun.GetPH(keyname))

	if check == clearStrings(unExpected) {
		t.Error("unexpected [" + unExpected + "] is present ")
		return false
	}
	return true
}

/*
func assertStringEquals(t *testing.T, actual, expected string) bool {
	check := clearStrings(actual)

	if check != clearStrings(expected) {
		t.Error("expected " + expected + " as variable. but got <" + check + ">")
		return false
	}
	return true
}
*/

// assertCaseLogLastEquals tests for a case (docs/tests/case<nr>) if the last output
// is equasl to the expected.
// depending on the executed targetName
func assertCaseLogLastEquals(t *testing.T, caseNr, targetName, expected string) {
	caseRunner(caseNr, t, func(t *testing.T) {
		taskrun.RunTargets(targetName, true)
		log := taskrun.GetPH("RUN." + targetName + ".LOG.LAST")
		if log != expected {
			t.Error("the last executed script should be "+expected+". not ", log)
		}
	})
}

func assertTestFolderVarFn(t *testing.T, folder, targetRun string, fn func()) {
	if runError := folderRunner("./../../docs/test/"+folder, t, func(t *testing.T) {
		taskrun.RunTargets(targetRun, false)
		fn()
	}); runError != nil {
		t.Error(runError)
	}
}
