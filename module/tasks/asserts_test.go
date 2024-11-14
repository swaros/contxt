package tasks_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/tasks"
	"gopkg.in/yaml.v2"
)

func createRuntimeByYamlString(yamlString string, messages *[]string) (*tasks.TaskListExec, error) {
	return createRuntimeByYamlStringWithAllMsg(yamlString, messages, nil, nil, nil)
}

func createRuntimeByYamlStringWithErrors(yamlString string, messages *[]string, errors *[]error) (*tasks.TaskListExec, error) {
	return createRuntimeByYamlStringWithAllMsg(yamlString, messages, errors, nil, nil)
}

func createRuntimeByYamlStringWithAllMsg(yamlString string, messages *[]string, errors *[]error, typeMsg *[]string, targetUpdates *[]string) (*tasks.TaskListExec, error) {
	var runCfg configure.RunConfig = configure.RunConfig{}

	if err := yaml.Unmarshal([]byte(yamlString), &runCfg); err != nil {
		return nil, err
	} else {
		outHandler := func(msg ...interface{}) {
			for _, m := range msg {
				switch mt := m.(type) {
				case tasks.MsgExecOutput: // this will be the output of the command
					*messages = append(*messages, string(mt.Output))
				case tasks.MsgError: // this will be the error of the command
					if errors != nil {
						*errors = append(*errors, mt.Err)
					}

				case tasks.MsgTarget:
					if targetUpdates != nil {
						*targetUpdates = append(*targetUpdates, mt.Target+":"+mt.Context+"["+mt.Info+"]")
					}

				case tasks.MsgType:
					if typeMsg != nil {
						*typeMsg = append(*typeMsg, string(mt))
					}
				}
			}

		}

		dmc := tasks.NewCombinedDataHandler()
		req := tasks.NewDefaultRequires(dmc, mimiclog.NewNullLogger())

		tsk := tasks.NewTaskListExec(
			runCfg,
			dmc,
			outHandler,
			tasks.ShellCmd,
			req,
		)
		// disbale hard exit for tests
		tsk.SetHardExistToAllTasks(false)
		return tsk, err
	}
}

func assertSliceContains(t *testing.T, slice []string, substr string) bool {
	t.Helper()
	for _, s := range slice {
		if strings.Contains(s, substr) {
			return true
		}
	}
	t.Errorf("expected slice to contain %q, but it did not", substr)
	t.Log("\n" + strings.Join(slice, "\n"))
	return false
}

func assertSliceNotContains(t *testing.T, slice []string, substr string) bool {
	t.Helper()
	for _, s := range slice {
		if strings.Contains(s, substr) {
			t.Errorf("expected slice not to contain %q, but it did", substr)
			t.Log("\n" + strings.Join(slice, "\n"))
			return false
		}
	}
	return true
}

func helpLogSlice(t *testing.T, slice []string) {
	t.Helper()
	t.Log("\n" + strings.Join(slice, "\n"))
}

func assertSliceContainsAtLeast(t *testing.T, slice []string, substr string, atLeast int) bool {
	t.Helper()
	count := 0
	for _, s := range slice {
		if strings.Contains(s, substr) {
			count++
		}
	}
	if count < atLeast {
		t.Errorf("expected slice to contain %q at least %d times, but it did only %d times", substr, atLeast, count)
		t.Log("\n" + strings.Join(slice, "\n"))
		return false
	}
	return true
}

func assertContainsCount(t *testing.T, slice []string, substr string, count int) {
	t.Helper()
	containsCount := 0
	for _, s := range slice {
		if s == substr {
			containsCount++
		}
	}
	if containsCount != count {
		t.Errorf("expected slice to contain %d instances of %q, but it contained %d", count, substr, containsCount)
		t.Log("\n" + strings.Join(slice, "\n"))
	}
}

func assertStrEqual(t *testing.T, expected string, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected [%q], but got [%q]", expected, actual)
	}
}

func assertIntEqual(t *testing.T, expected int, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected [%d], but got [%d]", expected, actual)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error, but got %v", err)
	}
}

func assertPositionInSliceBefore(t *testing.T, slice []string, substr string, before string) {
	t.Helper()
	substrIndex := -1
	beforeIndex := -1
	for i, s := range slice {
		if s == substr {
			substrIndex = i
		}
		if s == before {
			beforeIndex = i
		}
	}
	if substrIndex == -1 {
		t.Errorf("expected slice to contain %q, but it did not", substr)
	}
	if beforeIndex == -1 {
		t.Errorf("expected slice to contain %q, but it did not", before)
	}
	if substrIndex >= beforeIndex {
		t.Errorf("expected %q to be before %q, but it was not", substr, before)
	}
}

func helpsRunAsync(runCount int, runinngs []string, testDoSome func(name string, cnt int) bool) bool {
	var wg sync.WaitGroup
	allFine := true
	doInc := func(name string, n int) {
		for i := 0; i < n; i++ {
			allFine = allFine && testDoSome(name, i)
		}
		wg.Done()
	}

	wg.Add(len(runinngs))
	for _, name := range runinngs {
		go doInc(name, runCount)
	}
	wg.Wait()
	return allFine
}

func assertDirectoryMatch(t *testing.T, originFolder, targetFolder string) error {
	t.Helper()
	// walkinng the origin folder
	originFiles := make(map[string]bool)
	completeOrginFilesPath := make(map[string]string)

	err := filepath.Walk(originFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		trimmedPath := strings.TrimPrefix(path, originFolder)
		if trimmedPath == "" {
			// we are in the origin folder
			return nil
		}
		originFiles[trimmedPath] = info.IsDir()
		if !info.IsDir() {
			completeOrginFilesPath[trimmedPath] = path
		}
		return nil
	})
	if err != nil {
		return err
	}

	// walking the target folder
	err = filepath.Walk(targetFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		expectedPath := strings.TrimPrefix(path, targetFolder)
		if expectedPath == "" {
			// we are in the target folder
			return nil
		}
		originTrimmed := strings.TrimPrefix(path, targetFolder)
		if originTrimmed == "" {
			// we are in the origin folder
			return nil
		}
		_, ok := originFiles[expectedPath]
		originIsDir := originFiles[originTrimmed]
		if !ok {
			t.Errorf("file %q not found in origin", path)
			return nil
		}
		if originIsDir != info.IsDir() {
			t.Errorf("file %q is a directory in origin but not in target", path)
		}
		delete(originFiles, path)
		// file content comparison by assertFileMatch
		if !info.IsDir() {
			err := assertFileMatch(t, completeOrginFilesPath[expectedPath], path)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func removeFile(t *testing.T, file string) {
	t.Helper()
	if err := os.Remove(file); err != nil {
		t.Errorf("could not remove file %q: %v", file, err)
	}
}

func assertFileMatch(t *testing.T, originFile, targetFile string) error {
	return assertFileMatchAndRemoveOrig(t, originFile, targetFile, false)
}

func assertFileMatchAndRemoveOrig(t *testing.T, originFile, targetFile string, remove bool) error {
	t.Helper()
	originFileStat, err := os.Stat(originFile)
	if err != nil {
		return err
	}
	if remove {
		defer removeFile(t, originFile)
	}
	targetFileStat, err := os.Stat(targetFile)
	if err != nil {
		return err
	}
	if originFileStat.IsDir() != targetFileStat.IsDir() {
		t.Errorf("file %q is a directory in origin but not in target", targetFile)
	}
	// file content comparison
	originFileContent, err := os.ReadFile(originFile)
	if err != nil {
		return err
	}
	targetFileContent, err := os.ReadFile(targetFile)
	if err != nil {
		return err
	}
	if string(originFileContent) != string(targetFileContent) {
		t.Errorf("file %q content does not match", targetFile)
	}
	return nil
}

func assertFileNotExists(t *testing.T, file string) {
	t.Helper()
	if _, err := os.Stat(file); err == nil {
		t.Errorf("file %q exists, but it should not", file)
	}
}

func assertFileExists(t *testing.T, file string) {
	t.Helper()
	if _, err := os.Stat(file); err != nil {
		t.Errorf("file %q does not exist, but it should", file)
	}
}

func AnkoTestRunHelper(
	t *testing.T,
	cmd string,
	expectedExitCode int,
	expectedMessageCount int, expectedMessage []string) (*tasks.CombinedDh, *tasks.DefaultRequires) {
	t.Helper()
	return AnkoTestRunHelperWithErrors(t, cmd, false, expectedExitCode, expectedMessageCount, expectedMessage)
}

func AnkoTestRunHelperWithErrors(
	t *testing.T,
	cmd string,
	errorsExpected bool,
	expectedExitCode int,
	expectedMessageCount int, expectedMessage []string) (*tasks.CombinedDh, *tasks.DefaultRequires) {
	t.Helper()

	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					cmd,
				},
			},
		},
	}

	localMsg := []string{}
	errorMsg := []string{}
	outHandler := func(msg ...interface{}) {
		for _, m := range msg {
			switch s := m.(type) {
			case string:
				localMsg = append(localMsg, s)
			case tasks.MsgExecOutput:
				localMsg = append(localMsg, string(s.Output))
			case tasks.MsgError:
				errorMsg = append(errorMsg, s.Err.Error())
				if !errorsExpected {
					t.Error(s.Err)
				}
			case tasks.MsgErrDebug:
				t.Log(s)
				if !errorsExpected {
					errorMsg = append(errorMsg, s.Err.Error())
					t.Error(s.Err)
				}
			}
		}
	}
	dmc := tasks.NewCombinedDataHandler()
	req := tasks.NewDefaultRequires(dmc, mimiclog.NewNullLogger())
	tsk := tasks.NewTaskListExec(runCfg, dmc, outHandler, tasks.ShellCmd, req)
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected exit code ", expectedExitCode, " but got", code)
	}
	if expectedMessageCount > 0 {
		if len(localMsg) != expectedMessageCount {
			t.Error("expected ", expectedMessageCount, " message but got", len(localMsg))
			for i, m := range localMsg {
				t.Log(i, m)
			}

		} else {
			for i, expected := range expectedMessage {
				if i >= len(localMsg) {
					continue
				}
				cleanExpect := strings.TrimSpace(expected)
				cleanLocal := strings.TrimSpace(localMsg[i])
				cleanExpect = strings.ReplaceAll(cleanExpect, "\n", "")
				cleanLocal = strings.ReplaceAll(cleanLocal, "\n", "")
				cleanExpect = strings.ReplaceAll(cleanExpect, "\r", "")
				cleanLocal = strings.ReplaceAll(cleanLocal, "\r", "")
				cleanExpect = strings.ReplaceAll(cleanExpect, "\t", "")
				cleanLocal = strings.ReplaceAll(cleanLocal, "\t", "")
				if cleanLocal == "" && strings.ReplaceAll(cleanExpect, " ", "") == "" {
					continue
				}
				if cleanExpect == "" && strings.ReplaceAll(cleanLocal, " ", "") == "" {
					continue
				}
				if strings.Contains(cleanExpect, "!IGNORE") {
					continue
				}
				if cleanLocal != cleanExpect {
					t.Error("expected message[", cleanExpect, "] but got[", cleanLocal, "] line:", i)
				}

			}
		}
	} else {
		if len(localMsg) > 0 {
			t.Error("expected no message but got", len(localMsg))
			t.Log(localMsg)
		}
	}
	return dmc, req
}

// RunTargetHelper is a helper function to run a task with a given command and
// check the output and exit code.
// t: the testing.T object
// cmd: the command to run. like "run test-target"
// runCfg: the run configuration. there anything can be set related to the tasks
// expectedExitCode: the expected exit code. see systools/errorcodes.go
// expectedMessageCount: the expected message count. < 0 means no check
// expectedMessage: the expected messages. it depends on the expectedMessageCount. is this is < 1, this will be ignored
// resetTasks: if true, the tasks will be resetted after the run
// returns: the combined data handler and the default requires and the test logger
func RunTargetHelperWithErrors(
	t *testing.T,
	cmd string,
	runCfg configure.RunConfig,
	errorsExpected bool,
	expectedExitCode int,
	expectedMessageCount int,
	expectedMessage []string,
	resetTasks bool,
) (*tasks.CombinedDh, *tasks.DefaultRequires, *testLogger) {
	t.Helper()

	localMsg := []string{}
	errorMsg := []string{}
	outHandler := func(msg ...interface{}) {
		for _, m := range msg {
			switch s := m.(type) {
			case string:
				localMsg = append(localMsg, s)
			case tasks.MsgExecOutput:
				localMsg = append(localMsg, string(s.Output))
			case tasks.MsgError:
				errorMsg = append(errorMsg, s.Err.Error())
				if !errorsExpected {
					t.Error(s.Err)
				}
			case tasks.MsgErrDebug:
				t.Log(s)
				if !errorsExpected {
					errorMsg = append(errorMsg, s.Err.Error())
					t.Error(s.Err)
				}
			}
		}
	}
	testLogger := NewTestLogger(t)
	dmc := tasks.NewCombinedDataHandler()
	req := tasks.NewDefaultRequires(dmc, testLogger)
	tsk := tasks.NewTaskListExec(runCfg, dmc, outHandler, tasks.ShellCmd, req)
	tsk.SetLogger(testLogger)
	tsk.SetHardExistToAllTasks(false)
	if resetTasks {
		tsk.GetWatch().ResetAllTaskInfos()
	}
	code := tsk.RunTarget(cmd, !runCfg.Config.Sequencially)
	if code != expectedExitCode {
		t.Error("expected exit code ", expectedExitCode, " but got", code)
	}
	if expectedMessageCount > 0 {
		if len(localMsg) != expectedMessageCount {
			t.Error("expected ", expectedMessageCount, " message but got", len(localMsg))
			for i, m := range localMsg {
				t.Log(i, m)
			}

		} else {
			for i, expected := range expectedMessage {
				if i >= len(localMsg) {
					continue
				}
				cleanExpect := strings.TrimSpace(expected)
				cleanLocal := strings.TrimSpace(localMsg[i])
				cleanExpect = strings.ReplaceAll(cleanExpect, "\n", "")
				cleanLocal = strings.ReplaceAll(cleanLocal, "\n", "")
				cleanExpect = strings.ReplaceAll(cleanExpect, "\r", "")
				cleanLocal = strings.ReplaceAll(cleanLocal, "\r", "")
				cleanExpect = strings.ReplaceAll(cleanExpect, "\t", "")
				cleanLocal = strings.ReplaceAll(cleanLocal, "\t", "")
				if cleanLocal == "" && strings.ReplaceAll(cleanExpect, " ", "") == "" {
					continue
				}
				if cleanExpect == "" && strings.ReplaceAll(cleanLocal, " ", "") == "" {
					continue
				}
				if strings.Contains(cleanExpect, "!IGNORE") {
					continue
				}
				if cleanLocal != cleanExpect {
					t.Error("expected message[", cleanExpect, "] but got[", cleanLocal, "] line:", i)
				}

			}
		}
	} else {
		if len(localMsg) > 0 && expectedMessageCount > 0 {
			t.Error("expected no message but got", len(localMsg))
			t.Log(localMsg)
		}
	}
	return dmc, req, testLogger
}

// testing the assert itself

func TestAssertSliceContains(t *testing.T) {
	t.Parallel()
	slice := []string{"a", "b", "c"}
	assertSliceContains(t, slice, "b")
}

func TestAssertSliceNotContains(t *testing.T) {
	t.Parallel()
	slice := []string{"a", "b", "c"}
	assertSliceNotContains(t, slice, "d")
}

func TestAssertSliceContainsAtLeast(t *testing.T) {
	t.Parallel()
	slice := []string{"a", "b", "c", "b", "b"}
	assertSliceContainsAtLeast(t, slice, "b", 3)
}
