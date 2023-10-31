package tasks_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/swaros/contxt/module/tasks"
)

func TestTask(t *testing.T) {
	pids := []int{}
	tasks.HandleAllMyPid(func(pid int) error {
		pids = append(pids, pid)
		return nil
	})

	if len(pids) == 0 {
		t.Error("no pids found")
	}
}

func TestOwnProcInfoWindowsOnly(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.SkipNow()
		return
	}
	pid := os.Getpid()
	proc, err := tasks.WinProcInfo(pid)
	if err != nil {
		t.Error(err)
	}
	if proc.ProcName == "" {
		t.Error("no process info found")
	} else {
		t.Logf("Process info: %v", proc)
	}
}
