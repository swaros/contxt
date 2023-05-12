package runner

import (
	"github.com/swaros/contxt/module/ctxshell"
)

func shellRunner(c *CmdExecutorImpl) {
	// run the context shell

	// add cobra commands to menu
	ctxshell.NewCshell().SetCobraRootCommand(c.session.Cobra.RootCmd).Run()
}
