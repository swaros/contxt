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
}

func TestAnkoWithTimeoutCancelation(t *testing.T) {
	expectedExitCode := systools.ExitByTimeout
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
}
