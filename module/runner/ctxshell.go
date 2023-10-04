// Copyright (c) 2023 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// # Licensed under the MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package runner

import (
	"runtime"
	"strings"
	"time"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/ctxshell"
	"github.com/swaros/contxt/module/dirhandle"
)

func shellRunner(c *CmdExecutorImpl) {
	// run the context shell
	shell := ctxshell.NewCshell()

	// add cobra commands to the shell, so they can be used there too.
	// first we need to define the exceptions
	// we do not want to have in the menu
	shell.SetIgnoreCobraCmd("completion", "interactive")
	// afterwards we can add the commands by injecting the root command
	shell.SetCobraRootCommand(c.session.Cobra.RootCmd)

	// some of the commands are not working well async, because they
	// are switching between workspaces. so we have to disable async
	// for them
	shell.SetNeverAsyncCmd("workspace", "dir")

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
			if runtime.GOOS == "windows" {
				return windowsPrompt(tpl, err)
			} else {
				return linuxPrompt(tpl, err)
			}
		}
		if exists {
			tpl = template.Workspace.Project
			if template.Workspace.Role != "" {
				tpl += "/" + template.Workspace.Role
			}
		} else {
			tpl = "no template"
		}
		// depends runtime.GOOS
		if runtime.GOOS == "windows" {
			return windowsPrompt(tpl, nil)
		} else {
			return linuxPrompt(tpl, nil)
		}
	})
	c.session.OutPutHdnl = shell
	// start the shell
	shell.SetAsyncCobraExec(true).SetAsyncNativeCmd(true).Run()
}

func windowsPrompt(tpl string, err error) string {
	if err != nil {
		return ctxout.ToString(
			ctxout.NewMOWrap(),
			ctxout.ForeRed,
			"error loading template: ",
			ctxout.ForeYellow,
			err.Error(),
			ctxout.ForeRed,
			" › ",
			ctxout.ForeBlue,
			tpl,
			ctxout.CleanTag,
		)
	}

	return ctxout.ToString(
		ctxout.NewMOWrap(),
		ctxout.ForeBlue,
		tpl,
		" ",
		ctxout.ForeCyan,
		"› ",
		ctxout.CleanTag,
	)
}

func linuxPrompt(tpl string, err error) string {
	if err != nil {
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

	return ctxout.ToString(
		ctxout.NewMOWrap(),
		ctxout.BackWhite,
		ctxout.ForeBlue,
		tpl,
		ctxout.BackBlue,
		ctxout.ForeWhite,
		"ctx<f:yellow>shell:</><f:blue></> ",
	)
}
