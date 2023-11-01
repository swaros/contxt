package process_test

import (
	"os"
	"testing"

	"github.com/swaros/contxt/module/process"
)

func TestBasicRun(t *testing.T) {
	process := process.NewProcess("bash", "-c", "echo 'Hello World'")
	if err := process.Run(); err != nil {
		t.Error(err)
	}

}

func TestRunWithArgs(t *testing.T) {
	process := process.NewProcess("bash")
	process.SetRunArgs("echo 'Hello World'", "echo 'Hello World 2'")
	if err := process.Run(); err != nil {
		t.Error(err)
	}

}

func TestExecWithBash(t *testing.T) {
	process := process.NewProcess("bash")

	process.SetRunArgs("echo 'Hello World'", "echo 'Hello World 2'")

	process.SetOnOutput(func(msg string, err error) bool {
		t.Log("output[", msg, "]")
		return true
	})

	process.SetOnInit(func(proc *os.Process) {
		if proc == nil {
			t.Error("Process is nil")
		} else {
			t.Logf("Process started with pid %d", proc.Pid)
		}
	})

	realCode, internCode, err := process.Exec()
	if err != nil {
		t.Error(err)
	}
	if realCode != 0 {
		t.Error("realCode is not 0. It is ", realCode)
	}
	if internCode != 0 {
		t.Error("internCode is not 0, It is ", internCode)
	}

}

func TestExecWithBashAndStayOpen(t *testing.T) {
	process := process.NewProcess("bash")

	//process.SetStayOpen(true)

	process.SetOnOutput(func(msg string, err error) bool {
		t.Log("output[", msg, "]")
		return true
	})

	process.SetOnInit(func(proc *os.Process) {
		if proc == nil {
			t.Error("Process is nil")
		} else {
			t.Logf("Process started with pid %d", proc.Pid)
		}
	})

	process.Command("echo 'Hello World'")

	realCode, internCode, err := process.Exec()
	if err != nil {
		t.Error(err)
	}
	if realCode != 0 {
		t.Error("realCode is not 0. It is ", realCode)
	}
	if internCode != 0 {
		t.Error("internCode is not 0, It is ", internCode)
	}

}
