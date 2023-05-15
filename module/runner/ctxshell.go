package runner

import (
	"strings"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/ctxshell"
)

func shellRunner(c *CmdExecutorImpl) {
	// run the context shell
	shell := ctxshell.NewCshell()
	//// add cobra commands to menu
	shell.SetCobraRootCommand(c.session.Cobra.RootCmd)

	// add native commands to menu
	demoCmd := ctxshell.NewNativeCmd("demo", "demo command", func(args []string) error {
		ctxout.PrintLn("demo command executed:", strings.Join(args, " "))
		return nil
	})
	demoCmd.SetCompleterFunc(func(line string) []string {
		return []string{"demo"}
	})

	shell.AddNativeCmd(demoCmd)

	// set the prompt handler
	shell.SetPromptFunc(func() string {
		return ctxout.ToString("<f:yellow>contxt</> > ")
	})

	// start the shell
	shell.Run()
}
