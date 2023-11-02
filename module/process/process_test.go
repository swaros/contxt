package process_test

import (
	"os"
	"testing"
	"time"

	"github.com/swaros/contxt/module/process"
	"github.com/swaros/contxt/module/systools"
)

func TestBasicRun(t *testing.T) {
	process := process.NewProcess("bash", "-c", "echo 'Hello World'")
	if _, _, err := process.Exec(); err != nil {
		t.Error(err)
	}

}

func TestBasicRunButCommand(t *testing.T) {
	process := process.NewProcess("bash", "-c", "echo 'Hello World'")

	if err := process.Command("echo 'Hello World 2'"); err == nil {
		t.Error("Error is nil")
	} else {
		if err.Error() != "process is not started" {
			t.Error("Error is not 'process is not started'. It is ", err.Error())
		}
	}
	if _, _, err := process.Exec(); err != nil {
		t.Error(err)
	}

	if err := process.Command("echo 'Hello World 2'"); err == nil {
		t.Error("Error is nil")
	} else {
		if err.Error() != "process is not set to stay open" {
			t.Error("Error is not 'process is not set to stay open'. It is ", err.Error())
		}
	}
}

func TestBasicRunWithError(t *testing.T) {
	process := process.NewProcess("notExists")
	if inCode, realCode, err := process.Exec(); err == nil {
		t.Error("Error is nil")
	} else {
		if realCode != -1 {
			t.Error("realCode is not -1. It is ", realCode)
		}
		if inCode != systools.ExitCmdError {
			t.Error("internal Code is not ", systools.ExitCmdError, ". It is ", inCode)
		}
	}

}

func TestRunWithArgs(t *testing.T) {
	process := process.NewProcess("bash")
	process.AddStartCommands("echo 'Hello World'", "echo 'Hello World 2'")
	if _, _, err := process.Exec(); err != nil {
		t.Error(err)
	}

}

func TestExecWithBash(t *testing.T) {
	process := process.NewProcess("bash")
	process.AddStartCommands("echo 'Hello World'", "echo 'Hello World 2'")
	process.SetOnOutput(func(msg string, err error) bool {
		t.Log("output[", msg, "]")
		return true
	})
	process.SetOnInit(func(proc *os.Process) {
		if proc == nil {
			t.Error("Process is nil")
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
	outPuts := []string{}
	proc := process.NewProcess("bash")
	proc.SetKeepRunning(true)
	proc.SetOnOutput(func(msg string, err error) bool {
		t.Log("output[", msg, "]")
		outPuts = append(outPuts, msg)
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

	if len(outPuts) != 2 {
		t.Error("outPuts is not 2. It is ", len(outPuts))
	} else {
		if outPuts[0] != "Hello World" {
			t.Error("outPuts[0] is not 'Hello World'. It is ", outPuts[0])
		}
		if outPuts[1] != "test 2" {
			t.Error("outPuts[1] is not 'test 2'. It is ", outPuts[1])
		}
	}

}

func TestExecWithBashAndStayOpenAndError(t *testing.T) {
	outPuts := []string{}
	errors := []error{}
	proc := process.NewProcess("bash")
	proc.SetKeepRunning(true)
	proc.SetOnOutput(func(msg string, err error) bool {
		t.Log("MESSAGE:[", msg, "]")
		if err != nil {
			errors = append(errors, err)
			return false
		} else {
			outPuts = append(outPuts, msg)
			return true
		}
	})
	proc.SetOnInit(func(proc *os.Process) {
		if proc == nil {
			t.Error("Process is nil")
		}
	})

	if _, _, err := proc.Exec(); err != nil {
		t.Error(err)
		t.SkipNow()
	}

	if err := proc.Command("echo 'Hello World'"); err != nil {
		t.Error(err)
	}

	if err := proc.Command("notACommand"); err != nil {
		t.Error(err)
	}
	// give the process some time to execute the command
	time.Sleep(100 * time.Millisecond)
	internCode, realCode, err := proc.Stop()

	// this should fail because the command is not found
	// so bash will exit with 127
	if err != nil {
		// expected error is "exit status 127"
		if err.Error() != "exit status 127" {
			t.Error("unexpected error: ", err)
		}
	} else {
		t.Error("expected error but got none")
	}
	if realCode != 127 {
		t.Error("realCode is not 127. It is ", realCode)
	}
	if internCode != 103 {
		t.Error("internCode is not 103, It is ", internCode)
	}

	if len(outPuts) != 2 {
		t.Error("outPuts is not 2. It is ", len(outPuts))
	} else {
		if outPuts[0] != "Hello World" {
			t.Error("outPuts[0] is not 'Hello World'. It is ", outPuts[0])
		}
	}

}

func TestTimeOut(t *testing.T) {
	proc := process.NewProcess("bash")
	proc.SetKeepRunning(true)
	proc.SetTimeout(100 * time.Millisecond)

	if _, _, err := proc.Exec(); err != nil {
		t.Error(err)
		t.SkipNow()
	}
	messureStartTimeout := time.Now()
	if err := proc.Command("sleep 10"); err != nil {
		t.Error(err)
	}
	// using BlockWait to wait for the process to stop
	if err := proc.BlockWait(10 * time.Millisecond); err != nil {
		t.Error(err)
	}
	internCode, realCode, err := proc.Stop()
	expectedInternCode := process.ExitTimeout
	expectedRealCode := process.RealCodeNotPresent

	// check the needed time. it should be around 100ms
	// but we give it 300ms to be sure
	if time.Since(messureStartTimeout) > 300*time.Millisecond {
		t.Error("timeout is not working")
	}

	if err != nil {
		if err.Error() != "process stopped by timeout" {
			t.Error("unexpected error: ", err)
		}
	} else {
		t.Error("expected error but got none")
	}

	if realCode != expectedRealCode {
		t.Error("realCode is not ", expectedRealCode, ". It is ", realCode)
	}
	if internCode != expectedInternCode {
		t.Error("internCode is not ", expectedInternCode, ", It is ", internCode)
	}
}
