package terminal_test

import (
	"runtime"
	"testing"

	"github.com/swaros/contxt/module/process/terminal"
)

func TestFindOsTerminal(t *testing.T) {

	term, err := terminal.GetTerminal()
	if err != nil {
		t.Error(err)
	} else {
		if term.GetCmd() == "" {
			t.Error("Cmd is empty")
		}
		if len(term.GetArgs()) == 0 {
			t.Error("Args is empty")
		}
		if runtime.GOOS == "windows" {
			if term.GetCmd() != "powershell" {
				t.Error("Cmd is not powershell. It is ", term.GetCmd())
			}
			if len(term.GetArgs()) != 2 {
				t.Error("Args is not 2. It is ", len(term.GetArgs()))
			} else {
				if term.GetArgs()[0] != "-nologo" {
					t.Error("Args is not -nologo. It is ", term.GetArgs()[0])
				}
				if term.GetArgs()[1] != "-noprofile" {
					t.Error("Args is not -noprofile. It is ", term.GetArgs()[1])
				}
			}
		}

		if runtime.GOOS == "linux" {
			if term.GetCmd() != "bash" {
				t.Error("Cmd is not bash. It is ", term.GetCmd())
			}
			if len(term.GetArgs()) != 1 {
				t.Error("Args is not 1. It is ", len(term.GetArgs()))
			} else {
				if term.GetArgs()[0] != "-c" {
					t.Error("Args is not -c. It is ", term.GetArgs()[0])
				}
			}

			cmdWithArgs := term.CombineArgs("echo 'Hello World'")
			if len(cmdWithArgs) != 2 {
				t.Error("CmdWithArgs is not 2. It is ", len(cmdWithArgs))
			} else {

				if cmdWithArgs[0] != "-c" {
					t.Error("CmdWithArgs is not -c. It is ", cmdWithArgs[1])
				}
				if cmdWithArgs[1] != "echo 'Hello World'" {
					t.Error("CmdWithArgs is not echo 'Hello World'. It is ", cmdWithArgs[2])
				}
			}
		}
	}
}
