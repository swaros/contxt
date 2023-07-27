package tasks_test

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/systools"
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
				messages = append(messages, string(t.Output))
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
	tasksMain := tasks.NewStdTaskListExec(
		runCfg,
		tasks.NewDefaultDataHandler(),
		tasks.NewDefaultPhHandler(),
		outHandler,
		tasks.ShellCmd,
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
				messages = append(messages, string(mt.Output))
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
	tasksMain := tasks.NewStdTaskListExec(
		runCfg,
		tasks.NewDefaultDataHandler(),
		tasks.NewDefaultPhHandler(),
		outHandler,
		tasks.ShellCmd,
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
				messages = append(messages, string(mt.Output))
			}
		}

	}
	tasksMain := tasks.NewStdTaskListExec(
		runCfg,
		tasks.NewDefaultDataHandler(),
		tasks.NewDefaultPhHandler(),
		outHandler,
		tasks.ShellCmd,
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
					messages = append(messages, string(mt.Output))
				}
			}

		}
		tasksMain := tasks.NewStdTaskListExec(
			runCfg,
			tasks.NewDefaultDataHandler(),
			tasks.NewDefaultPhHandler(),
			outHandler,
			tasks.ShellCmd,
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
					messages = append(messages, string(mt.Output))
				}
			}

		}
		tasksMain := tasks.NewStdTaskListExec(
			runCfg,
			tasks.NewDefaultDataHandler(),
			tasks.NewDefaultPhHandler(),
			outHandler,
			tasks.ShellCmd,
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
					messages = append(messages, string(mt.Output))
				}
			}

		}
		tasksMain := tasks.NewStdTaskListExec(
			runCfg,
			tasks.NewDefaultDataHandler(),
			tasks.NewDefaultPhHandler(),
			outHandler,
			tasks.ShellCmd,
		)
		// execute the task and check the output
		code := tasksMain.RunTarget("test", true) // we run the task in async mode
		if code != 0 {                            // we expect a code 0
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
					messages = append(messages, string(mt.Output))
				}
			}

		}
		tasksMain := tasks.NewStdTaskListExec(
			runCfg,
			outHandler,
			tasks.ShellCmd,
		)

		// execute the task and check the output
		code := tasksMain.RunTarget("test", true) // we run the task async
		if code != 0 {                            // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}

		code = tasksMain.RunTarget("test", true) // we run the task again async
		if code != 0 {                           // we still expect a code 0
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
      - echo ${test}
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
					messages = append(messages, string(mt.Output))
				}
			}

		}

		tasksMain := tasks.NewStdTaskListExec(
			runCfg,
			outHandler,
			tasks.ShellCmd,
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
      - echo ${test}
      - echo first-${second}
    next:
      - testAddVar

  - id: testAddVar
    variables:
        test: "variable2"
    script:
      - echo check-${test} 
      - echo second-${second} 
`
	messages := []string{}
	if taskMain, err := createRuntimeByYamlString(source, &messages); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("test", true) // we run the task
		if code != 0 {                           // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}

		assert.Contains(t, messages, "variable")
		assert.Contains(t, messages, "check-variable2")
		assert.Contains(t, messages, "first-second")
		assert.Contains(t, messages, "second-second")

	}

}

func TestTryParse(t *testing.T) {
	source := `
config:
    variables:
        default_1: "variable"
        default_2: "second"
task:
  - id: test
    script:
      - "#@set default_1 rewrite"
      - echo ${default_1}
      - "#@var default_2 echo new-world"
      - echo ${default_2}
`
	messages := []string{}
	if taskMain, err := createRuntimeByYamlString(source, &messages); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("test", true) // we run the task
		if code != 0 {                           // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}

	}
	assert.Contains(t, messages, "rewrite")
	assert.Contains(t, messages, "new-world")
}

func TestCheats(t *testing.T) {
	source := `
config:
    variables:
        default_1: "variable"
        default_2: "second"
        count: "5"
task:
  - id: test
    require:
        variables:
            default_1: "variable"
            default_2: "second"
    script:
        - echo ${default_1}
        - echo ${default_2}
    
  - id: testCount1
    require:
        variables:
            count: ">4"
    script:
        - echo result_1_${count}

  - id: testCount2
    require:
        variables:
            count: "<6"
    script:
        - echo result_2_${count}

  - id: testCount2
    require:
        variables:
            count: "!6"
    script:
      - echo result_3_${count}


  - id: testCount2
    require:
        variables:
            count: "=7"
    script:
        - echo should_not_be_shown

  - id: testCount2
    require:
        variables:
            count: "*"
    script:
      - echo result_4_${count}
    next:
      - testVarRewrite
  - id: testVarRewrite
    script:
        - "#@set default_1 rewrite"
        - echo new_default_1_${default_1}

  - id: testSomeCheatMacros
    variables:
       json_content: '{"key1": "value1", "key2": "value2"}'
    needs:
        - testVarRewrite
    script:
        - "#@var default_2 echo new-world"
        - echo new_default_2_${default_2}
        - echo just_to_be_sure_${default_1}
        - "#@if-equals default_1 variable"
        - echo "default_1 is variable"
        - "#@end"
        - "#@if-equals rewrite rewrite"
        - echo "default_1 is rewrite"
        - "#@end"
        - echo default2_recheck_${default_2}
        - "#@if-equals ${default_2} new-world"
        - echo "default_2 is new-world"
        - "#@end"
        - "#@if-equals ${default_2} !new-world"
        - echo "we should not see this message No1"
        - "#@end"
        - "#@if-not-equals ${default_2} new-world"
        - echo "we should not see this message No2"
        - "#@end"
        - "#@import-json JSON-CONTENT '${json_content}'" 
        - echo "key1 is ${JSON-CONTENT:key1}" 
        - "#@import-json JSON-CONTENT-2 ${json_content}" 
        - echo "key1 is again ${JSON-CONTENT-2:key1}"
        - "#@add default_1 +++"
        - echo "after add value ${default_1}" 
`
	messages := []string{}
	if taskMain, err := createRuntimeByYamlString(source, &messages); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("test", true) // we run the task
		if code != 0 {                           // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}

		assert.Contains(t, messages, "variable")
		assert.Contains(t, messages, "second")

		// testing count requires
		code = taskMain.RunTarget("testCount1", true) // we run the task to testing count requires
		if code != 0 {                                // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}
		assert.Contains(t, messages, "result_1_5")

		// here we test a whole chain of tasks
		// we expect that the variable will be rewritten
		code = taskMain.RunTarget("testCount2", true) // we run the task to testing count requires
		if code != 0 {                                // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}
		assert.Contains(t, messages, "result_2_5")
		assert.Contains(t, messages, "result_3_5")
		assert.Contains(t, messages, "result_4_5")
		assert.Contains(t, messages, "new_default_1_rewrite")
		assert.NotContains(t, messages, "should_not_be_shown")

		code = taskMain.RunTarget("testSomeCheatMacros", true) // we run the task to testing count requires
		if code != 0 {                                         // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}
		assert.Contains(t, messages, "new_default_2_new-world")
		assert.Contains(t, messages, "just_to_be_sure_rewrite")
		assert.NotContains(t, messages, "default_1 is variable")
		assert.Contains(t, messages, "default_1 is rewrite")
		assert.Contains(t, messages, "default2_recheck_new-world")
		assert.Contains(t, messages, "default_2 is new-world")
		assert.NotContains(t, messages, "we should not see this message No1")
		assert.NotContains(t, messages, "we should not see this message No2")
		assert.Contains(t, messages, "key1 is value1")
		assert.Contains(t, messages, "key1 is again value1")
		assert.Contains(t, messages, "after add value rewrite+++")
	}

}

func TestTryParseCheatsWithErrors(t *testing.T) {
	// testing the errorcases of the cheats
	source := `
config:
    variables:
        default_1: "variable"
        default_2: "second"
        count: "5"
task:
  - id: test
    script:
      - "#@if-equals default_1"

  - id: failure_1
    script:
      - "#@if-equals default_1 variable"
      - "#@if-equals b a"
  - id: failure_2
    script:
        - "#@if-not-equals default_1"
  - id: failure_3
    script:
        - "#@if-equals default_1 variable"
        - "#@if-not-equals default_1 variable"
  - id: failure_4
    script:
        - "#@import-json"
  - id: failure_5
    script:
        - "#@import-json JSON-DDD {this-is-not-json hello world"
  - id: failure_6
    script:
        - "#@import-json-exec FAILTEST echo {this-is-not-json hello world"
  - id: failure_7
    script:
        - "#@import-json-exec FAILTEST2"
  - id: failure_8
    script:
        - "#@import-json-exec FAILTEST3 invalidcmd blub"
  - id: failure_9
    script:
        - "#@add"
  - id: failure_10
    script:    
        - "#@set"
  - id: failure_11
    script:
        - "#@set-in-map"
  - id: failure_12 
    script:
        - "#@set-in-map notexists key value"
  - id: failure_13
    script:
        - "#@var-to-file test key"
  - id: failure_14
    script:
        - "#@var-to-file test"
  - id: failure_15
    script:
        - "#@export-to-json"
  - id: failure_16 
    script:
        - "#@export-to-json notexists some"
  - id: failure_17
    script:
        - "#@export-to-yaml"
  - id: failure_18
    script:
        - "#@export-to-yaml notexists some"
  - id: failure_19
    script:
        - "#@var checkx notexistscmd"
  - id: failure_20
    script:
        - "#@var"
  - id: failure_21
    script:
        - "#@add nonextsting something"


`
	messages := []string{}
	errorMsg := []error{}
	if taskMain, err := createRuntimeByYamlStringWithErrors(source, &messages, &errorMsg); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {

		type TestRuns struct {
			target        string
			expectedCode  int
			expectedError string
			linuxOnly     bool
		}

		testRuns := []TestRuns{
			{target: "test", expectedCode: 8, expectedError: "invalid usage #@if-equals need: str1 str2"},
			{target: "failure_1", expectedCode: 8, expectedError: "invalid usage #@if-equals can not be used in another if"},
			{target: "failure_2", expectedCode: 8, expectedError: "invalid usage #@if-not-equals need: str1 str2"},
			{target: "failure_3", expectedCode: 8, expectedError: "invalid usage #@if-not-equals can not be used in another if"},
			{target: "failure_4", expectedCode: 8, expectedError: "invalid usage #@import-json needs 2 arguments. <keyname> <json-source-string>"},
			{target: "failure_5", expectedCode: 8, expectedError: "error while parsing json: invalid character 't' looking for beginning of object key string"},
			{target: "failure_6", expectedCode: 8, expectedError: "error while parsing json: invalid character 't' looking for beginning of object key string", linuxOnly: true},
			{target: "failure_7", expectedCode: 8, expectedError: "invalid usage #@import-json-exec needs 2 arguments at least. <keyname> <bash-command>"},
			{target: "failure_8", expectedCode: 8, expectedError: "error while executing command: exit status 127", linuxOnly: true},
			{target: "failure_9", expectedCode: 8, expectedError: "invalid usage #@add needs 2 arguments at least. <keyname> <value>"},
			{target: "failure_10", expectedCode: 8, expectedError: "invalid usage #@set needs 2 arguments at least. <keyname> <value>"},
			{target: "failure_11", expectedCode: 8, expectedError: "invalid usage #@set-in-map needs 3 arguments at least. <mapName> <json.path> <value>"},
			{target: "failure_12", expectedCode: 8, expectedError: "error while setting value in map: the key [notexists] does not exists"},
			{target: "failure_13", expectedCode: 8, expectedError: "error while writing variable to file: variable 'test' can not be used for export to file. this variable not exists"},
			{target: "failure_14", expectedCode: 8, expectedError: "invalid usage #@var-to-file needs 2 arguments at least. <variable> <filename>"},
			{target: "failure_15", expectedCode: 8, expectedError: "invalid usage #@export-to-json needs 2 arguments at least. <map-key> <variable>"},
			{target: "failure_16", expectedCode: 8, expectedError: "map with key notexists not exists"},
			{target: "failure_17", expectedCode: 8, expectedError: "invalid usage #@export-to-yaml needs 2 arguments at least. <map-key> <variable>"},
			{target: "failure_18", expectedCode: 8, expectedError: "map with key notexists not exists"},
			{target: "failure_19", expectedCode: 8, expectedError: "error while executing command: exit status 127", linuxOnly: true},
			{target: "failure_20", expectedCode: 8, expectedError: "invalid usage #@var needs 2 arguments at least. <varibale-name> <bash-command>", linuxOnly: true},
			{target: "failure_21", expectedCode: 8, expectedError: "variable must exists for add #@add nonextsting"},
		}

		for i, testRun := range testRuns {

			if (testRun.linuxOnly && runtime.GOOS == "linux") || !testRun.linuxOnly {
				// reset the messages and error messages
				messages = []string{}
				errorMsg = []error{}

				code := taskMain.RunTarget(testRun.target, true) // we run the task
				if code != testRun.expectedCode {                // we expect a code 0
					t.Errorf("Expected code %d, got %d", testRun.expectedCode, code)
				}
				assert.Contains(
					t,
					errorMsg,
					errors.New(testRun.expectedError),
					"Error message not found for target %s in round %v  -> %v",
					testRun.target,
					i,
					errorMsg,
				)
			}
		}

	}
}

func TestTriggerExecution(t *testing.T) {
	source := `
version: "1"
task:
  - id: base
    script:
      - echo "triggered"
    listener:
      - trigger: 
          onoutContains: 
            - triggered
        action:
            script: 
              - echo "reaction"


`
	messages := []string{}
	errorMsg := []error{}
	if taskMain, err := createRuntimeByYamlStringWithErrors(source, &messages, &errorMsg); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("base", true) // we run the task
		if code != 0 {                           // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}
		assert.Contains(t, messages, "triggered", "triggered not found in messages")
		assert.Contains(t, messages, "reaction", "reaction not found in messages")
	}
}

func TestTriggerExecutionWithTarget(t *testing.T) {
	source := `
version: "1"
task:
  - id: subTarget
    script:
        - echo "reaction"
  - id: base
    script:
      - echo "triggered"
    listener:
      - trigger: 
          onoutContains: 
            - triggered
        action:
            target: subTarget
              
`
	messages := []string{}
	errorMsg := []error{}
	if taskMain, err := createRuntimeByYamlStringWithErrors(source, &messages, &errorMsg); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("base", true) // we run the task
		if code != 0 {                           // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}
		assert.Contains(t, messages, "triggered", "triggered not found in messages")
		assert.Contains(t, messages, "reaction", "reaction not found in messages")
	}
}

func assertErrormsgContains(t *testing.T, expected string, actual []error) bool {
	t.Helper()
	for _, msg := range actual {
		if strings.Contains(msg.Error(), expected) {
			return true // all fine
		}
	}
	formsg := []string{}
	for _, msg := range actual {
		formsg = append(formsg, msg.Error())
	}
	_, file, line, _ := runtime.Caller(1)
	label := fmt.Sprintf("%s:%d", file, line)
	t.Errorf(label+" Error message not found:\n%s\nin\n[%s]", expected, strings.Join(formsg, ", "))
	return false
}

func TestTriggerError(t *testing.T) {
	source := `
version: "1"
task:
  - id: subTarget
    script:
        - some-not-working-command
  - id: base
    script:
      - echo "triggered"
    listener:
      - trigger: 
          onoutContains: 
            - triggered
        action:
            target: subTarget
              
`
	messages := []string{}
	errorMsg := []error{}
	if taskMain, err := createRuntimeByYamlStringWithErrors(source, &messages, &errorMsg); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("base", true) // we run the task
		if code != 0 {                           // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}
		assert.Contains(t, messages, "triggered", "triggered not found in messages")
		if runtime.GOOS == "windows" {
			assertErrormsgContains(t, "some-not-working-command fails with error: exit status 1", errorMsg)
		} else {
			assertErrormsgContains(t, "some-not-working-command fails with error: exit status 127", errorMsg)
		}

	}
}

func TestTriggerErrorIgnored(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on windows")
	}
	source := `
version: "1"
task:
  - id: subTarget
    options:
      ignoreCmdError:  true
    script:
        - echo "blub" | grep "la" # this will fail
  - id: base
    
    script:
      - echo "triggered"
    listener:
      - trigger: 
          onoutContains: 
            - triggered
        action:
            target: subTarget
              
`
	messages := []string{}
	errorMsg := []error{}
	if taskMain, err := createRuntimeByYamlStringWithErrors(source, &messages, &errorMsg); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("base", true) // we run the task
		if code != 0 {                           // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}
		assert.Contains(t, messages, "triggered", "triggered not found in messages")
		assertErrormsgContains(t, "exit status 1", errorMsg)

	}
}

func TestRunError2(t *testing.T) {
	source := `
version: "1"
task:
  - id: base
    script:
      - me-is-already-not-working
`
	messages := []string{}
	errorMsg := []error{}
	targetUpdates := []string{}
	if taskMain, err := createRuntimeByYamlStringWithAllMsg(source, &messages, &errorMsg, nil, &targetUpdates); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("base", true) // we run the task
		if code != systools.ExitCmdError {       // we expect a error code similar to 103
			t.Errorf("Expected code 103, got %d", code)
		}
		if runtime.GOOS == "windows" {
			assertErrormsgContains(t, "me-is-already-not-working fails with error: exit status 1", errorMsg)
		} else {
			assertErrormsgContains(t, "me-is-already-not-working fails with error: exit status 127", errorMsg)
		}

	}
}

func TestTriggerExecutionStopReason(t *testing.T) {
	source := `
version: "1"
task:
  - id: base
    script:
      - echo "start-loop"
      - while true; do echo "loop"; sleep 1; done      
    stopreasons:
        onoutContains:
         - loop         
              
`
	if runtime.GOOS == "windows" { // we skip this test on windows, because the stopreasons are not working on windows depending on the shell while loop
		t.Skip("Skip on windows")
	}
	messages := []string{}
	errorMsg := []error{}
	if taskMain, err := createRuntimeByYamlStringWithErrors(source, &messages, &errorMsg); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("base", true) // we run the task
		if code != systools.ExitByStopReason {   // we expect the run is aborted by stopreason
			t.Errorf("Expected code 101, got %d", code)
		}
		assert.Contains(t, messages, "start-loop", "start-loop not found in messages")
		assert.NotContains(t, messages, "loop", "loop not found in messages") // the trigger source itself should not be in the output
	}
}

func TestTriggerInSomeCombinations(t *testing.T) {
	source := `
version: "1"
task:
    - id: subTarget
      options:
        displaycmd: true
      script:
        - echo "reaction"
    - id: base1
      options:
        displaycmd: true
      script:
          - echo "triggered_base1"
      listener:
        - trigger:
            onoutContains:
              - triggered
          action:
            target: subTarget
    - id: base2
      script:
        - echo "triggered_base2"
      listener:
        - trigger:
            onoutContains:
             - triggered
          action:
            target: subTarget
    - id: base3
      script:
        - echo "triggered_base3"
      listener:
        - trigger:
           onoutContains:
             - triggered

    - id: base4
      script:
        - echo "triggered_base4"
      listener:
          - trigger:
              onoutContains:
                - triggered
            action:
                target: notExists

    - id: base5
      needs:                
        - base1        
      script:
        - echo "triggered_base5"

    - id: base6
      options:
        displaycmd: true
        ignoreCmdError: true
        stickcursor: true
      script:
        - echo "triggered_base6"
        - echo "lala" | grep "oops"
      stopreasons:
         onerror: true

    - id: base7
      options:
        displaycmd: true
        ignoreCmdError: true
        stickcursor: true
        maincmd: not-existing-command
      script:
        - will not work anyway
        
`
	// the whitelist is used to linit the tests to some targets for debugging.
	// if some api changes happens, it is hard to step through all tests.
	// so this list should not have any content if this file is checked in.
	// (if this slice is empty, it will be ignored)
	whiteList := []string{}
	messages := []string{}
	errorMsg := []error{}
	targetUpdates := []string{}
	if taskMain, err := createRuntimeByYamlStringWithAllMsg(source, &messages, &errorMsg, nil, &targetUpdates); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		type TestRuns struct {
			target           string
			expectedCode     int
			expectedError    string
			expectInMessages string
			targetUpdate     string
			linuxOnly        bool
		}

		testRuns := []TestRuns{
			{target: "base1", expectedCode: 0, expectInMessages: "reaction"},
			{target: "base2", expectedCode: 0, expectInMessages: "reaction"},
			{target: "base3", expectedCode: 0, expectedError: "trigger defined without any action"},
			{target: "base4", expectedCode: 0, targetUpdate: "notExists:not_found[]"},
			{target: "base5", expectedCode: 0, expectInMessages: "reaction"},
			{target: "base6", expectedCode: 101, expectInMessages: "exit status 1", expectedError: "exit status 1", linuxOnly: true},
			{target: "base7", expectedCode: 102, expectedError: "exec: \"not-existing-command\": executable file not found in $PATH", linuxOnly: true},
		}
		taskMain.SetHardExistToAllTasks(false) // we set the hard exit to false, so we can test the exit codes
		for _, testRun := range testRuns {
			if len(whiteList) == 0 || systools.StringInSlice(testRun.target, whiteList) {
				if (testRun.linuxOnly && runtime.GOOS == "linux") || !testRun.linuxOnly {
					// reset the messages and error messages
					messages = []string{}
					errorMsg = []error{}
					targetUpdates = []string{}

					code := taskMain.RunTarget(testRun.target, true) // we run the task
					if code != testRun.expectedCode {                // we expect a code 0
						t.Errorf("Expected code %d, got %d", testRun.expectedCode, code)
					}
					// if we expect a error message, we check if it is in the error messages
					if testRun.expectedError != "" {
						if alfine := assertErrormsgContains(t, testRun.expectedError, errorMsg); !alfine {
							t.Errorf("Failed Test for Target [%s] (check expected error message)", testRun.target)
						}
					}
					// if we expect a message, we check if it is in the messages
					if testRun.expectInMessages != "" {
						if allFine := assert.Contains(
							t,
							messages,
							testRun.expectInMessages,
							testRun.expectInMessages+" not found in messages ["+strings.Join(messages, ",")+"]"); !allFine {
							t.Errorf("Failed Test for Target [%s] (check expected message)", testRun.target)
						}
					}

					if testRun.targetUpdate != "" {
						if allFine := assert.Contains(
							t,
							targetUpdates,
							testRun.targetUpdate,
							testRun.targetUpdate+" not found in targetUpdates ["+strings.Join(targetUpdates, ",")+"]"); !allFine {
							t.Errorf("Failed Test for Target [%s] (check expected target Update)", testRun.target)
						}
					}
				}
			}
		}
	}

}

func TestRequire(t *testing.T) {
	source := `
version: "1"
config:
    variables:
       check: "hello"
task:
  - id: subTarget
    require:
       variables:
          check: "hello" 
    options:
        displaycmd: true
    script:
        - echo "reaction"
`

	messages := []string{}
	errorMsg := []error{}
	targetUpdates := []string{}
	if taskMain, err := createRuntimeByYamlStringWithAllMsg(source, &messages, &errorMsg, nil, &targetUpdates); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("subTarget", true) // we run the task
		if code != 0 {                                // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}

		assert.Contains(t, messages, "reaction", "reaction not found in messages ["+strings.Join(messages, ",")+"]")
		assert.Contains(t, targetUpdates, "subTarget:command[echo \"reaction\"]", "subTarget:command... not found in targetUpdates ["+strings.Join(targetUpdates, ",")+"]")
	}
}

func TestRequire2(t *testing.T) {
	source := `
version: "1"
config:
    variables:
       check: "hello"
task:
  - id: subTarget
    require:
       variables:
          check: "!winter" 
    options:
        displaycmd: true
    script:
        - echo "reaction"
`

	messages := []string{}
	errorMsg := []error{}
	targetUpdates := []string{}
	if taskMain, err := createRuntimeByYamlStringWithAllMsg(source, &messages, &errorMsg, nil, &targetUpdates); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("subTarget", true) // we run the task
		if code != 0 {                                // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}

		assert.Contains(t, messages, "reaction", "reaction not found in messages ["+strings.Join(messages, ",")+"]")
		assert.Contains(t, targetUpdates, "subTarget:command[echo \"reaction\"]", "subTarget:command... not found in targetUpdates ["+strings.Join(targetUpdates, ",")+"]")
	}
}

func TestRequire3(t *testing.T) {
	source := `
task:
  - id: subTarget
    require:
       system: !windows          
    options:
        displaycmd: true
    script:
        - echo "hello other then linux"
  - id: subTarget
    require:
        system: windows         
    options:
        displaycmd: true
    script:
        - echo "hello windows"

`

	messages := []string{}
	errorMsg := []error{}
	targetUpdates := []string{}
	if taskMain, err := createRuntimeByYamlStringWithAllMsg(source, &messages, &errorMsg, nil, &targetUpdates); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("subTarget", true) // we run the task
		if code != 0 {                                // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}
		if runtime.GOOS == "windows" {
			assert.Contains(t, messages, "hello windows", "'hello windows' found in messages ["+strings.Join(messages, ",")+"]")
			assert.Contains(t, targetUpdates, "subTarget:command[echo \"hello windows\"]", "subTarget:command... not found in targetUpdates ["+strings.Join(targetUpdates, ",")+"]")
		} else {
			assert.Contains(t, messages, "hello other then linux", "hello other then linux not found in messages ["+strings.Join(messages, ",")+"]")
			assert.Contains(t, targetUpdates, "subTarget:command[echo \"hello other then linux\"]", "subTarget:command... not found in targetUpdates ["+strings.Join(targetUpdates, ",")+"]")
		}
	}
}

func TestRequire4(t *testing.T) {
	source := `
version: "1"
config:
    variables:
       check: "hello"
task:
  - id: subTarget
    require:
       exists: 
         - "test.txt"
         
    options:
        displaycmd: true
    script:
        - echo "reaction"
`

	messages := []string{}
	errorMsg := []error{}
	targetUpdates := []string{}
	if taskMain, err := createRuntimeByYamlStringWithAllMsg(source, &messages, &errorMsg, nil, &targetUpdates); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("subTarget", true) // we run the task
		if code != 107 {                              // we expect a code 107
			t.Errorf("Expected code 107, got %d", code)
		}

		assert.Contains(t, targetUpdates, "subTarget:requirement-check-failed[required file (test.txt) not found ]", ".subTarget:crequirement.. not found in targetUpdates ["+strings.Join(targetUpdates, ",")+"]")
	}
}

func TestRequire5(t *testing.T) {
	source := `
version: "1"
config:
    variables:
       check: "hello"
task:
  - id: subTarget
    require:
       exists: 
         - "tasks_test.go"
         
    options:
        displaycmd: true
    script:
        - echo "reaction"
`

	messages := []string{}
	errorMsg := []error{}
	targetUpdates := []string{}
	if taskMain, err := createRuntimeByYamlStringWithAllMsg(source, &messages, &errorMsg, nil, &targetUpdates); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("subTarget", true) // we run the task
		if code != 0 {                                // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}
		assert.Contains(t, messages, "reaction", "reaction not found in messages ["+strings.Join(messages, ",")+"]")
		assert.Contains(t, targetUpdates, "subTarget:wait_next_done[index 0]", ".subTarget:command.. not found in targetUpdates ["+strings.Join(targetUpdates, ",")+"]")
	}
}

func TestRequire6(t *testing.T) {
	source := `
version: "1"
config:
    variables:
       check: "hello"
task:
  - id: subTarget
    require:
       notExists: 
         - "lulu.chk"
         
    options:
        displaycmd: true
    script:
        - echo "reaction"
`

	messages := []string{}
	errorMsg := []error{}
	targetUpdates := []string{}
	if taskMain, err := createRuntimeByYamlStringWithAllMsg(source, &messages, &errorMsg, nil, &targetUpdates); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("subTarget", true) // we run the task
		if code != 0 {                                // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}
		assert.Contains(t, messages, "reaction", "reaction not found in messages ["+strings.Join(messages, ",")+"]")
		assert.Contains(t, targetUpdates, "subTarget:wait_next_done[index 0]", ".subTarget:command.. not found in targetUpdates ["+strings.Join(targetUpdates, ",")+"]")
	}
}

func TestRequire7(t *testing.T) {
	source := `
version: "1"
config:
    variables:
       check: "hello"
task:
  - id: subTarget
    require:
       notExists: 
         - tasks_test.go
         
    options:
        displaycmd: true
    script:
        - echo "reaction"
`

	messages := []string{}
	errorMsg := []error{}
	targetUpdates := []string{}
	if taskMain, err := createRuntimeByYamlStringWithAllMsg(source, &messages, &errorMsg, nil, &targetUpdates); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("subTarget", true) // we run the task
		if code != 107 {                              // we expect a code 0
			t.Errorf("Expected code 107, got %d", code)
		}
		assert.Contains(t, targetUpdates, "subTarget:requirement-check-failed[unexpected file (tasks_test.go)  found ]", ".subTarget:req... not found in targetUpdates ["+strings.Join(targetUpdates, ",")+"]")
	}
}

func TestRequire8(t *testing.T) {
	if os.Getenv("OLDPWD") == "" {
		t.Skip("OLDPWD is empty")
	}
	source := `
version: "1"
task:
  - id: subTarget
    require:
       environment:
          OLDPWD: "!hello"
         
    options:
        displaycmd: true
    script:
        - echo "works"
`

	messages := []string{}
	errorMsg := []error{}
	targetUpdates := []string{}
	if taskMain, err := createRuntimeByYamlStringWithAllMsg(source, &messages, &errorMsg, nil, &targetUpdates); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("subTarget", true) // we run the task
		if code != 0 {                                // we expect a code 0
			t.Errorf("Expected code 0, got %d", code)
		}
		assert.Contains(t, messages, "works", "'works' not found in messages ["+strings.Join(messages, ",")+"]")
		assert.Contains(t, targetUpdates, "subTarget:command[echo \"works\"]", ".subTarget:command.. not found in targetUpdates ["+strings.Join(targetUpdates, ",")+"]")
	}
}

func TestRequire9(t *testing.T) {
	if os.Getenv("OLDPWD") == "" {
		t.Skip("OLDPWD is empty")
	}
	if os.Getenv("OLDPWD") == "hello" {
		t.Skip("OLDPWD is hello")
	}
	source := `
version: "1"
task:
  - id: subTarget
    require:
       environment:
          OLDPWD: "hello"
         
    options:
        displaycmd: true
    script:
        - echo "works"
`

	messages := []string{}
	errorMsg := []error{}
	targetUpdates := []string{}
	if taskMain, err := createRuntimeByYamlStringWithAllMsg(source, &messages, &errorMsg, nil, &targetUpdates); err != nil {
		t.Errorf("Error parsing yaml: %v", err)
	} else {
		code := taskMain.RunTarget("subTarget", true) // we run the task
		if code != 107 {
			t.Errorf("Expected code 0, got %d", code)
		}

		assert.Contains(t, targetUpdates, "subTarget:requirement-check-failed[environment variable[OLDPWD] not matching with hello]", ".subTarget:command.. not found in targetUpdates ["+strings.Join(targetUpdates, ",")+"]")
	}
}
