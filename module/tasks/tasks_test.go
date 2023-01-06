package tasks_test

import (
	"strings"
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/tasks"
)

// this test would fail because the requirment handler is not set

func TestFailureBecauseNoRequirementCheck(t *testing.T) {
	messages := []string{}
	// we hook into the output handler to capture the output
	ctxout.PreHook = func(msg ...interface{}) bool {
		messages = append(messages, ctxout.ToString(msg...))
		return true
	}
	// we create a task that will fail because the requirement check is not set
	var testTask configure.Task = configure.Task{
		ID:     "test",                // we need an ID
		Script: []string{"echo test"}, // we need a script
		Options: configure.Options{
			Displaycmd: true, // we want to see the command. we capture the output so we can check it
		},
	}
	// we create a task list with the task
	var runCfg configure.RunConfig = configure.RunConfig{}
	runCfg.Task = []configure.Task{testTask}

	tasksM := tasks.NewTaskListExec(runCfg) // we create a task list

	code := tasksM.RunTarget("test", true) // we run the task
	if code != 107 {                       // we expect a code 107
		t.Errorf("Expected code 107, got %d", code)
	}

	msg := strings.Join(messages, "; ")
	if !strings.Contains(msg, "no requirement check handler set") {
		t.Errorf("Expected message 'no requirement check handler set' missing, got '%s'", msg)
	}

}

func TestWithRequirementCheck(t *testing.T) {
	// we create a slice to store the output
	messages := []string{}

	// create a outputhandler
	// we hook into the output handler to capture the output
	// if we get the MsgExecOutput message we append it to the messages slice
	// we use this to check the output
	outHandler := func(msg ...interface{}) {
		t.Logf("msg: %v", msg)
		for _, m := range msg {
			switch t := m.(type) {
			case tasks.MsgExecOutput: // this will be the output of the command
				messages = append(messages, string(t))
			}
		}

	}

	// we setup a task that will succeed
	var testTask configure.Task = configure.Task{
		ID:     "test",                // we need an ID
		Script: []string{"echo test"}, // we need a script
		Options: configure.Options{
			Displaycmd: true, // we want to see the command. we capture the output so we can check it
		},
	}
	// we create a tasklist with the task and the requirement check handler
	var runCfg configure.RunConfig = configure.RunConfig{}
	runCfg.Task = []configure.Task{testTask} // we add the task to the task list

	// we create a task list with the task and the requirement check handler
	tasksMain := tasks.NewTaskListExec(
		runCfg,
		tasks.NewDefaultDataHandler(),
		tasks.NewDefaultPhHandler(),
		outHandler,
		tasks.ShellCmd,
		func(require configure.Require) (bool, string) { return true, "" },
	)

	// execute the task and check the output
	code := tasksMain.RunTarget("test", false) // we run the task
	if code != 0 {                             // we expect a code 0
		t.Errorf("Expected code 0, got %d", code)
	}

	// verify the output contains the expected message
	msg := strings.Join(messages, "; ")
	if !strings.Contains(msg, "test") {
		t.Errorf("Expected message 'test' missing, got '%s'", msg)
	}
}

func TestMultipleTask(t *testing.T) {
	// we create a slice to store the output
	messages := []string{}

	// create a outputhandler
	// we hook into the output handler to capture the output
	// if we get the MsgExecOutput message we append it to the messages slice
	// we use this to check the output
	outHandler := func(msg ...interface{}) {
		t.Logf("msg: %v", msg)
		for _, m := range msg {
			switch mt := m.(type) {
			case tasks.MsgExecOutput: // this will be the output of the command
				t.Logf("cmd output: %v", msg)
				messages = append(messages, string(mt))
			}
		}

	}

	// we setup a task that will succeed
	var testTask configure.Task = configure.Task{
		ID:     "test",                // we need an ID
		Script: []string{"echo test"}, // we need a script
		Options: configure.Options{
			Displaycmd: true, // we want to see the command. we capture the output so we can check it
		},
	}
	// we create a tasklist with the task and the requirement check handler
	var runCfg configure.RunConfig = configure.RunConfig{}
	runCfg.Task = []configure.Task{testTask} // we add the task to the task list

	taskCopy := testTask
	taskCopy.Script = []string{"echo marmelade"}
	runCfg.Task = append(runCfg.Task, taskCopy) // we add the second task to the task list

	// another task that should not being executed
	var testTaskToIgnore configure.Task = configure.Task{
		ID:     "other",                // we need an ID
		Script: []string{"echo other"}, // we need a script
		Options: configure.Options{
			Displaycmd: true, // we want to see the command. we capture the output so we can check it
		},
	}

	runCfg.Task = append(runCfg.Task, testTaskToIgnore) // we add the task to the task list that should being executed

	// we create a task list with the task and the requirement check handler
	tasksMain := tasks.NewTaskListExec(
		runCfg,
		tasks.NewDefaultDataHandler(),
		tasks.NewDefaultPhHandler(),
		outHandler,
		tasks.ShellCmd,
		func(require configure.Require) (bool, string) { return true, "" },
	)

	// execute the task and check the output
	code := tasksMain.RunTarget("test", false) // we run the task
	if code != 0 {                             // we expect a code 0
		t.Errorf("Expected code 0, got %d", code)
	}

	// verify the output contains the expected message
	msg := strings.Join(messages, "; ") + ";"
	if !strings.Contains(msg, "test;") {
		t.Errorf("Expected message 'test' missing, got '%s'", msg)
	}

	if !strings.Contains(msg, "marmelade;") {
		t.Errorf("Expected message 'test' missing, got '%s'", msg)
	}

	if strings.Contains(msg, "other;") {
		t.Errorf("Unexpected message 'other', got '%s'", msg)
	}
}
