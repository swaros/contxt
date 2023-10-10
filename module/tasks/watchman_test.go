package tasks_test

import (
	"os"
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

// testing the basic workflow of tracking a process
func TestWatchManTaskInfo(t *testing.T) {
	runner := tasks.GetShellRunner()
	command := "echo 'hello watchman'"

	target := "task1"
	wman := tasks.NewWatchman()
	wman.IncTaskCount(target)
	trackPid := 0

	// internalExitCode, processExitCode, err
	intCode, realCode, err := runner.Exec(
		command,
		func(msg string, err error) bool {
			if err != nil {
				t.Errorf("expected no error, but got %v", err)
			}
			t.Log("output:", msg)
			return true
		},
		func(proc *os.Process) {
			trackPid = proc.Pid
			if wtask, found := wman.GetTask(target); found {
				wtask.StartTrackProcess(proc)
				wtask.LogCmd(runner.GetCmd(), runner.GetArgs(), command)
				if err := wman.UpdateTask(target, wtask); err != nil {
					t.Errorf("expected no error, but got %v", err)
					// if this is not working, anything else will also fail for sure.
					// no need to spam the test with more errors
					t.SkipNow()
				}

				// testing if the task is still running
				// this is done by checking the processInfo of the task.
				// it depends on the os, if the processInfo is still valid
				if !wtask.IsProcessRunning() {
					t.Error("expected process to be running, but it is not")
				}
			}
		},
	)
	assertNoError(t, err)
	assertIntEqual(t, 0, intCode)
	assertIntEqual(t, 0, realCode)

	if wtask, found := wman.GetTask(target); found {
		checkTheTask := wtask.GetProcess()
		if checkTheTask == nil {
			t.Errorf("expected task to have process, but it does not")
		}

		if wtask.IsProcessRunning() {
			t.Errorf("expected is no longer running, but it is")
		}
		if wtask.GetProcess() == nil {
			t.Errorf("expected process not to be nil")
		}
		if len(wtask.GetProcessLog()) != 1 {
			t.Errorf("expected processLog to have 1 entry, but it has %d", len(wtask.GetProcessLog()))
		} else {
			// doing this in else, because we need the first entry
			firstLog := wtask.GetProcessLog()[0]
			if firstLog.Cmd != runner.GetCmd() {
				t.Errorf("expected processLog to have cmd %q, but it has %q", runner.GetCmd(), firstLog.Cmd)
			}
			if len(firstLog.Args) != len(runner.GetArgs()) {
				t.Errorf("expected processLog to have args %q, but it has %q", runner.GetArgs(), firstLog.Args)
			}
			if firstLog.Command != command {
				t.Errorf("expected processLog to have command %q, but it has %q", command, firstLog.Command)
			}
			if firstLog.Pid != trackPid {
				t.Errorf("expected processLog to have pid %d, but it has %d", trackPid, firstLog.Pid)
			}
		}
	} else {
		t.Errorf("expected task to be found, but it is not")
	}

}
