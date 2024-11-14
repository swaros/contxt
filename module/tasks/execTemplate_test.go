package tasks_test

import (
	"strings"
	"testing"
	"time"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
)

func TestTemplateExec(t *testing.T) {
	localMsg := []string{}
	outHandler := func(msg ...interface{}) {
		for _, m := range msg {
			switch s := m.(type) {
			case string:
				localMsg = append(localMsg, s)

			case tasks.MsgExecOutput:
				localMsg = append(localMsg, string(s.Output))
			case tasks.MsgError:
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Script: []string{
					"echo Hello",
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)

	code := tsk.RunTarget("test", false)
	if code != 0 {
		t.Error("expected code 0 but got", code)
	}
	if len(localMsg) != 1 {
		t.Error("expected 1 message but got", len(localMsg))
	} else {
		if strings.TrimRight(localMsg[0], "\n") != "Hello" {
			t.Error("expected message 'Hello' but got", localMsg[0])
		}
	}
}

func TestTepmlateRunBasic(t *testing.T) {
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Script: []string{
					"echo Hello",
				},
			},
		},
	}
	_, _, logger := RunTargetHelperWithErrors(t, "test", runCfg, false, systools.ExitOk, -1, []string{}, true)
	if len(logger.debugs) != 5 {
		t.Error("expected 5 debug message but got", len(logger.debugs))
		helpLogSlice(t, logger.debugs)
	} else {
		assertSliceContainsAtLeast(t, logger.debugs, "findOrCreateTask: target already created in subTasks?%!(EXTRA string=test, bool=false)", 1)
		assertSliceContainsAtLeast(t, logger.debugs, "executeTemplate next definition%!(EXTRA mimiclog.Fields=map[current-target:test nexts:[]])", 1)
		assertSliceContainsAtLeast(t, logger.debugs, "onError:false onLess:0 onMore:0 testing-at:Hello", 1)
	}
}

func TestTepmlateWithRequire(t *testing.T) {
	// this requirement will not match
	requireVars := map[string]string{
		"test": "Hello",
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test-require",
				Requires: configure.Require{
					Variables: requireVars,
				},
				Script: []string{
					"echo Hello",
				},
			},
		},
	}
	// because the requirements are not matching, we end up with exit code ExitByNothingToDo (107)
	_, _, logger := RunTargetHelperWithErrors(t, "test-require", runCfg, false, systools.ExitByNothingToDo, 0, []string{}, true)

	assertSliceContainsAtLeast(t, logger.infos, "executeTemplate IGNORE because requirements not matching", 1)

}

func TestVersionCheck(t *testing.T) {
	expectedErrorcode := systools.ExitByUnsupportedVersion
	expectedMessageCount := 0
	configure.SetVersion("0", "1", "0")
	localMsg := []string{}
	outHandler := func(msg ...interface{}) {
		for _, m := range msg {
			switch s := m.(type) {
			case string:
				localMsg = append(localMsg, s)

			case tasks.MsgExecOutput:
				localMsg = append(localMsg, string(s.Output))
			case tasks.MsgError:
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Version: "99.99.99", // don't think this will be sometime a valid version
		Task: []configure.Task{
			{
				ID: "test",
				Script: []string{
					"echo Hello",
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)

	code := tsk.RunTarget("test", false)
	if code != expectedErrorcode {
		t.Error("expected code ", expectedErrorcode, " but got", code)
	}
	if len(localMsg) != expectedMessageCount {
		t.Error("expected ", expectedMessageCount, " messages, but got", len(localMsg))
	}
}

func TestTemplateExecMoreLines(t *testing.T) {
	localMsg := []string{}
	outHandler := func(msg ...interface{}) {
		for _, m := range msg {
			switch s := m.(type) {
			case string:
				localMsg = append(localMsg, s)

			case tasks.MsgExecOutput:
				localMsg = append(localMsg, string(s.Output))
			case tasks.MsgError:
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Script: []string{
					"echo Hello",
					"echo World",
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)

	code := tsk.RunTarget("test", false)
	if code != 0 {
		t.Error("expected code 0 but got", code)
	}
	if len(localMsg) != 2 {
		t.Error("expected 1 message but got", len(localMsg))
	} else {
		if strings.TrimRight(localMsg[0], "\n") != "Hello" {
			t.Error("expected message 'Hello' but got", localMsg[0])
		}
		if strings.TrimRight(localMsg[1], "\n") != "World" {
			t.Error("expected message 'World' but got", localMsg[1])
		}
	}
}

func TestTemplateExecUnknowTask(t *testing.T) {
	outHandler := func(msg ...interface{}) {
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Script: []string{
					"echo Hello",
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("not_exists", false)
	if code != 106 {
		t.Error("expected code 106 but got", code)
	}

}

func TestTemplateAnko(t *testing.T) {
	localMsg := []string{}
	outHandler := func(msg ...interface{}) {
		for _, m := range msg {
			switch s := m.(type) {
			case string:
				localMsg = append(localMsg, s)

			case tasks.MsgExecOutput:
				localMsg = append(localMsg, string(s.Output))
			case tasks.MsgError:
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					"print('Hello World')",
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)
	if code != 0 {
		t.Error("expected code 0 but got", code)
	}

	if len(localMsg) != 1 {
		t.Error("expected 1 message but got", len(localMsg))
	} else {
		if strings.TrimRight(localMsg[0], "\n") != "Hello World" {
			t.Error("expected message 'Hello World' but got", localMsg[0])
		}
	}

}

func TestTemplateAnkoReqiurements(t *testing.T) {
	localMsg := []string{}
	outHandler := func(msg ...interface{}) {
		for _, m := range msg {
			switch s := m.(type) {
			case string:
				localMsg = append(localMsg, s)

			case tasks.MsgExecOutput:
				localMsg = append(localMsg, string(s.Output))
			case tasks.MsgError:
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Requires: configure.Require{
					System: "not_exists",
				},
				Cmd: []string{
					"print('Hello World')",
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)
	if code != 107 {
		t.Error("expected code 107 but got", code)
	}

	if len(localMsg) != 0 {
		t.Error("expected 0 message but got", len(localMsg))
	}
}

func TestAnkoListener(t *testing.T) {
	expectedExitCode := 0
	expectedMessageCount := 5
	localMsg := []string{}
	outHandler := func(msg ...interface{}) {
		for _, m := range msg {
			switch s := m.(type) {
			case string:
				localMsg = append(localMsg, s)

			case tasks.MsgExecOutput:
				localMsg = append(localMsg, string(s.Output))
			case tasks.MsgError:
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Listener: []configure.Listener{
					{
						Trigger: configure.Trigger{
							OnoutContains: []string{"trigger-exec-1"},
						},
						Action: configure.Action{
							Target: "other-test",
						},
					},
				},
				Cmd: []string{
					"println('Hello World')",
					"println('trigger-exec-1')",
					"println('i am done')",
				},
			},
			{
				ID: "other-test",
				Cmd: []string{
					"println('i am the other test')",
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

	if len(localMsg) != expectedMessageCount {
		t.Error("expected ", expectedMessageCount, " message but got", len(localMsg))
		t.Log(localMsg)
	}
}

func TestListenerTrimSpaces(t *testing.T) {
	expectedExitCode := 0
	expectedMessageCount := 5
	localMsg := []string{}
	outHandler := func(msg ...interface{}) {
		for _, m := range msg {
			switch s := m.(type) {
			case string:
				localMsg = append(localMsg, s)

			case tasks.MsgExecOutput:
				localMsg = append(localMsg, string(s.Output))
			case tasks.MsgError:
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Listener: []configure.Listener{
					{
						Trigger: configure.Trigger{
							OnoutContains: []string{"trigger-exec-1"},
						},
						Action: configure.Action{
							Target: " other-test ", // for some reasons, we have spaces arounf the target. this needs to trim
						},
					},
				},
				Cmd: []string{
					"println('Hello World')",
					"println('trigger-exec-1')",
					"println('i am done')",
				},
			},
			{
				ID: "other-test",
				Cmd: []string{
					"println('i am the other test')",
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

	if len(localMsg) != expectedMessageCount {
		t.Error("expected ", expectedMessageCount, " message but got", len(localMsg))
		t.Log(localMsg)
	}
}

func TestAnkoWithCancelation(t *testing.T) {
	expectedExitCode := systools.ExitByStopReason
	expectedMessageCount := 2
	localMsg := []string{}
	errorMsg := ""
	outHandler := func(msg ...interface{}) {
		for _, m := range msg {
			switch s := m.(type) {
			case string:
				localMsg = append(localMsg, s)

			case tasks.MsgExecOutput:
				localMsg = append(localMsg, string(s.Output))
			case tasks.MsgError:
				errorMsg = s.Err.Error()
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Stopreasons: configure.Trigger{
					OnoutContains: []string{"trigger-exec-1"},
				},
				Cmd: []string{
					"println('Hello World')",
					"println('trigger-exec-1')",
					"println('i am done')",
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

	if len(localMsg) != expectedMessageCount {
		t.Error("expected ", expectedMessageCount, " message but got", len(localMsg))
		t.Log(localMsg)
	}
	expectedError := "execution interrupted"
	if errorMsg != expectedError {
		t.Error("expected error", expectedError, "but got", errorMsg)
	}
}

func TestAnkoWithTimeoutCancelation(t *testing.T) {
	expectedExitCode := systools.ExitByTimeout
	expectedMessageCount := 1
	errorMsg := ""
	localMsg := []string{}
	outHandler := func(msg ...interface{}) {
		for _, m := range msg {
			switch s := m.(type) {
			case string:
				localMsg = append(localMsg, s)

			case tasks.MsgExecOutput:
				localMsg = append(localMsg, string(s.Output))
			case tasks.MsgError:
				errorMsg = s.Err.Error()
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Options: configure.Options{

					CmdTimeout: 100,
				},
				Cmd: []string{
					"println('before sleep')",
					"sleep(200)",
					"println('after sleep')",
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

	if len(localMsg) != expectedMessageCount {
		t.Error("expected ", expectedMessageCount, " message but got", len(localMsg))
		t.Log(localMsg)
	}

	expectedError := "execution interrupted"
	if errorMsg != expectedError {
		t.Error("expected error", expectedError, "but got", errorMsg)
	}
}

// testing we report not an timeout error, if the commands stops because of an error
func TestAnkoWithTimeoutAsErrDetect(t *testing.T) {
	expectedExitCode := systools.ExitCmdError
	expectedMessageCount := 0
	localMsg := []string{}
	outHandler := func(msg ...interface{}) {
		for _, m := range msg {
			switch s := m.(type) {
			case string:
				localMsg = append(localMsg, s)

			case tasks.MsgExecOutput:
				localMsg = append(localMsg, string(s.Output))
			case tasks.MsgError:
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Options: configure.Options{

					CmdTimeout: 100,
				},
				Cmd: []string{
					"println('before sleep')",
					"some weird command",
					"sleep(200)",
					"println('after sleep')",
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

	if len(localMsg) != expectedMessageCount {
		t.Error("expected ", expectedMessageCount, " message but got", len(localMsg))
		t.Log(localMsg)
	}
}

// testing we report not an timeout error, if the commands stops because of an error
func TestAnkoExitCmd(t *testing.T) {
	expectedExitCode := systools.ExitCmdError
	expectedMessageCount := 1
	errorMsg := ""
	localMsg := []string{}
	outHandler := func(msg ...interface{}) {
		for _, m := range msg {
			switch s := m.(type) {
			case string:
				localMsg = append(localMsg, s)

			case tasks.MsgExecOutput:
				localMsg = append(localMsg, string(s.Output))
			case tasks.MsgError:
				errorMsg = s.Err.Error()
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					"println('before sleep')",
					"exit()", // this should stop the execution
					"println('after sleep')",
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

	if len(localMsg) != expectedMessageCount {
		t.Error("expected ", expectedMessageCount, " message but got", len(localMsg))
		t.Log(localMsg)
	}
	if errorMsg != "execution interrupted" {
		t.Error("expected error 'execution interrupted' but got", errorMsg)
	}
}

func TestAnkoOsCmd(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
	localMsg := []string{}
	outHandler := func(msg ...interface{}) {
		for _, m := range msg {
			switch s := m.(type) {
			case string:
				localMsg = append(localMsg, s)

			case tasks.MsgExecOutput:
				localMsg = append(localMsg, string(s.Output))
			case tasks.MsgError:
				t.Error(s.Err)
			}
		}
	}
	os := configure.GetOs()
	expectedMessage := "os is " + os

	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					`if ifos("` + os + `") {
						println('os is ` + os + `')
					}
					`,
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

	if len(localMsg) != expectedMessageCount {
		t.Error("expected ", expectedMessageCount, " message but got", len(localMsg))
		t.Log(localMsg)
	} else {
		if localMsg[0] != expectedMessage {
			t.Error("expected message ", expectedMessage, " but got", localMsg[0])
		}
	}
}

func TestAnkoJsonImport(t *testing.T) {
	expectedExitCode := systools.ExitOk
	outHandler := func(msg ...interface{}) {
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					`importJson("test_import_data", '{"master": "data"}')
					`,
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)

	testData, ok := dmc.GetData("test_import_data")
	if !ok {
		t.Error("expected to have data in test_import_data")
	} else {

		expected := make(map[string]interface{})
		expected["master"] = "data"
		if testData["master"] != expected["master"] {
			t.Error("expected data in test_import_data to be", expected, " but got", testData)
		}
	}

	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

}

func TestAnkoJsonImportError(t *testing.T) {
	expectedExitCode := systools.ExitCmdError
	outHandler := func(msg ...interface{}) {
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					`importJson("test_import_data", '{"master": "data"') 
					`,
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

}

func TestAnkoExec01(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
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
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					"msg = exec('echo exec was running')",
					"println(msg)",
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

	if len(errorMsg) > 0 {
		t.Error("expected no error but got", errorMsg)
	}

	if len(localMsg) != expectedMessageCount {
		t.Error("expected ", expectedMessageCount, " message but got", len(localMsg))
		t.Log(localMsg)
	}

}

func TestAnkoVars(t *testing.T) {
	expectedExitCode := systools.ExitOk
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
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					`
					varSet("var1", "misaka")
					importJson("test_import_data", '{"master": "hello ${var1}"}')
					println(varAsJson("test_import_data"))

					`,
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)

	testData, ok := dmc.GetData("test_import_data")
	if !ok {
		t.Error("expected to have data in test_import_data")
	} else {

		expected := make(map[string]interface{})
		expected["master"] = "hello misaka"
		if testData["master"] != expected["master"] {
			t.Error("expected data in test_import_data to be", expected, " but got", testData)
		}
	}

	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

}

func TestAnkoVars02(t *testing.T) {
	expectedExitCode := systools.ExitOk

	expectedDataKeyvalues := map[string]string{
		"test_import_data": `{"master": "hello misaka", "test": "verify hello misaka", "adress": {"street": {"name": "bakerstreet", "number": 33}}}`,
	}
	expectedKeyValues := map[string]string{
		"var1":   "hello misaka",
		"street": "bakerstreet 33",
	}

	expectedOutputs := []string{
		"  adress:bakerstreet 33",
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
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					`
					varSet("var1", "hello")
					varAppend("var1", " misaka")
					importJson("test_import_data", '{"test": "verify ${var1}"}')
					varMapSet("test_import_data", "master", "${var1}")
					varMapSet("test_import_data", "adress.street.name", "bakerstreet")
					varMapSet("test_import_data", "adress.street.number", "33")
					
					varSet("street", "${test_import_data:adress.street.name} ${test_import_data:adress.street.number}")
					
					street = varGet("street")
					println("  adress:" + street)

					`,
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}
	// simple key values
	for key, value := range expectedKeyValues {
		check := dmc.GetPH(key)
		if check != value {
			t.Error("expected", key, "to be:", value, "but got:", check)
		}
	}
	// data containers
	for key, value := range expectedDataKeyvalues {
		data, ok := dmc.GetData(key)
		if !ok {
			t.Error("expected to have data in", key)
		} else {
			for k, v := range data {
				if data, ok := expectedDataKeyvalues[key]; ok {
					if data != value {
						t.Error("expected data in", key, "to be", value, " but got", data)
					}
				} else {
					t.Error("expected key", k, "to be", v, " but got", value)
				}
			}
		}
	}

	// outputs
	for i, expected := range expectedOutputs {
		if localMsg[i] != expected {
			t.Error("expected message", expected, "but got", localMsg[i])
		}
	}

}

func TestAnckoCopy(t *testing.T) {

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
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					`
					copy("testdata/", "temp/")
					`,
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)

	expectedExitCode := systools.ExitOk
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

	if err := assertDirectoryMatch(t, "testdata", "temp"); err != nil {
		t.Error(err)
	}
}

func TestAnckoCopySkip(t *testing.T) {

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
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					`
					copyButSkip("testdata/", "temp/skipcheck", ".sh")
					`,
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)

	expectedExitCode := systools.ExitOk
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

	if exists, err := systools.Exists("temp/skipcheck/case01/.contxt.yml"); err != nil {
		t.Error(err)
	} else if !exists {
		t.Error("expected temp/skipcheck/case01/.contxt.yml to exist")
	}

	if exists, err := systools.Exists("temp/skipcheck/case01/test.sh"); err != nil {
		t.Error(err)
	} else if exists {
		t.Error("expected temp/skipcheck/case01/test.sh not to exists")
	}

}

func TestAnckoCopySingleFile(t *testing.T) {

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
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					`
					err = copy("testdata/case01/test.sh", "temp/singlefilecopy/test.sh")
					`,
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)

	expectedExitCode := systools.ExitOk
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}
	if exists, err := systools.Exists("temp/singlefilecopy/test.sh"); err != nil {
		t.Error(err)
	} else if !exists {
		t.Error("expected temp/singlefilecopy/test.sh to exists")
	}
}

func TestMkdirAndRm(t *testing.T) {

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
				t.Error(s.Err)
			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					`
					mkdir("temp/testout")
					`,
				},
			},
			{
				ID: "rm",
				Cmd: []string{
					`
					remove("temp/testout")
					`,
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)

	expectedExitCode := systools.ExitOk
	code := tsk.RunTarget("test", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

	if exists, err := systools.Exists("temp/testout"); err != nil {
		t.Error(err)
	} else if !exists {
		t.Error("expected temp/testout to exist")
	}
	code = tsk.RunTarget("rm", false)
	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}

	if exists, err := systools.Exists("temp/testout"); err != nil {
		t.Error(err)
	} else if exists {
		t.Error("expected temp/testout not exists")
	}
}

func TestAnkoBase64Encode(t *testing.T) {
	expectedExitCode := systools.ExitOk
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
				t.Error(s.Err)

			case tasks.MsgErrDebug:
				t.Log(s)
				errorMsg = append(errorMsg, s.Err.Error())
				t.Error(s.Err)

			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					`data = base64Encode("hello world")
					varSet("base64", data)
					`,
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)

	testData, ok := dmc.GetPHExists("base64")
	if !ok {
		t.Error("expected to have data in base64")
	} else {

		expected := "aGVsbG8gd29ybGQ="
		if testData != expected {
			t.Error("expected data in test_import_data to be", expected, " but got", testData)
		}
	}

	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}
}

func TestAnkoBase64Decode(t *testing.T) {
	expectedExitCode := systools.ExitOk
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
				t.Error(s.Err)

			case tasks.MsgErrDebug:
				t.Log(s)
				errorMsg = append(errorMsg, s.Err.Error())
				t.Error(s.Err)

			}
		}
	}
	var runCfg configure.RunConfig = configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					`data,_ = base64Decode("aGVsbG8gd29ybGQ=")
					varSet("base64", data)
					`,
				},
			},
		},
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
	tsk.SetHardExistToAllTasks(false)
	code := tsk.RunTarget("test", false)

	testData, ok := dmc.GetPHExists("base64")
	if !ok {
		t.Error("expected to have data in base64")
	} else {

		expected := "hello world"
		if testData != expected {
			t.Error("expected data in test_import_data to be", expected, " but got", testData)
		}
	}

	if code != expectedExitCode {
		t.Error("expected code ", expectedExitCode, " but got", code)
	}
}

func TestStringReplace(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
	expectedMessage := "hello_world"
	cmd := `data = stringReplace("hello world", " ", "_")
	println(data)`

	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{expectedMessage})
}

func TestStringContains(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
	expectedMessage := "true"
	cmd := `data = stringContains("hello world", "world")
	println(data)`

	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{expectedMessage})
}

func TestCopyFile(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 0
	cmd := `
	err = mkdir("temp/copyfile")
	if err != nil {
		println(err)
	}
	err = copyFile("testdata/case01/test.sh", "temp/copyfile/test.sh")
	if err != nil {
		println(err)
	}`

	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{""})
	assertFileMatch(t, "testdata/case01/test.sh", "temp/copyfile/test.sh")
}

func TestCopyFileFail(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
	cmd := `
	err = copyFile("testdata/case01/testNOT_THERE.sh", "temp/copyfile/test.sh")
	if err != nil {
		println(err)
	}`

	AnkoTestRunHelper(
		t, cmd, expectedExitCode, expectedMessageCount,
		[]string{"open testdata/case01/testNOT_THERE.sh: no such file or directory"})
}

func TestCopyInclude(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 0
	cmd := `
	err = copy("testdata/case01", "temp/copyinclude", ".sh")
	if err != nil {
		println(err)
	}`

	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{""})
	assertFileExists(t, "temp/copyinclude/test.sh")
	assertFileNotExists(t, "temp/copyinclude/.contxt.yml")
}

// import a json file into the data store. e.g. importJsonFile('key','path/to/file.json')
func TestImportJsonFile(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 0
	cmd := `
	err =importJsonFile("imported", "testdata/data/file01.json")
	if err != nil {
		println(err)
	}`

	dmc, _ := AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{""})
	if data, ok := dmc.GetData("imported"); ok {
		if data["testString"] != "testString" {
			t.Error("expected name to be 'testString' but got", data["testString"])
		}
	} else {
		t.Error("expected to have data in imported")
	}

}

func TestImportJsonFileFail(t *testing.T) {
	expectedExitCode := systools.ExitCmdError
	expectedMessageCount := 2
	cmd := `
	err =importJsonFile("imported", "testdata/data/file01_NOT_THERE.json")
	if err != nil {
		println(err)
	}`

	AnkoTestRunHelperWithErrors(
		t, cmd, true, expectedExitCode, expectedMessageCount,
		[]string{
			"open testdata/data/file01_NOT_THERE.json: no such file or directory",
			"Error in script: open testdata/data/file01_NOT_THERE.json: no such file or directory errType: *fs.PathError ",
		})
}

func TestImportJsonFileFail02(t *testing.T) {
	expectedExitCode := systools.ExitCmdError
	expectedMessageCount := 2
	cmd := `
	err =importJsonFile("imported", "testdata/data/file01.yaml")
	if err != nil {
		println(err)
	}`

	AnkoTestRunHelperWithErrors(
		t, cmd, true, expectedExitCode, expectedMessageCount,
		[]string{
			"invalid character 'e' in literal true (expecting 'r')",
			"Error in script: invalid character 'e' in literal true (expecting 'r') errType: *json.SyntaxError ",
		})
}

func TestImportYamlFile(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 0
	cmd := `
	err =importYamlFile("imported", "testdata/data/file01.yaml")
	if err != nil {
		println(err)
	}`

	dmc, _ := AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{""})
	if data, ok := dmc.GetData("imported"); ok {
		if subdata, ok := data["stringMap"].(map[string]interface{}); ok {
			if subdata2, ok := subdata["test01"].(string); ok {
				if subdata2 != "value01" {
					t.Error("expected name to be 'testString' but got", subdata2)
				}
			} else {
				t.Error("expected to have data in imported as imported named test01")
			}
		} else {
			t.Error("expected to have data in imported as imported named stringMap")
		}

	} else {
		t.Error("expected to have data in imported as imported named data")
	}

}

func TestImportYamlFail01(t *testing.T) {
	expectedExitCode := systools.ExitCmdError
	expectedMessageCount := 2
	cmd := `
	err =importYamlFile("imported", "testdata/data/file01_NOT_THERE.yaml")
	if err != nil {
		println(err)
	}`

	AnkoTestRunHelperWithErrors(
		t, cmd, true, expectedExitCode, expectedMessageCount,
		[]string{
			"open testdata/data/file01_NOT_THERE.yaml: no such file or directory",
			"Error in script: open testdata/data/file01_NOT_THERE.yaml: no such file or directory errType: *fs.PathError ",
		})
}

func TestImportYamlFail02(t *testing.T) {
	expectedExitCode := systools.ExitCmdError
	expectedMessageCount := 3
	cmd := `
	err =importYamlFile("imported", "testdata/data/file01.txt")
	if err != nil {
		println(err)
	}`

	AnkoTestRunHelperWithErrors(
		t, cmd, true, expectedExitCode, expectedMessageCount,
		[]string{
			"yaml: unmarshal errors:",
			"  line 1: cannot unmarshal !!str `this fi...` into []interface {}",
			"Error in script: yaml: unmarshal errors:  line 1: cannot unmarshal !!str `this fi...` into []interface {} errType: *yaml.TypeError",
		})
}

func TestGetOs(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
	expectedMessage := "os is " + configure.GetOs()
	cmd := `os = getos()
	println("os is " + os)`
	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{expectedMessage})
}

func TestVarAsYaml(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 7
	expectedMessage := `adress:
	    street:
	      name: bakerstreet
          number: "33"
		  master: hello misaka
		  test: verify hello misaka
		  !IGNORE
		  `
	cmd := `
varSet("var1", "hello")
varAppend("var1", " misaka")
importJson("test_import_data", '{"test": "verify ${var1}"}')
varMapSet("test_import_data", "master", "${var1}")
varMapSet("test_import_data", "adress.street.name", "bakerstreet")
varMapSet("test_import_data", "adress.street.number", "33")
println(varAsYaml("test_import_data"))
`
	expectedSlicedMessage := strings.Split(expectedMessage, "\n")
	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, expectedSlicedMessage)
}

func TestFromJson(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
	expectedMessage := "hello world"
	cmd := `
	  json = '{"test": "hello world"}'
	  data,err = fromJson(json)
	  if err != nil {
	  	println(err)
      }
	  println(data["test"])`
	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{expectedMessage})
}

func TestFromJsonError(t *testing.T) {
	expectedExitCode := systools.ExitCmdError
	expectedMessageCount := 1
	cmd := `
	  json = '{"test": "hello world"'
	  data,err = fromJson(json)
	  if err != nil {
	  	println(err)
	  }
	  println(data["test"])`
	AnkoTestRunHelperWithErrors(t, cmd, true, expectedExitCode, expectedMessageCount, []string{"unexpected end of JSON input"})
}

func TestStringSplit(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
	expectedMessage := "hello"
	cmd := `
	  data = stringSplit("hello world", " ")
	  println(data[0])`
	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{expectedMessage})
}

func TestVarParse(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
	expectedMessage := "hello misaka"
	cmd := `
	  varSet("var1", "hello")
	  varAppend("var1", " misaka")
	  println(varParse("${var1}"))`
	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{expectedMessage})
}

func TestVarExists(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
	expectedMessage := "true"
	cmd := `
	  varSet("var1", "hello")
	  println(varExists("var1"))`
	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{expectedMessage})
}

func TestVarExistsFalse(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
	expectedMessage := "false"
	cmd := `
	  println(varExists("var1"))`
	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{expectedMessage})
}

func TestVarMapSetError(t *testing.T) {
	expectedExitCode := systools.ExitCmdError
	expectedMessageCount := 1
	cmd := `
	  varMapSet("test_import_data", "master", "${var1}")
	  `
	AnkoTestRunHelperWithErrors(t, cmd, true, expectedExitCode, expectedMessageCount,
		[]string{"Error in script: the key [test_import_data] does not exists errType: *errors.errorString"})
}

func TestVarMapToJson(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
	expectedMessage := `{"master":"hello misaka"}`
	cmd := `
	  json = '{"master": "hello misaka"}'
	  importJson("test_import_data", json)
	  result,err = varMapToJson("test_import_data")
	  println(result)`
	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{expectedMessage})
}

func TestVarMapToJsonError(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
	cmd := `
	  data, err = varMapToJson("test_import_data")
	  if err != nil {
	  	println(err)
	  }`
	AnkoTestRunHelperWithErrors(t, cmd, true, expectedExitCode, expectedMessageCount,
		[]string{"map named test_import_data not found"})
}

func TestVarMapToYaml(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 2 // newline in the yaml
	expectedMessage := `master: hello misaka
`
	cmd := `
	  json = '{"master": "hello misaka"}'
	  importJson("test_import_data", json)
	  result,err = varMapToYaml("test_import_data")
	  println(result)`
	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{expectedMessage})
}

func TestVarMapToYamlError(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
	cmd := `
	  data, err = varMapToYaml("test_import_data")
	  if err != nil {
	  	println(err)
	  }`
	AnkoTestRunHelperWithErrors(t, cmd, true, expectedExitCode, expectedMessageCount,
		[]string{"map named test_import_data not found"})
}

func TestVarWrite(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 0
	cmd := `	  
	  varSet("test_import_data", "hello world")
	  err = varWrite("test_import_data", "temp/test_import_data.temp")
	  if err != nil {
	  	println(err)
	  }`
	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{""})
	assertFileExists(t, "temp/test_import_data.temp")
	assertFileMatchAndRemoveOrig(t, "temp/test_import_data.temp", "expected/test_import_data.verify", true)
}

func TestWriteFile(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 0
	cmd := `	  
	  err = writeFile("temp/test_varWrite.temp", "hello world")
	  if err != nil {
	  	println(err)
	  }`
	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{""})
	assertFileExists(t, "temp/test_varWrite.temp")
	assertFileMatchAndRemoveOrig(t, "temp/test_varWrite.temp", "expected/test_import_data.verify", true)
}

func TestWriteFileAndCheckReplacedVars(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 0
	cmd := `	  
	  varSet("test_import_data", "hello world")
	  err = writeFile("temp/test_varWrite.temp", "hello ${test_import_data}")
	  if err != nil {
	  	println(err)
	  }`
	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, []string{""})
	assertFileExists(t, "temp/test_varWrite.temp")
	assertFileMatchAndRemoveOrig(t, "temp/test_varWrite.temp", "expected/test_import_data.verify", true)
}

func TestReadFile(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 4
	expectedMessages := []string{
		"line1",
		"line2 two",
		"line3 three",
		"line4 four",
	}

	cmd := `	  
	  data,err = readFile("testdata/data/file02.txt")
	  if err != nil {
	  	println(err)
	  }
	  println(data)`
	AnkoTestRunHelper(t, cmd, expectedExitCode, expectedMessageCount, expectedMessages)
}

func TestBase64DecoeError(t *testing.T) {
	expectedExitCode := systools.ExitCmdError
	expectedMessageCount := 2
	cmd := `
	  data,err = base64Decode("aGVsbG8gd29ybGQ")
	  if err != nil {
	  	println(err)
	  }`
	AnkoTestRunHelperWithErrors(t, cmd, true, expectedExitCode, expectedMessageCount,
		[]string{"illegal base64 data at input byte 12", "Error in script: illegal base64 data at input byte 12 errType: base64.CorruptInputError "})
}

func TestNeedsExecutedOnce(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := -1
	expectedMessage := []string{"needcheck_2 true", "needcheck_1 true", "true"}
	runConfig := configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "neededOne",
				Cmd: []string{
					`varSet("var1", "hello")`,
					`println(varExists("var1"))`,
				},
			},
			{
				ID:    "taskInBetween",
				Needs: []string{"neededOne"},
				Cmd: []string{
					`println("needcheck_1",varExists("var1"))`,
				},
			},
			{
				ID: "neededTwo",
				Options: configure.Options{
					Displaycmd: true,
				},
				Needs: []string{"neededOne", "taskInBetween"},
				Cmd: []string{
					`println("needcheck_2", varExists("var1"))`,
				},
			},
		},
	}
	dh, _, logger := RunTargetHelperWithErrors(t, "neededTwo", runConfig, false, expectedExitCode, expectedMessageCount, expectedMessage, true)
	if len(logger.errors) > 0 {
		t.Error("expected no errors but got", len(logger.errors))
	}
	// the need is defined twice, but we have to make sure it is running only once
	if !assertSliceContainsAtLeast(t, logger.debugs, "need already handled neededOne", 1) {
		t.Error("expected messages \"need already handled neededOne\" but got", logger.debugs)
	}

	var1 := dh.GetPH("var1")
	if var1 != "hello" {
		t.Error("expected var1 to be hello but got[", var1, "]")
	}
}

func TestNeedsExecutedOnceWithoutCode(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := -1
	expectedMessage := []string{"needcheck_2 true", "needcheck_1 true", "true"}
	runConfig := configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "neededOne",
				Variables: map[string]string{
					"var1":   "hello",
					"checkA": "needcheck_1",
					"run_1":  "true",
				},
			},
			{
				ID:    "taskInBetween",
				Needs: []string{"neededOne"},
				Variables: map[string]string{
					"checkA": "needcheck_1",
					"run_2":  "true",
				},
			},
			{
				ID:    "neededTwo",
				Needs: []string{"neededOne", "taskInBetween"},
				Variables: map[string]string{
					"checkA": "needcheck_2",
					"run_3":  "true",
				},
			},
		},
	}
	dh, _, logger := RunTargetHelperWithErrors(t, "neededTwo", runConfig, false, expectedExitCode, expectedMessageCount, expectedMessage, true)
	if len(logger.errors) > 0 {
		t.Error("expected no errors but got", len(logger.errors))
	}
	// like in the test above, the need is defined twice, but we have to make sure it is running only once
	// but different, there is no code in the tasks, so we have to check the tasks get executed anyway
	if !assertSliceContainsAtLeast(t, logger.debugs, "need already handled neededOne", 1) {
		t.Error("expected messages \"need already handled neededOne\" but got", logger.debugs)
	}

	var1 := dh.GetPH("var1")
	if var1 != "hello" {
		t.Error("expected var1 to be hello but got[", var1, "]")
		t.Log(logger.debugs)
	}
	keys := dh.GetDataKeys()
	if !assertSliceContainsAtLeast(t, keys, "run_1", 1) {
		t.Error("expected to have run_1 in the data keys but got", keys)
	}
	if !assertSliceContainsAtLeast(t, keys, "run_2", 1) {
		t.Error("expected to have run_2 in the data keys but got", keys)
	}
	if !assertSliceContainsAtLeast(t, keys, "run_3", 1) {
		t.Error("expected to have run_3 in the data keys but got", keys)
	}
}

func TestExecNoTasks(t *testing.T) {
	expectedExitCode := systools.ExitByNoTargetExists
	expectedMessageCount := 0
	runConfig := configure.RunConfig{}
	_, _, logger := RunTargetHelperWithErrors(t, "neededTwo", runConfig, false, expectedExitCode, expectedMessageCount, []string{}, true)
	if len(logger.errors) > 0 {
		t.Error("expected no errors but got", len(logger.errors))
	}
}

func TestAnkoWaitMillis(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 1
	cmd := `
	  waitMillis(100)
	  println("done")`

	runConfig := configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test",
				Cmd: []string{
					cmd,
				},
			},
		},
	}
	startTime := time.Now()
	_, _, logger := RunTargetHelperWithErrors(t, "test", runConfig, false, expectedExitCode, expectedMessageCount, []string{"done"}, true)
	execTimeInMillis := time.Since(startTime).Milliseconds()

	if execTimeInMillis < 100 {
		t.Error("expected to wait at least 100ms but got", execTimeInMillis)
	}
	if len(logger.errors) > 0 {
		t.Error("expected no errors but got", len(logger.errors))
	}

}

func TestInvalidTargetName(t *testing.T) {
	expectedExitCode := systools.ErrorInvalidTargetName
	expectedMessageCount := 0
	cmd := `
	  println("done")`

	runConfig := configure.RunConfig{
		Task: []configure.Task{
			{
				ID: "test#$",
				Cmd: []string{
					cmd,
				},
			},
		},
	}
	_, _, logger := RunTargetHelperWithErrors(t, "test#$", runConfig, false, expectedExitCode, expectedMessageCount, []string{}, true)
	if len(logger.errors) < 1 {
		t.Error("expected at least one error")
	} else {
		if !strings.Contains(logger.errors[0], "invalid target name") {
			t.Error("expected error to contain 'invalid target name' but got", logger.errors[0])
		}
	}
}

func TestWorkingDirSplit(t *testing.T) {
	expectedExitCode := systools.ExitOk
	expectedMessageCount := 3
	expectedMessages := []string{"task_diffdir", "task_samedir", "task_main"}
	runConfig := configure.RunConfig{
		Config: configure.Config{
			Sequencially: false,
		},

		Task: []configure.Task{
			{
				ID:    "maintask",
				Needs: []string{"samedir", "diffdir"},
				Script: []string{
					"echo task_main",
				},
			},

			{
				ID: "samedir",
				Script: []string{
					"echo task_samedir",
				},
			},
			{
				ID: "diffdir",
				Script: []string{
					"echo task_diffdir",
				},
				Options: configure.Options{
					WorkingDir: "testdata",
				},
			},
		},
	}
	_, _, logger := RunTargetHelperWithErrors(
		t, "maintask", runConfig, false, expectedExitCode, expectedMessageCount, expectedMessages, true,
	)
	if len(logger.errors) > 0 {
		t.Error("expected no errors but got", len(logger.errors))
	}
}
