package tasks_test

import (
	"strings"
	"testing"

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
			t.Log(localMsg)
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
