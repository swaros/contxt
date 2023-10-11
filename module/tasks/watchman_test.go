package tasks_test

import (
	"os"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/swaros/contxt/module/systools"
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

func helperForFindTask(t *testing.T, cmdlike string) []string {
	t.Helper()
	// we use linux special features here
	if runtime.GOOS != "linux" {
		t.Skip("not supported on anything else than linux")
	}
	runner := tasks.GetShellRunner()
	command := "ps -ef"
	result := []string{}
	runner.Exec(
		command,
		func(msg string, err error) bool {
			if strings.Contains(msg, cmdlike) {
				result = append(result, msg)
			}
			return true
		},
		func(proc *os.Process) {
		},
	)
	return result
}

func helperKillTask(t *testing.T, pid string) []string {
	t.Helper()
	// we use linux special features here
	if runtime.GOOS != "linux" {
		t.Skip("not supported on anything else than linux")
	}
	runner := tasks.GetShellRunner()
	command := "kill -9 " + pid
	result := []string{"command [" + command + "]"}
	runner.Exec(
		command,
		func(msg string, err error) bool {
			result = append(result, msg)
			return true
		},
		func(proc *os.Process) {
		},
	)
	return result
}

func helperLaunchShellCmdInBackround(t *testing.T, command, target string, wman *tasks.Watchman, forceStop bool) {
	t.Helper()
	runner := tasks.GetShellRunner()
	// make sure we create a new task
	wman.IncTaskCount(target)

	go runner.Exec(
		command,
		func(msg string, err error) bool {
			if forceStop {
				return false
			}
			return err == nil
		},
		func(proc *os.Process) {
			if wtask, found := wman.GetTask(target); found {
				wtask.StartTrackProcess(proc)
				wtask.LogCmd(runner.GetCmd(), runner.GetArgs(), command)
				if err := wman.UpdateTask(target, wtask); err != nil {
					t.Errorf("expected no error, but got %v", err)
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
}

// testing the process tracking by an background task, and stop it
// by sending a SIGTERM and SIGKILL to the process.
// here we do ot use the watchman to stop the task
// we use the os directly, so none of watchman kill methods are used.
// this is done to test the process tracking
// kill methods of the watchman are tested in TestTaskStoppingByWatchman
func TestTaskTracking(t *testing.T) {
	// we use linux special features here
	if runtime.GOOS != "linux" {
		t.Skip("not supported on anything else than linux")
	}

	// start a task in the background and let it run for 1 second
	// and try to stop it
	// this is done by sending a SIGTERM to the process
	// and then checking if the process is still running
	// if it is still running, we send a SIGKILL to the process
	// and check again if the process is still running
	// if it is still running, we fail the test

	// this is the command we want to run
	command := "sleep 1"
	// this is the target we want to use
	target := "taskKillSleeper"
	// this is the watchman we want to use
	wman := tasks.NewWatchman()
	// we use the helper to launch the command in the background
	helperLaunchShellCmdInBackround(t, command, target, wman, false)

	// wait till the process is started
	if success, timeUsed := wman.WaitForProcessStart(target, 10*time.Millisecond, 10); !success {
		t.Error("failed to start process in time", timeUsed)
	}
	// get the task
	if wtask, found := wman.GetTask(target); found {
		if !wtask.IsProcessRunning() {
			t.Error("expected process to be running, but it is not")
		}
		// get the process
		proc := wtask.GetProcess()
		if proc == nil {
			t.Error("expected process not to be nil")
		}
		// get the processInfo
		// get the pid
		pid, _ := wtask.GetProcessPid()
		if pid == 0 {
			t.Error("expected pid not to be 0")
		}
		// send the SIGTERM
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			t.Errorf("expected no error, but got %v", err)
		}
		// wait a bit, so the process can stop
		time.Sleep(100 * time.Millisecond)
		// check if the process is still running
		if wtask.IsProcessRunning() {
			// send the SIGKILL
			if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
				t.Errorf("expected no error, but got %v", err)
			}
			// wait a bit, so the process can stop
			time.Sleep(100 * time.Millisecond)
			// check if the process is still running
			if wtask.IsProcessRunning() {
				t.Error("expected process not to be running, but it is")
			}
		}
	} else {
		t.Errorf("expected task to be found, but it is not")
	}

}

func TestTaskStoppingByWatchman(t *testing.T) {
	// we use linux special features here
	if runtime.GOOS != "linux" {
		t.Skip("not supported on anything else than linux")
	}
	// see TestTaskTracking for more details about the propsed workflow
	command := "sleep 10"
	target := "watchmanKillTask"
	wman := tasks.NewWatchman()
	helperLaunchShellCmdInBackround(t, command, target, wman, false)

	if success, timeUsed := wman.WaitForProcessStart(target, 10*time.Millisecond, 10); !success {
		t.Error("failed to start process in time", timeUsed)
	}
	if wtask, found := wman.GetTask(target); found {
		err := wtask.KillProcess()
		if err != nil {
			t.Errorf("expected no error, but got [%v]", err)
		}
		time.Sleep(100 * time.Millisecond)
		if wtask.IsProcessRunning() {
			t.Error("expected process not to be running, but it is")
		}
	} else {
		t.Errorf("expected task to be found, but it is not")
	}

}

// testing if we run in the timeout, because the task is not started
// so we expect the WaitForProcessStart to run into the timeout
func TestWMWaitForProceesStart(t *testing.T) {
	wman := tasks.NewWatchman()
	target := "neverStatedTask"
	if success, timeUsed := wman.WaitForProcessStart(target, 10*time.Millisecond, 10); !success {
		assertIntEqual(t, 100000000, timeUsed)
	} else {
		t.Error("expected to fail, but it did not")
	}
}

func TestMultipleTaskManagement(t *testing.T) {
	command := "sleep 5"

	wman := tasks.NewWatchman()

	targets := map[string]string{
		"multiTask1": command,
		"multiTask2": command,
		"multiTask3": command,
	}

	for target, command := range targets {
		helperLaunchShellCmdInBackround(t, command, target, wman, false)
	}

	for target := range targets {
		if success, timeUsed := wman.WaitForProcessStart(target, 5*time.Millisecond, 10); !success {
			t.Error("failed to start process in time", timeUsed)
		}
	}
	tasks.ShutDownProcesses()
	for target := range targets {
		_, found := wman.GetTask(target)
		if !found {
			t.Errorf("expected task %q to be found, but it is not", target)
		}

	}

}

func TestMultipleTaskManagementWithChildProcs(t *testing.T) {
	command := "sleep 5"

	wman := tasks.NewWatchman()

	targets := map[string]string{
		"multiTask1": command,
		"multiTask2": "watch date > dateout.bin.tmp",
		"multiTask3": command,
	}

	for target, command := range targets {
		helperLaunchShellCmdInBackround(t, command, target, wman, false)
	}

	for target := range targets {
		if success, timeUsed := wman.WaitForProcessStart(target, 5*time.Millisecond, 10); !success {
			t.Error("failed to start process in time", timeUsed)
		}
	}
	wman.StopAllTasks(func(target string, time int, succeed bool) {
		t.Log("task", target, "stopped in", time, "ms")
		if !succeed {
			t.Error("task", target, "did not succeed by stopping his process")
		}
	})
	for target := range targets {
		_, found := wman.GetTask(target)
		if !found {
			t.Errorf("expected task %q to be found, but it is not", target)
		}

	}
	tasks := helperForFindTask(t, "watch date")
	for _, task := range tasks {

		t.Error("expected 'watch date' to be stopped, but it is not:\n", task)
		task = systools.TrimAllSpaces(task)
		t.Log(task)
		parts := strings.Split(task, " ")
		if len(parts) > 1 {
			pid := parts[1]
			t.Log("killing task", pid)
			killInfo := helperKillTask(t, pid)
			for _, killMsg := range killInfo {
				t.Log(killMsg)
			}
		}

	}

}
