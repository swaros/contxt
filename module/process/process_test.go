package process_test

import (
	"os"
	"testing"
	"time"

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
	proc := process.NewProcess("bash")

	proc.SetStayOpen(true)

	proc.SetOnOutput(func(msg string, err error) bool {
		t.Log("output[", msg, "]")
		return true
	})

	proc.SetOnInit(func(proc *os.Process) {
		if proc == nil {
			t.Error("Process is nil")
		} else {
			t.Logf("Process started with pid %d", proc.Pid)
		}
	})

	realCode, internCode, err := proc.Exec()
	if err != nil {
		t.Error(err)
	}
	if realCode != 0 {
		t.Error("realCode is not 0. It is ", realCode)
	}
	if internCode != process.ExitInBackGround {
		t.Error("internCode is not 0, It is ", internCode)
	}
	proc.Command("echo 'Hello World'")
	proc.Command("echo 'test 2'")
	// give the process some time to execute the command
	time.Sleep(100 * time.Millisecond)
	proc.Stop()

}
