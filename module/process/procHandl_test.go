package process_test

import (
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/swaros/contxt/module/process"
	"github.com/swaros/contxt/module/systools"
)

// using the current process to test the ReadProc function
func TestReadProc(t *testing.T) {
	proc, err := process.ReadProc(os.Getpid())
	if err != nil {
		t.Error(err)
	} else {
		if proc.Cmd == "" {
			t.Error("Cmd is empty")
		}
		if proc.Pid != os.Getpid() {
			t.Error("Pid is not correct")
		}

	}
}

// same as TestReadProc but using NewProc
func TestNewProcessWatcher(t *testing.T) {
	proc, err := process.NewProc(os.Getpid())
	if err != nil {
		t.Error(err)
	} else {
		if proc.Cmd == "" {
			t.Error("Cmd is empty")
		}
		if proc.Pid != os.Getpid() {
			t.Error("Pid is not correct")
		}

	}
}

// same as TestReadProc but using NewProcessWatcher
func TestNewProcessWatcherWithProcess(t *testing.T) {
	pInfo, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Error(err)
		t.Skip("Can't find process")
	}
	procWatch, err := process.NewProcessWatcherByProcessInfo(pInfo)
	if err != nil {
		t.Error(err)
	} else {
		if procWatch.GetCmd() == "" {
			t.Error("Cmd is empty")
		}
		if procWatch.GetPid() != os.Getpid() {
			t.Error("Pid is not correct")
		}
	}
}

// test we can stop a process
// that will lauch a new process that will run infinitely
// the issue is that we have to kill all the child processes
func TestKillProcessTree(t *testing.T) {
	SkipOnGithubCi(t)
	if runtime.GOOS == "windows" {
		t.Skip("not supported on windows")
	}
	cmdLine := "watch date > just4checkHeDidSomething.tmp"
	command := process.NewProcess("bash", "-c", cmdLine)
	defer os.Remove("just4checkHeDidSomething.tmp")
	waitForIt := make(chan bool)

	wg := sync.WaitGroup{}
	wg.Add(2)

	// this goroutine will kill the process
	// after it started and waited for 100ms
	go func() {
		// use WaitForIt to wait for the process to start
		// before we kill it
		<-waitForIt
		// wait 100ms to kill the process
		// this command would run infinitely
		// if we don't kill it
		time.Sleep(100 * time.Millisecond)
		// kill the process with fire
		// this is the hard way and should result in a ExitCmdError 137
		if _, _, err := command.Kill(); err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	// running the executable in a goroutine
	// so we can wait for the command.Stop() to be called in the other goroutine
	go func() {
		internCode, realCode, err := command.Exec()

		// checking if the process watcher is set
		// and if it has childs, what it should have
		procHndl, perr := command.GetProcessWatcher()
		if perr != nil {
			t.Error(perr)
		}
		if procHndl == nil {
			t.Error("ProcessWatcher is nil")
		} else {
			pids := procHndl.GetChilds()
			if len(pids) == 0 {
				t.Error("ProcessWatcher has no childs")
			}
			t.Log("ProcessWatcher has childs:", pids)
		}

		if err != nil {
			// expectation:
			// exit status 137
			// ExitCode: -1
			if err.Error() != "exit status 137" {
				t.Error("unexpected error: ", err)
			}

		} else {
			t.Error("expected error but got none")
		}

		if internCode != systools.ExitCmdError {
			t.Error("internCode is not ExitCmdError. It is ", internCode)
		}
		if realCode != 137 {
			t.Error("realCode is not 137. It is ", realCode)
		}

		if CmdIsRunning(t, cmdLine) {
			t.Error("Process is still running")
		}
		wg.Done()
	}()
	waitForIt <- true

	wg.Wait()
	// last check is to check if the temporary file is still there
	// so we know the process was ever started
	if _, err := os.Stat("just4checkHeDidSomething.tmp"); err != nil {
		t.Error(err)
	}
}

func TestSoftKill(t *testing.T) {
	SkipOnGithubCi(t)
	if runtime.GOOS == "windows" {
		t.Skip("not supported on windows")
	}
	fileName := "testSoftKillTemp.tmp"     // this file will be created by the process
	watchCmd := "watch date > " + fileName // this command will run infinitely
	command := process.NewProcess("bash")  // we will just spawn a bash process
	defer os.Remove(fileName)              // remove the file after the test

	command.SetTimeout(250 * time.Millisecond) // wait not to long for the process to finish
	command.SetKeepRunning(true)               // we want to keep the process running

	// now run the command in the background.
	// so because SetKeepRunning is true we will not wait for the process to finish
	if internalExitCode, realExitCode, err := command.Exec(); err != nil {
		t.Error(err)
	} else {
		// internalExitCode should be ExitOk
		if internalExitCode != systools.ExitOk {
			t.Error("internalExitCode is not ExitOk. It is ", internalExitCode)
		}
		// realExitCode should be ExitTimeout
		if realExitCode != process.ExitInBackGround {
			t.Error("realExitCode is not ExitInBackGround. It is ", realExitCode)
		}
	}

	// we should get the process watcher
	// but we should wait a little bit
	// because the process needs some time to start
	time.Sleep(100 * time.Millisecond)
	procHndl, err := command.GetProcessWatcher()
	if err != nil {
		t.Error(err)
	}
	if procHndl == nil {
		t.Error("ProcessWatcher is nil")
	} else {
		// we should have a child process
		if err := procHndl.Update(); err != nil {
			t.Error(err)
		}
		pids := procHndl.GetChilds()
		if len(pids) == 0 {
			t.Error("ProcessWatcher has no childs")
		}
		t.Log("ProcessWatcher has childs:", pids)
	}

	// now we start the watch command
	if err := command.Command(watchCmd); err != nil {
		t.Error(err)
	}

	// again we wait a little bit. this time a little bit longer
	// just by experience
	time.Sleep(20 * time.Millisecond)
	if procHndl == nil {
		t.Error("(recheck) ProcessWatcher is nil")
	} else {
		// we should have a child process
		pids := procHndl.GetChilds()
		if len(pids) == 0 {
			t.Error("(recheck) ProcessWatcher has no childs")
		}
		t.Log("(rechek) ProcessWatcher has childs:", pids)
	}

	// wait til the process is done (what will never happen), the timeout is reached or the process is killed
	command.BlockWait(100 * time.Millisecond)

	if _, err := os.Stat(fileName); err != nil {
		t.Error("it seems the process was never running, because the tempfile is not there")
	}
	command.Stop()
}

func TestWaitForStopWorking(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on windows")
	}
	testLogger := NewMimicTestLogger()

	command := process.NewProcess("bash")
	command.SetKeepRunning(true)
	command.SetTimeout(100 * time.Millisecond)
	command.SetLogger(testLogger)
	command.Exec()

	command.Command("sleep 0.1")
	command.Command("exit")

	command.WaitUntilRunning(5 * time.Millisecond)
	if watch, err := command.GetProcessWatcher(); err != nil {
		t.Error(err)
	} else {
		if _, err := watch.WaitForStop(200*time.Millisecond, 10*time.Millisecond); err != nil {
			t.Error(err)
		}
	}
	testLogger.LogsToTestLog(t)
}
