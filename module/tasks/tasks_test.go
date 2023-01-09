package tasks_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/tasks"
	"gopkg.in/yaml.v2"
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

func TestTargetWithNeeds(t *testing.T) {
	source := `
task:
  - id: test
    needs: [subtask]
    script:
      - echo test
  - id: subtask
    script: 
      - echo i-am-subtask
`
	var runCfg configure.RunConfig = configure.RunConfig{}

	if err := yaml.Unmarshal([]byte(source), &runCfg); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	}

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

	assert.Contains(t, messages, "test")
	assert.Contains(t, messages, "i-am-subtask")
	assert.Equal(t, "i-am-subtask; test", strings.Join(messages, "; "))

}

func TestTargetWithRunTargets(t *testing.T) {
	source := `
task:
  - id: test
    runTargets: [subtask]
    script:
      - echo test  
  - id: subtask
    script:
      - echo i-am-subtask
`
	var runCfg configure.RunConfig = configure.RunConfig{}

	if err := yaml.Unmarshal([]byte(source), &runCfg); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {

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

		assert.Contains(t, messages, "test")
		assert.Contains(t, messages, "i-am-subtask")

	}
}

func TestTargetWithNext(t *testing.T) {
	source := `
task:
  - id: test
    next: [subtask]
    script:
      - echo test  
  - id: subtask
    script:
      - echo i-am-subtask
`
	var runCfg configure.RunConfig = configure.RunConfig{}

	if err := yaml.Unmarshal([]byte(source), &runCfg); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {

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

		assert.Contains(t, messages, "test")
		assert.Contains(t, messages, "i-am-subtask")
		assert.Equal(t, "test; i-am-subtask", strings.Join(messages, "; "))
	}
}

func TestTargetComplexWith2Needs(t *testing.T) {
	source := `
task:
  - id: test
    needs: 
      - subtask
      - subtask2
    script:
      - echo test  
  - id: subtask
    script:
      - echo i-am-subtask
  - id: subtask2
    script:
      - echo i-am-subtask2
`
	var runCfg configure.RunConfig = configure.RunConfig{}

	if err := yaml.Unmarshal([]byte(source), &runCfg); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {

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

		assert.Contains(t, messages, "test")
		assert.Contains(t, messages, "i-am-subtask")
		assert.Contains(t, messages, "i-am-subtask2")
		assertContainsCount(t, messages, "test", 1)
		assertContainsCount(t, messages, "i-am-subtask", 1)
		assertContainsCount(t, messages, "i-am-subtask2", 1)
		assertPositionInSliceBefore(t, messages, "i-am-subtask", "test")
		assertPositionInSliceBefore(t, messages, "i-am-subtask2", "test")
	}
}

func TestTargetComplexWith2NestedNeeds(t *testing.T) {
	source := `
task:
  - id: test
    needs: 
      - subtask
      - subtask2
    script:
      - echo test  
  - id: subtask
    script:
      - echo i-am-subtask
  - id: subtask2
    needs:
      - subtask
    script:
      - echo i-am-subtask2
`
	var runCfg configure.RunConfig = configure.RunConfig{}

	if err := yaml.Unmarshal([]byte(source), &runCfg); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {

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

		code = tasksMain.RunTarget("test", false) // we run the task again
		if code != 0 {                            // we still expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}

		assert.Contains(t, messages, "test")
		assert.Contains(t, messages, "i-am-subtask")
		assert.Contains(t, messages, "i-am-subtask2")
		assertContainsCount(t, messages, "test", 2)         // we run the task twice
		assertContainsCount(t, messages, "i-am-subtask", 1) // any needs should not being executed twice
		assertContainsCount(t, messages, "i-am-subtask2", 1)
		assertPositionInSliceBefore(t, messages, "i-am-subtask", "test")
		assertPositionInSliceBefore(t, messages, "i-am-subtask2", "test")
	}
}

func TestTargetVariables(t *testing.T) {
	source := `
config:
    variables:
        test: "variable"
task:
  - id: test
    script:
      - echo {test}
`
	var runCfg configure.RunConfig = configure.RunConfig{}

	if err := yaml.Unmarshal([]byte(source), &runCfg); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {

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

		tasksMain := tasks.NewTaskListExec(
			runCfg,
			tasks.NewCombinedDataHandler(),
			outHandler,
			tasks.ShellCmd,
			func(require configure.Require) (bool, string) { return true, "" },
		)
		code := tasksMain.RunTarget("test", false) // we run the task
		if code != 0 {                             // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}

		assert.Contains(t, messages, "variable")

	}
}

func TestVariables(t *testing.T) {
	source := `
config:
    variables:
        test: "variable"
        second: "second"
task:
  - id: test
    script:
      - echo {test}
      - echo first-{second}
    next:
      - testAddVar

  - id: testAddVar
    variables:
        test: "variable2"
    script:
      - echo check-{test} 
      - echo second-{second} 
`
	messages := []string{}
	if taskMain, err := createRuntimeByYamlString(source, &messages); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("test", false) // we run the task
		if code != 0 {                            // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}

		assert.Contains(t, messages, "variable")
		assert.Contains(t, messages, "check-variable2")
		assert.Contains(t, messages, "first-second")
		assert.Contains(t, messages, "second-second")

	}

}
