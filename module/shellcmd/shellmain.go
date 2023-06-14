// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

package shellcmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/taskrun"
	"github.com/swaros/manout"
)

var (
	runCmdAdded = false
	forceExit   = false
)

func RunIShell() {
	if !systools.IsStdOutTerminal() {
		noShellScreen()
		return
	}
	taskrun.MainInit()

	shell := ishell.New()

	// display welcome info.
	headScreen(shell)
	runCmdAdded = CreateRunCommands(shell)
	CreateDefaultComands(shell)
	CreateWsCmd(shell)
	updatePrompt(shell)
	CreateCnCmd(shell)
	CreateDebugLevelCmd(shell)
	CreateMenuCommands(shell)
	// do not let the application stops by an error case in execution
	systools.AddExitListener("iShell", func(code int) systools.ExitBehavior {
		if code != 0 {
			taskrun.CtxOut("ERROR while execution. reported exit code ", code)
		}
		if forceExit {
			return systools.Continue // do not interrupt the exit if the forceExit is set
		}
		return systools.Interrupt
	})
	shell.Run()
}

// updatePrompt updates the prompt
func updatePrompt(shell *ishell.Shell) {
	dir, err := dirhandle.Current()
	if err != nil {
		panic(err)
	}
	// if the runcommand was not added already
	// (this can be the case, if no comands aviable)
	// we will check again
	if !runCmdAdded {
		runCmdAdded = CreateRunCommands(shell)
	}
	bufferSize := 10
	if width, _, err := systools.GetStdOutTermSize(); err == nil {
		prompt := ">> "
		ctxPromt := "CTX.SHELL "
		if width > 15 {

			sizeForInput := width / 2 // half of the screen should be for the left for the input
			need := sizeForInput + systools.StrLen(configure.GetGlobalConfig().UsedV2Config.CurrentSet) + bufferSize + systools.StrLen(ctxPromt)
			if need <= width { // we have size left, so compose the longer version of the prompt
				sizeLeft := width - (need - 5) // 5 chars buffer
				dirlabel := ""
				if sizeLeft > 5 { // at least something usefull shold be displayed. so lets say at least 5 chars
					dirlabel = systools.StringSubRight(dir, sizeLeft)         // cut the path string if needed
					pathColor := manout.ForeGreen                             // green color by default for the path
					if !configure.GetGlobalConfig().PathMeightPartOfWs(dir) { // check if we are in the workspace
						pathColor = manout.ForeMagenta // if not, then set color to magenta
					}
					dirlabel = pathColor + dirlabel
				}
				prompt = manout.Message(manout.ForeBlue, ctxPromt, dirlabel, manout.ForeCyan, " [", configure.GetGlobalConfig().UsedV2Config.CurrentSet, "] ", manout.ForeLightYellow, ">> ", manout.CleanTag)

			}
		}
		shell.SetPrompt(prompt)
	} else {
		panic(err)
	}
}

// noShellScreen prints info if we do not detect running in an terminal
func noShellScreen() {
	manout.Om.Println("Contxt ", configure.GetVersion(), " build: ", configure.GetBuild())
	manout.Om.Println("no terminal detected. ctx.shell skipped")
}

// headScreen renders the welcome screen
func headScreen(shell *ishell.Shell) {
	manout.Om.Println("welcome to contxt interactive shell")
	manout.Om.Println("   version: ", configure.GetVersion())
	manout.Om.Println("  build-no: ", configure.GetBuild())
	manout.Om.Println(" workspace: ", configure.GetGlobalConfig().UsedV2Config.CurrentSet)
	manout.Om.Println(" ---")
	manout.Om.Println(`
	you entered the interactive shell because you run contxt 
	without any argument.
	`)
	if !inWs() {
		autoRecoverWs()
		manout.Om.Println("... we changed the the workspace path automatically")
		manout.Om.Println("... ")

	}
}

// CreateDefaultComands simple comands they just used as command.
// - lint
// - version
func CreateDefaultComands(shell *ishell.Shell) {

	shell.AddCmd(&ishell.Cmd{
		Name: "lint",
		Help: "checks the current .contxt.yml",
		Func: func(c *ishell.Context) {
			taskrun.LintOut(50, 50, false, false)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "version",
		Help: "print the current version",
		Func: func(c *ishell.Context) {
			c.Println(configure.GetVersion())
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "vars",
		Help: "shows current variables",
		Func: func(c *ishell.Context) {
			taskrun.GetPlaceHoldersFnc(func(phKey, phValue string) {
				manout.Om.Println(manout.Message(manout.ForeCyan, phKey, ":", manout.ForeYellow, phValue))
			})
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "rescan",
		Help: "rescan the workspaces for project and roles updates",
		Func: func(c *ishell.Context) {
			manout.Om.Println("rescaning workspace...")
			taskrun.FindWorkspaceInfoByTemplate(func(ws string, cnt int, update bool, info configure.WorkspaceInfoV2) {
				if update {
					taskrun.CtxOut(ws, " ", manout.ForeDarkGrey, info.Path, manout.ForeGreen, " updated")
				} else {
					taskrun.CtxOut(ws, " ", manout.ForeDarkGrey, info.Path, manout.ForeYellow, " ignored. nothing to do.")
				}
			})
			manout.Om.Println("done")
		},
	})

}

func CreateCnCmd(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "cn",
		Help: "change path in workspace",
		Completer: func(args []string) []string {
			var paths []string = []string{}
			configure.GetGlobalConfig().PathWorkerNoCd(func(i string, s string) {
				paths = append(paths, fmt.Sprintf("%v", i))
			})
			return paths
		},
		Func: func(c *ishell.Context) {
			if len(c.Args) > 0 {
				if path := taskrun.DirFind(c.Args); path != "." {
					os.Chdir(path)
					//configure.GetGlobalConfig().SaveActualPathByPath(path)
					resetShell()
					updatePrompt(shell)
				}
			} else {
				if handleContexNavigation(c) {
					resetShell()
					updatePrompt(shell)
				}
			}
		},
	})
}

func CreateMenuCommands(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name:    "menu",
		Help:    "show the menu",
		Aliases: []string{"ui"},
		Func:    mainMenu,
	})
}

// CreateWsCmd command to switch the workspaces
func CreateWsCmd(shell *ishell.Shell) {

	shell.AddCmd(&ishell.Cmd{
		Name:    "switch",
		Aliases: []string{"ws", "workspace"},
		Help:    "switch workspace for this session",
		Completer: func(args []string) []string {

			return configure.GetGlobalConfig().ListWorkSpaces()
		},
		Func: func(c *ishell.Context) {
			if len(c.Args) > 0 {

				resetShell()
				configure.GetGlobalConfig().ChangeWorkspace(c.Args[0], taskrun.CallBackOldWs, taskrun.CallBackNewWs)
				autoRecoverWs()
				updatePrompt(shell)

			} else {
				if handleWorkSpaces(c) {
					updatePrompt(shell)
				}
			}
		},
	})
}

// CreateRunCommands to display any run target. without an targetname, we will display a pick list
func CreateRunCommands(shell *ishell.Shell) bool {
	if _, found := taskrun.GetAllTargets(); found {

		shell.AddCmd(&ishell.Cmd{
			Name: "run",
			Help: "run one target <target>. press tab for the target ",
			Completer: func(args []string) []string {
				targets, _ := taskrun.GetAllTargets()
				return targets
			},
			Func: func(c *ishell.Context) {
				if len(c.Args) > 0 {
					taskrun.RunTargets(strings.Join(c.Args, " "), true)
				} else {
					handleRunCmds(c)
				}
			},
		})
		return true
	}
	return false
}

func CreateDebugLevelCmd(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name:    "loglevel",
		Aliases: []string{},
		Func: func(c *ishell.Context) {
			if len(c.Args) > 0 {
				loglevel := c.Args[0]
				if loglevel != "" {
					lvl, err := logrus.ParseLevel(loglevel)
					if err != nil {
						taskrun.GetLogger().Fatal(err)
					}
					taskrun.GetLogger().SetLevel(lvl)
				}
			} else {
				c.Println("valid loglevel is required")
			}
		},
		Help:     "set the loglevel to trace, debug, info, warn, error or critical",
		LongHelp: "",
		Completer: func(args []string) []string {
			return []string{"TRACE", "DEBUG", "INFO", "WARN", "CRITICAL"}
		},
	})
}
