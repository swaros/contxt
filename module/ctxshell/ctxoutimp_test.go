package ctxshell_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/ctxshell"
)

func TestOutImpBase(t *testing.T) {
	shell := ctxshell.NewCshell()

	// we have to use RunOnceWithCmd, because without
	// an initialized readline, the output will be
	// written to stdout. this is the fallback for non interactive
	// shells like the one used in the tests.
	shell.RunOnceWithCmd(func() {
		ctxout.PrintLn(shell, "hello")
	})

}
