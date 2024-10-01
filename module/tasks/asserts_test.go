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

func assertFileMatch(t *testing.T, originFile, targetFile string) error {
	t.Helper()
	originFileStat, err := os.Stat(originFile)
	if err != nil {
		return err
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
