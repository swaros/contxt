package tasks_test

import (
	"strings"
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/mimiclog"
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
