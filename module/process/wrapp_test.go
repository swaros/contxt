package process_test

import (
	"os"
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

func TestKillProcessTree(t *testing.T) {
	command := process.NewProcess("bash", "-c", "watch date > just4checkHeDidSomething.tmp")
	defer os.Remove("just4checkHeDidSomething.tmp")
	waitForIt := make(chan bool)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		// use WaitForIt to wait for the process to start
		// before we kill it
		<-waitForIt

		// wait 100ms to kill the process
		// this command would run infinitely
		// if we don't kill it

		time.Sleep(100 * time.Millisecond)
		// kill the process with fire
		if _, _, err := command.Stop(); err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	go func() {
		internCode, realCode, err := command.Exec()

		if err != nil {
			// expectation:
			// Signal: killed
			// ExitCode: -1
			if err.Error() != "signal: killed" {
				t.Error("unexpected error: ", err)
			}

		} else {
			t.Error("expected error but got none")
		}
		if internCode != systools.ExitCmdError {
			t.Error("internCode is not ExitCmdError. It is ", internCode)
		}
		if realCode != -1 {
			t.Error("realCode is not -1. It is ", realCode)
		}

		if CmdIsRunning(t, "watch date") {
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
