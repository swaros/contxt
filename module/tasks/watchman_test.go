package tasks_test

import (
	"testing"
	"time"

	"github.com/swaros/contxt/module/tasks"
)

func TestWatchmanInc(t *testing.T) {
	wm := tasks.NewWatchman()
	runCnt := 1000
	helpsRunAsync(runCnt, []string{"task1", "task1", "task2", "task1", "task2"}, func(name string, cnt int) bool {
		wm.IncTaskCount(name)
		return true
	})

	expectedCnt := runCnt * 3
	if wm.GetTaskCount("task1") != expectedCnt {
		t.Errorf("expected task1 to have run at least %d times, but it ran %d times", expectedCnt, wm.GetTaskCount("task1"))
	}

	expectedCnt = runCnt * 2
	if wm.GetTaskCount("task2") != expectedCnt {
		t.Errorf("expected task2 to have run at least %d times, but it ran %d times", expectedCnt, wm.GetTaskCount("task2"))
	}
}

func TestWatchmanDoneCheck(t *testing.T) {
	wm := tasks.NewWatchman()
	runCnt := 100
	overAllOk := helpsRunAsync(runCnt, []string{"task1", "task1", "task2", "task1", "task2"}, func(name string, cnt int) bool {
		wm.IncTaskCount(name)
		time.Sleep(5 * time.Millisecond)
		wm.IncTaskDoneCount(name)
		return wm.GetTaskDone(name)
	})

	if !overAllOk {
		t.Errorf("expected all tasks to be done, but they are not")
	}

	if !wm.GetTaskDone("task1") {
		t.Errorf("expected task1 to be done, but it is not")
	}

	if !wm.GetTaskDone("task2") {
		t.Errorf("expected task2 to be done, but it is not")
	}

}
