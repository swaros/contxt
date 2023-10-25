package ctxshell_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxshell"
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
