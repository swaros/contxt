package tasks_test

import (
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/tasks"
	"gopkg.in/yaml.v2"
)

func createRuntimeByYamlString(yamlString string, messages *[]string) (*tasks.TaskListExec, error) {
	var runCfg configure.RunConfig = configure.RunConfig{}

	if err := yaml.Unmarshal([]byte(yamlString), &runCfg); err != nil {
		return nil, err
	} else {
		outHandler := func(msg ...interface{}) {
			for _, m := range msg {
				switch mt := m.(type) {
				case tasks.MsgExecOutput: // this will be the output of the command
					//messages = append(messages, string(mt))
					*messages = append(*messages, string(mt))
				}
			}

		}

		return tasks.NewTaskListExec(
			runCfg,
			tasks.NewCombinedDataHandler(),
			outHandler,
			tasks.ShellCmd,
			func(require configure.Require) (bool, string) { return true, "" },
		), err
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
