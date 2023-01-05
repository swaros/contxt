package tasks_test

import (
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/tasks"
)

// this test would fail because the requirment handler is not set
func TestFailureBecauseNoRequirementCheck(t *testing.T) {

	var testTask configure.Task = configure.Task{
		ID:     "test",
		Script: []string{"echo test"},
	}

	var runCfg configure.RunConfig = configure.RunConfig{}
	runCfg.Task = []configure.Task{testTask}

	tasks := tasks.NewTaskListExec(runCfg)

	code := tasks.RunTarget("test", true)
	if code != 107 {
		t.Errorf("Expected code 107, got %d", code)
	}

}
