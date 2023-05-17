package runner

import (
	"strings"
	"time"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/ctxshell"
	"github.com/swaros/contxt/module/dirhandle"
)

func shellRunner(c *CmdExecutorImpl) {
	// run the context shell
	shell := ctxshell.NewCshell()
	//// add cobra commands to menu
	shell.SetCobraRootCommand(c.session.Cobra.RootCmd)

	// add native commands to menu
	// this one is for testing only
	demoCmd := ctxshell.NewNativeCmd("demo", "demo command", func(args []string) error {
		c.Println("demo command executed:", strings.Join(args, " "))
		for i := 0; i < 5000; i++ {
			time.Sleep(10 * time.Millisecond)
			c.Println("i do something .. we are in round ", i)
		}
		return nil
	})

	// while developing, you can use this to test the completer
	// and the command itself
	demoCmd.SetCompleterFunc(func(line string) []string {
		return []string{"demo"}
	})

	shell.AddNativeCmd(demoCmd)

	// set the prompt handler
	shell.SetPromptFunc(func() string {
		tpl := ""
		if dir, err := dirhandle.Current(); err == nil {
			tpl = dir
		}
		template, exists, err := c.session.TemplateHndl.Load()
		if err != nil {

			c.session.Log.Logger.Error(err)
			return ctxout.ToString(
				ctxout.NewMOWrap(),
				ctxout.BackYellow,
				ctxout.ForeRed,
				"error loading template: ",
				ctxout.BackRed,
				ctxout.ForeYellow,
				err.Error(),
				ctxout.BackBlue,
				ctxout.ForeRed,
				"",
				"<f:white><b:blue>",
				tpl,
				"</><f:blue></> ",
			)
		}
		if exists {
			tpl = template.Workspace.Project
			if template.Workspace.Role != "" {
				tpl += "/" + template.Workspace.Role
			}
		} else {
			tpl = "no template"
		}
		return ctxout.ToString(
			ctxout.NewMOWrap(),
			ctxout.BackWhite,
			ctxout.ForeBlue,
			tpl,
			"<f:white><b:blue>ctx<f:yellow>shell:</><f:blue></> ",
		)
	})
	c.session.OutPutHdnl = shell
	// start the shell
	shell.SetAsyncCobraExec(true).SetAsyncNativeCmd(true).Run()
}
