package ctxshell_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/swaros/contxt/module/ctxshell"
	"github.com/swaros/contxt/module/systools"
)

func TestSimpleRun(t *testing.T) {
	shell := ctxshell.NewCshell()
	shutDownExecuted := false

	// doing all the test in the shutdown function
	shell.OnShutDownFunc(func() {
		shutDownExecuted = true
		if shell.GetLastInput() != "hello" {
			t.Error("expected hello, got[", shell.GetLastInput(), "]")
		}
	})

	if err := shell.RunOnce([]string{"hello"}); err != nil {
		t.Error(err)
	}

	if !shutDownExecuted {
		t.Error("shutdown function was not executed")
	}
}

// create a helper for easy testing commands
// and a callback to setup the shell

func helpCreateShell(initFn func(shell *ctxshell.Cshell)) *ctxshell.Cshell {
	shell := ctxshell.NewCshell()
	initFn(shell)
	return shell
}

func helpCreateShellAndExecute(initFn func(shell *ctxshell.Cshell), cmds ...string) error {
	shell := helpCreateShell(initFn)
	return shell.RunOnce(cmds)
}

// testing the helper functions and is also a template
// for any command based test
func TestInternHelperIsWorking(t *testing.T) {

	testFunction := func(shell *ctxshell.Cshell) {
		// do the setup here
		// like shell.AddCommand(...)

		// add a shutdown function
		shell.OnShutDownFunc(func() {
			expected := "hello"
			if shell.GetLastInput() != expected {
				t.Error("expected '", expected, "', got[", shell.GetLastInput(), "]")
			}
		})
	}

	helpCreateShellAndExecute(testFunction, "hello")
}

// testing native commands
func TestPromptUpdate(t *testing.T) {
	gotNotifiedByHello := false
	gotNotifiedByWorld := false
	testFunction := func(shell *ctxshell.Cshell) {
		// do the setup here

		shell.AddNativeCmd(ctxshell.NewNativeCmd("hello", "the hello function", func(args []string) error {
			gotNotifiedByHello = true
			return nil
		}))

		shell.AddNativeCmd(ctxshell.NewNativeCmd("world", "the world function", func(args []string) error {
			gotNotifiedByWorld = true
			return nil
		}))

		// add a shutdown function
		shell.OnShutDownFunc(func() {
			expected := "world"
			if shell.GetLastInput() != expected {
				t.Error("expected '", expected, "', got[", shell.GetLastInput(), "]")
			}
		})
	}

	helpCreateShellAndExecute(testFunction, "hello", "world")

	if !gotNotifiedByHello {
		t.Error("did not get notified from hello command")
	}

	if !gotNotifiedByWorld {
		t.Error("did not get notified from world command")
	}
}

func TestPromtMessage(t *testing.T) {

	messageBuffer := []string{}
	testFunction := func(shell *ctxshell.Cshell) {
		// add the promt handler so we can get notified
		shell.SetPromptFunc(func(reason int) string {
			if reason == ctxshell.UpdateByNotify {
				if found, msg := shell.GetCurrentMessage(); found {
					messageBuffer = append(messageBuffer, msg.GetMsg())
					return "got a message:>"
				}
			}
			return "promtp:>"
		})
		// enable the prompt update by notify and set the update period
		shell.UpdatePromptEnabled(true).
			UpdatePromptPeriod(time.Millisecond * 10)

			// define a command that will just wait longer then the update period
		shell.AddNativeCmd(ctxshell.NewNativeCmd("wait", "wait 20 milliseconds", func(args []string) error {
			time.Sleep(time.Millisecond * 20)
			return nil
		}))

		// add a shutdown function
		shell.OnShutDownFunc(func() {
			expected := "hello"
			if shell.GetLastInput() != expected {
				t.Error("expected '", expected, "', got[", shell.GetLastInput(), "]")
			}
		})
	}

	helpCreateShellAndExecute(testFunction, "test", "wait", "hello")
	if !systools.SliceContains(messageBuffer, "unknown command: test") {
		t.Error("message buffer contains test:", messageBuffer)
	}
}

func TestNativeCmdWithError(t *testing.T) {
	errorTriggered := false
	testFunction := func(shell *ctxshell.Cshell) {
		// do the setup here

		shell.AddNativeCmd(ctxshell.NewNativeCmd("hello", "the hello function", func(args []string) error {
			return fmt.Errorf("hello error")
		}))

		shell.OnErrorFunc(func(err error) {
			errorTriggered = true
			expectedError := "error executing native command: hello error"
			if err.Error() != expectedError {
				t.Error("expected '", expectedError, "', got[", err.Error(), "]")
			}

		})

		// add a shutdown function
		shell.OnShutDownFunc(func() {
			expected := "hello"
			if shell.GetLastInput() != expected {
				t.Error("expected '", expected, "', got[", shell.GetLastInput(), "]")
			}
		})
	}

	helpCreateShellAndExecute(testFunction, "hello")

	if !errorTriggered {
		t.Error("did not get notified from error")
	}
}

func TestNativeCmdWithErrorAsync(t *testing.T) {
	errorTriggered := false
	testFunction := func(shell *ctxshell.Cshell) {
		// do the setup here

		shell.AddNativeCmd(ctxshell.NewNativeCmd("hello", "the hello function", func(args []string) error {
			return fmt.Errorf("hello error")
		}))

		shell.AddNativeCmd(ctxshell.NewNativeCmd("wait", "wait 20 milliseconds", func(args []string) error {
			time.Sleep(time.Millisecond * 20)
			return nil
		}))

		shell.OnErrorFunc(func(err error) {
			errorTriggered = true
			expectedError := "error executing native command: hello error"
			if err.Error() != expectedError {
				t.Error("expected '", expectedError, "', got[", err.Error(), "]")
			}

		})
		// enable async execution. but disable it for the wait command
		shell.SetAsyncNativeCmd(true).SetNeverAsyncCmd("wait")
		// this way we have the wait in place just to make sure
		// we get enough time to handle the error before the shell is destroyed

		// add a shutdown function
		shell.OnShutDownFunc(func() {
			expected := "wait"
			if shell.GetLastInput() != expected {
				t.Error("expected '", expected, "', got[", shell.GetLastInput(), "]")
			}
		})
	}

	helpCreateShellAndExecute(testFunction, "hello", "wait")

	if !errorTriggered {
		t.Error("did not get notified from error")
	}
}

func TestPromtMessageLoop(t *testing.T) {

	messageBuffer := []string{}
	testFunction := func(shell *ctxshell.Cshell) {
		// add the promt handler so we can get notified
		shell.SetPromptFunc(func(reason int) string {
			if reason == ctxshell.UpdateByNotify {
				if found, msg := shell.GetCurrentMessage(); found {
					messageBuffer = append(messageBuffer, msg.GetMsg())
					return "got a message:>"
				}
			}
			return "promtp:>"
		})
		// any message is displayed for 10 milliseconds
		shell.SetMessageDisplayTime(time.Millisecond * 10)

		// enable the prompt update by notify and set the update period
		shell.UpdatePromptEnabled(true).
			UpdatePromptPeriod(time.Millisecond * 10)

			// define a command that will just wait longer then the update period
		shell.AddNativeCmd(ctxshell.NewNativeCmd("wait", "wait a couple of milliseconds", func(args []string) error {
			time.Sleep(time.Millisecond * 60)
			return nil
		}))

		// add a shutdown function
		shell.OnShutDownFunc(func() {
			expected := "wait"
			if shell.GetLastInput() != expected {
				t.Error("expected '", expected, "', got[", shell.GetLastInput(), "]")
			}
		})
	}

	helpCreateShellAndExecute(testFunction, "test", "lala", "lulu", "gaga", "wait")
	if !systools.SliceContains(messageBuffer, "unknown command: test") {
		t.Error("message buffer do not contains test:", messageBuffer)
	}
	expectedNotFound := []string{"lala", "lulu", "gaga"}
	for _, expected := range expectedNotFound {
		if !systools.SliceContains(messageBuffer, "unknown command: "+expected) {
			t.Error("message buffer do not contains", expected, ":", messageBuffer)
		}
	}
}

func TestHooksExecuted(t *testing.T) {
	messageBuffer := []string{}
	testFunction := func(shell *ctxshell.Cshell) {
		// add the promt handler so we can get notified
		shell.SetPromptFunc(func(reason int) string {
			if reason == ctxshell.UpdateByNotify {
				if found, msg := shell.GetCurrentMessage(); found {
					messageBuffer = append(messageBuffer, msg.GetMsg())
					return "got a message:>"
				}
			}
			return "promtp:>"
		})
		// any message is displayed for 10 milliseconds
		shell.SetMessageDisplayTime(time.Millisecond * 10)

		// enable the prompt update by notify and set the update period
		shell.UpdatePromptEnabled(true).
			UpdatePromptPeriod(time.Millisecond * 10)

			// define a command that will just wait longer then the update period
		shell.AddNativeCmd(ctxshell.NewNativeCmd("wait", "wait a couple of milliseconds", func(args []string) error {
			time.Sleep(time.Millisecond * 60)
			return nil
		}))

		shell.AddHook(ctxshell.NewHook("tes*", func() error {
			messageBuffer = append(messageBuffer, "before test")
			return nil
		}, func() error {
			messageBuffer = append(messageBuffer, "after test")
			return nil
		}))

		// add a shutdown function
		shell.OnShutDownFunc(func() {
			expected := "wait"
			if shell.GetLastInput() != expected {
				t.Error("expected '", expected, "', got[", shell.GetLastInput(), "]")
			}
		})
	}

	helpCreateShellAndExecute(testFunction, "test", "lala", "lulu", "gaga", "wait")
	if !systools.SliceContains(messageBuffer, "unknown command: test") {
		t.Error("message buffer do not contains test:", messageBuffer)
	}

	expectedBeFound := []string{"before test", "after test"}
	for _, expected := range expectedBeFound {
		if !systools.SliceContains(messageBuffer, expected) {
			t.Error("message buffer do not contains", expected, ":", messageBuffer)
		}
	}

	expectedNotFound := []string{"lala", "lulu", "gaga"}
	for _, expected := range expectedNotFound {
		if !systools.SliceContains(messageBuffer, "unknown command: "+expected) {
			t.Error("message buffer do not contains", expected, ":", messageBuffer)
		}
	}

}
