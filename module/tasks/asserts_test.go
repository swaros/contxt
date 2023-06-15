package tasks_test

import (
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
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
					*messages = append(*messages, string(mt))
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
		req := tasks.NewDefaultRequires(dmc, logrus.New())

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
	}
}

func assertStrEqual(t *testing.T, expected string, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected [%q], but got [%q]", expected, actual)
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
