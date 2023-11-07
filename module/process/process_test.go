package process_test

import (
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/swaros/contxt/module/process"
	"github.com/swaros/contxt/module/process/terminal"
	"github.com/swaros/contxt/module/systools"
)

func TestBasicRun(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}
	process := process.NewProcess("bash", "-c", "echo 'Hello World'")
	if _, _, err := process.Exec(); err != nil {
		t.Error(err)
	}

}

func TestBasicRunWithTermFind(t *testing.T) {
	term, err := terminal.GetTerminal()
	if err != nil {
		t.Error(err)
		t.SkipNow()
	}
	process := process.NewProcess(term.GetCmd(), term.CombineArgs(`echo "Hello World"`)...)
	if _, rcode, err := process.Exec(); err != nil {
		t.Error(err)
	} else {
		if rcode != 0 {
			t.Error("rcode is not 0. It is ", rcode)
		}
	}
}

func TestNewTerminal(t *testing.T) {
	process := process.NewTerminal(`echo "Hello World"`)
	if _, rcode, err := process.Exec(); err != nil {
		t.Error(err)
	} else {
		if rcode != 0 {
			t.Error("rcode is not 0. It is ", rcode)
		}
	}
}

func TestNewTerminalError(t *testing.T) {
	process := process.NewTerminal(`notExistsCmd`)
	if intCode, rcode, err := process.Exec(); err != nil {
		if intCode != systools.ExitCmdError {
			t.Error("intCode is not ", systools.ExitCmdError, ". It is ", intCode)
		}

		// this is linux specific
		if runtime.GOOS == "linux" {
			if rcode != 127 {
				t.Error("rcode is not 127. It is ", rcode)
			}

			if err.Error() != "exit status 127" {
				t.Error("err is not 'exit status 127'. It is ", err.Error())
			}
		}

	} else {
		t.Error("Error is nil. error is expected")
	}

}

func TestBasicRunButCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}
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
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}
	process := process.NewProcess("bash")
	process.AddStartCommands("echo 'Hello World'", "echo 'Hello World 2'")
	if _, _, err := process.Exec(); err != nil {
		t.Error(err)
	}
	if _, _, err := process.Stop(); err != nil {
		t.Error(err)
	}

}

func TestRunWithArgsAndTerminal(t *testing.T) {

	process := process.NewTerminal()
	process.AddStartCommands("echo 'Hello World'", "echo 'Hello World 2'")

	// track any output
	outputs := []string{}
	process.SetOnOutput(func(msg string, err error) bool {
		outputs = append(outputs, msg)
		return true
	})

	if _, _, err := process.Exec(); err != nil {
		t.Error(err)
	}

	expectedCount := 2
	if runtime.GOOS == "windows" {
		expectedCount = 5 // windows has some additional output
	}

	// did we get the output?
	if len(outputs) != expectedCount {
		t.Errorf("outputs is not %d. It is %d", expectedCount, len(outputs))
		t.Log("outputs: ", strings.Join(outputs, "\n"))
	} else {
		if runtime.GOOS == "windows" {
			if outputs[1] != "Hello World" {
				t.Error("outputs[1] is not 'Hello World'. It is ", outputs[1])
			}
			if outputs[3] != "Hello World 2" {
				t.Error("outputs[2] is not 'Hello World 2'. It is ", outputs[3])
			}
		} else {
			if outputs[0] != "Hello World" {
				t.Error("outputs[0] is not 'Hello World'. It is ", outputs[0])
			}
			if outputs[1] != "Hello World 2" {
				t.Error("outputs[1] is not 'Hello World 2'. It is ", outputs[1])
			}
		}
	}
}

func TestExecWithBash(t *testing.T) {
	process := process.NewTerminal()
	process.AddStartCommands("echo 'Hello World'", "echo 'Hello World 2'")
	// check if the OnInit is called
	// the output is tested a couple of times in other tests
	// so no need to test it here
	initIsCalled := false
	process.SetOnInit(func(proc *os.Process) {
		initIsCalled = true
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

	if !initIsCalled {
		t.Error("OnInit is not called")
	}
}

func TestExecWithBashAndStayOpen(t *testing.T) {
	outPuts := []string{}
	proc := process.NewTerminal()
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
	proc.WaitUntilRunning(500 * time.Millisecond)
	proc.Stop()
	// on windows we have a total different output then on linux
	// so we have to check for the output
	if runtime.GOOS == "windows" {
		// the powershell bevahe all the time differently depending where it is executed.
		// so instead of checking the line count and checking ther text on a specific line,
		// we just check if the output contains the expected text
		if len(outPuts) < 3 {
			t.Error("outPuts is not 3. It is ", len(outPuts))
			expected := []string{"Hello World", "test 2"}
			hit := 0
			for i, out := range outPuts {
				if strings.Contains(out, expected[0]) {
					hit++
				}
				if strings.Contains(out, expected[1]) {
					hit++
				}
				t.Log("output[", i, "]: ", out)
			}
			if hit != 2 {
				t.Error("hit is not 2. It is ", hit)
			}
		}

	} else {
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

}

// testing getting errors from the process
// while running in the background.
// this means we have to watch the output
// if they are written to the error channel.
// this indicates that the process executes a command that
// is not available, have a typo, exists with an error code and so on
func TestExecWithBashAndStayOpenAndError(t *testing.T) {

	mimicTestLog := NewMimicTestLogger()
	mimicTestLog.SetLevel("info")

	outPuts := []string{}
	errors := []error{}
	proc := process.NewTerminal()
	proc.SetKeepRunning(true)
	proc.SetLogger(mimicTestLog)
	proc.SetOnOutput(func(msg string, err error) bool {

		if err != nil {
			mimicTestLog.Error(err)
			errors = append(errors, err)
			return false
		} else {
			mimicTestLog.Info(msg)
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

	proc.WaitUntilRunning(10 * time.Millisecond)

	if err := proc.Command("echo 'Hello World'"); err != nil {
		t.Error(err)
	}

	if err := proc.Command("notACommand"); err != nil {
		t.Error(err)
	}
	// give the process some time to execute the command
	proc.Stop()

	if runtime.GOOS == "windows" {
		if len(outPuts) != 3 {
			t.Error("outPuts is not 3. It is ", len(outPuts))
		} else {
			if outPuts[1] != "Hello World" {
				t.Error("outPuts[1] is not 'Hello World'. It is ", outPuts[1])
			}
		}
	} else {
		if len(outPuts) != 1 {
			t.Error("outPuts is not 2. It is ", len(outPuts))
		} else {
			if outPuts[0] != "Hello World" {
				t.Error("outPuts[0] is not 'Hello World'. It is ", outPuts[0])
			}
		}
	}
	if len(errors) < 1 {
		t.Error("errors is not 1. It is ", len(errors), " errors: ", errors)
	} else {
		if !strings.Contains(errors[0].Error(), "notACommand") {
			t.Error("errors[0] does not contain 'notACommand'. It is ", errors[0].Error())
		}

	}

	mimicTestLog.LogsToTestLog(t)
}

func TestTimeOut(t *testing.T) {
	// create a simple bash process
	proc := process.NewProcess("bash")
	proc.SetKeepRunning(true)
	proc.SetTimeout(100 * time.Millisecond)

	if _, _, err := proc.Exec(); err != nil {
		t.Error(err)
		t.SkipNow()
	}
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
