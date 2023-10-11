package tasks_test

import (
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
