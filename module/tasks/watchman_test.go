package tasks_test

import (
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
)

func assertStringEqual(t *testing.T, expected, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected\n%s\nbut got\n%s", expected, actual)
	}
}

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

func TestIncAndDec(t *testing.T) {
	wm := tasks.NewWatchman()
	if ok, err := wm.IncTaskDoneCount("task1"); ok {
		t.Error("expected to fail, but it did not")
	} else {
		assertStringEqual(t, "can not increase done count for task \"task1\", because it does not exists", err.Error())
	}

	cnt := wm.IncTaskCount("test1")
	if cnt != 1 {
		t.Errorf("expected to get 1, but got %d", cnt)
	}
	cnt = wm.IncTaskCount("test1")
	if cnt != 2 {
		t.Errorf("expected to get 2, but got %d", cnt)
	}
	if ok, err := wm.IncTaskDoneCount("test1"); !ok {
		t.Error("expected to succeed, but it did not")
		if err != nil {
			t.Errorf("expected no error, but got %v", err)
		}
	}

	if !wm.TaskRunning("test1") {
		t.Error("expected task to be running, but it is not")
	}

	cnt = wm.IncTaskCount("test1")
	if cnt != 3 {
		t.Errorf("expected to get 3, but got %d", cnt)
	}

	for i := 0; i < 2; i++ {
		if ok, err := wm.IncTaskDoneCount("test1"); !ok {
			t.Error("expected to succeed, but it did not")
			if err != nil {
				t.Errorf("expected no error, but got %v", err)
			}
		}
	}
	if wm.TaskRunning("test1") {
		t.Error("expected task not to be running, but it is")
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

// Linux only
// running a simple ps -ef and check if the command is found
// returns any finding as a slice of strings.
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

// Linux only
// running a simple kill -9 <pid> to stop a process
// returns any output as a slice of strings.
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

// testing the process tracking by an background task.
// this function is just start any command in the default shell in the background
func helperLaunchShellCmdInBackround(t *testing.T, command, target string, wman *tasks.Watchman, forceStop bool, waitForPid bool) int {
	t.Helper()
	runner := tasks.GetShellRunner()
	// make sure we create a new task
	wman.IncTaskCount(target)
	pid := 0
	procRecived := false
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
				pid = proc.Pid
				procRecived = true
			}
		},
	)
	if waitForPid {
		maxTicks := 100
		tick := 0
		for !procRecived {
			time.Sleep(10 * time.Millisecond)
			tick++
			if tick > maxTicks {
				t.Error("expected to get process, but it did not happen")
				break
			}
		}
	}
	return pid
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
	helperLaunchShellCmdInBackround(t, command, target, wman, false, false)

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
		helperLaunchShellCmdInBackround(t, command, target, wman, false, false)
	}

	for target := range targets {
		if success, timeUsed := wman.WaitForProcessStart(target, 5*time.Millisecond, 10); !success {
			t.Error("failed to start process in time", timeUsed)
		}
	}
	tasks.ShutDownProcesses(nil)
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
		helperLaunchShellCmdInBackround(t, command, target, wman, false, false)
	}

	for target := range targets {
		if success, timeUsed := wman.WaitForProcessStart(target, 5*time.Millisecond, 10); !success {
			t.Error("failed to start process in time", timeUsed)
		}
	}
	wman.StopAllTasks(func(target string, time int, succeed bool) {
		//t.Log("task", target, "stopped in", time, "ms")
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

func TestVerify(t *testing.T) {
	wm := tasks.NewWatchman()

	_, ok := tasks.GetNameOfWatchman(wm)
	if !ok {
		t.Error("expected to get the name of the watchman, but it did not")
	}

	// nothing runs any task1 yet. so they shold be reported as not found
	if !wm.TryCreate("task1") {

		t.Error("there should be no task1, but it is")

	}

	// the check itself should also change nothing dependig the tasks
	if wm.TryCreate("task1") {

		t.Error("there should be a task1, but it is not")
	}

}

// testing VerifyTaskExists with a raise condition.
// see #178
// this test is right now there just to hit the race condition
// on 100 runs we need luck to hit it.
// on 1000 runs we hit it for sure.
// --------------------------
// to fix the raise condition, we will simplify the code
// and remove the VerifyTaskExists method.
// instead we use TryCreate to get the task created and having the
// calback afterwards just to report if the task was created or not.
// at this point, the lock is released and we can do what ever we want.
func TestWatchmanVerify(t *testing.T) {
	wm := tasks.NewWatchman()
	runCnt := 1000
	timesCreated := 0
	helpsRunAsync(runCnt, []string{"task_1", "task_2", "task_1", "task_2", "task_1"}, func(name string, cnt int) bool {
		if wm.TryCreate(name) {
			timesCreated++
		}

		return true
	})

	if timesCreated != 2 {
		t.Errorf("expected to create 2 tasks, but created %d", timesCreated)
	}
}
