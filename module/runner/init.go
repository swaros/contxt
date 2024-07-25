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
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
)

func setShutDownBehavior() {
	// add exit listener for shutting down all processes
	systools.AddExitListener("main", func(code int) systools.ExitBehavior {
		ctxout.PrintLn(ctxout.NewMOWrap(), ctxout.ForeDarkGrey, " stop all tasks: ", ctxout.CleanTag)
		tasks.ShutDownProcesses(func(target string, time int, succeed bool) {
			if succeed {
				ctxout.PrintLn(ctxout.NewMOWrap(), ctxout.ForeDarkGrey, "  task stopped: ", ctxout.ForeBlue, target, ctxout.CleanTag)
			} else {
				ctxout.PrintLn(ctxout.NewMOWrap(), ctxout.ForeDarkGrey, "  stop failure: ", ctxout.ForeRed, target, ctxout.CleanTag)
			}
		})
		return systools.Continue
	})

	// add exit listener for shutting down all processes
	// different to the main listener, this one will kill all processes
	// that are not stopped by the main listener.
	// this depends on the behavior of the system, where it is hard to get all child processes
	// so we use the HandleAllMyPid function to get all child processes.
	// this function wraps the ps command and filters the output for the current pid on linux.
	systools.AddExitListener("killProcs", func(code int) systools.ExitBehavior {
		ctxout.PrintLn(ctxout.NewMOWrap(), ctxout.ForeDarkGrey, " Cleanup all child Processes if possible. ", ctxout.CleanTag)
		tasks.HandleAllMyPid(func(pid int) error {

			if proc, err := os.FindProcess(pid); err == nil {
				if err := proc.Kill(); err != nil {
					if err == os.ErrProcessDone {
						ctxout.PrintLn(
							ctxout.NewMOWrap(),
							ctxout.ForeDarkGrey,
							"  task is already stopped: ",
							ctxout.ForeCyan,
							pid,
							ctxout.CleanTag,
						)
					} else {
						ctxout.PrintLn(
							ctxout.NewMOWrap(),
							ctxout.ForeDarkGrey,
							"  error while stooping task: ",
							ctxout.ForeCyan,
							pid,
							ctxout.ForeRed,
							err.Error(),
							ctxout.CleanTag,
						)
					}
					return err
				} else {
					ctxout.PrintLn(ctxout.NewMOWrap(), ctxout.ForeDarkGrey, "  stopped: ", ctxout.ForeBlue, pid, ctxout.CleanTag)
				}
			} else {
				ctxout.PrintLn(ctxout.NewMOWrap(), ctxout.ForeDarkGrey, "  failed to stop: ", ctxout.ForeRed, pid, ctxout.CleanTag)
				return err
			}
			return nil
		})
		return systools.Continue
	})
	// capture the sigterm signal so we are able to cleanup all processes
	// nil means that we use the default behavior for the exit control flow
	// so any systool.Exit() call will trigger the exit listeners
	// this is experimental and is only enabled if the env CTX_SUTDOWN_BEHAVIOR is set to "true"
	// this is a workaround for the problem that the exit listeners are not called if the application
	// is killed by the system.
	if os.Getenv("CTX_SUTDOWN_BEHAVIOR") == "true" {
		systools.WatchSigTerm(nil)
	}
}

// Init initializes the application
// and starts the main loop
func Init() error {
	// create the application session
	app := NewCmdSession()

	// set the TemplateHndl OnLoad function to parse required files
	onLoadFn := func(template *configure.RunConfig) error {
		return app.SharedHelper.MergeRequiredPaths(template, app.TemplateHndl)
	}
	app.TemplateHndl.SetOnLoad(onLoadFn)

	// set the default log level
	app.Log.Logger.SetLevel(logrus.ErrorLevel)
	// create the the command executor instance
	functions := NewCmd(app)

	// add support for utf-8 signs
	glyps := ctxout.NewSignFilter(nil)
	glyps.AddSign(ctxout.Sign{Glyph: "🭬", Name: "runident", Fallback: "»"})
	glyps.AddSign(ctxout.Sign{Glyph: "🭮", Name: "stopident", Fallback: "«"})
	glyps.AddSign(ctxout.Sign{Glyph: "", Name: "prompt", Fallback: "»"})
	glyps.AddSign(ctxout.Sign{Glyph: "⠄⠆⠇⠋⠙⠸⠰⠠⠐⠈", Name: "pbar", Fallback: "-_\\|/"})
	ctxout.AddPostFilter(glyps)

	// enable the sign filter if possible
	// in current ctxout version, must be done before NewTabOut
	if runtime.GOOS != "windows" && systools.IsStdOutTerminal() {
		// check if unicode is supported. if not, disable the sign filter
		// this is the only way i see to check for unicode support
		code1, code2, errorEx := tasks.Execute("bash", []string{"-c"}, "echo -e $TERM", func(s string, err error) bool {
			if err == nil {
				terminalsTheyNotSupportUt8Chars := []string{"xterm", "screen", "tmux"}
				// if one of these terminals is used, we do not enable the sign filter
				for _, term := range terminalsTheyNotSupportUt8Chars {
					if s == term {
						return false
					}
				}
			}
			glyps.Enable()
			return true
		}, func(p *os.Process) {

		})

		// just to be sure, if anything gos wrong by checking the terminal, we disable the sign filter
		if errorEx != nil && code1 == 0 && code2 == 0 {
			glyps.Disable()
		}
	}

	// set the default output filter
	ctxout.AddPostFilter(ctxout.NewTabOut())

	// initialize the application functions
	functions.MainInit()

	// set the shutdown behavior
	setShutDownBehavior()
	// initialize the cobra commands
	if err := app.Cobra.Init(functions); err != nil {
		return err
	}
	// and execute the root command
	if err := app.Cobra.RootCmd.Execute(); err != nil {
		return err
	}
	// show variables if the verbose flag is set
	if app.Cobra.Options.ShowVars {
		functions.PrintVariables("%s=%s[nl]")
	}
	return nil
}
