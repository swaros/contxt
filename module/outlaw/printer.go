package outlaw

import (
	"fmt"
	"os"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/swaros/contxt/configure"
	"github.com/swaros/contxt/dirhandle"
	"github.com/swaros/contxt/systools"
	"github.com/swaros/contxt/taskrun"
	"github.com/swaros/manout"
)

var runCmdAdded = false

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
	// do not let the application stops by an error case in execution
	systools.AddExitListener("iShell", func(code int) systools.ExitBehavior {
		shell.Println("ERROR while execution. reported exit code ", code)
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
			need := sizeForInput + systools.StrLen(configure.UsedConfig.CurrentSet) + bufferSize + systools.StrLen(ctxPromt)
			if need <= width { // we have size left, so compose the longer version of the prompt
				sizeLeft := width - (need - 5) // 5 chars buffer
				dirlabel := ""
				if sizeLeft > 5 { // at least something usefull shold be displayed. so lets say at least 5 chars
					dirlabel = systools.StringSubRight(dir, sizeLeft) // cut the path string if needed
					pathColor := manout.ForeGreen                     // green color by default for the path
					if !configure.PathMeightPartOfWs(dir) {           // check if we are in the workspace
						pathColor = manout.ForeMagenta // if not, then set color to magenta
					}
					dirlabel = pathColor + dirlabel
				}
				prompt = manout.Message(manout.ForeBlue, ctxPromt, dirlabel, manout.ForeCyan, " [", configure.UsedConfig.CurrentSet, "] ", manout.ForeLightYellow, ">> ", manout.CleanTag)

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
	manout.Om.Println(" workspace: ", configure.UsedConfig.CurrentSet)
	manout.Om.Println(" ---")
	manout.Om.Println(`
	you entered the interactive shell because you run contxt 
	without any argument.
	`)

	if !inWs() {
		autoRecoverWs()
		manout.Om.Println("... we change the the workspace path automatically")
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
		Name: "test",
		Help: "testing ui",
		Func: func(c *ishell.Context) {
			handleWorkSpaces(c)
		},
	})

}

func CreateCnCmd(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "cn",
		Help: "change path in workspace",
		Completer: func(args []string) []string {
			var paths []string = []string{}
			configure.PathWorker(func(i int, s string) {

				paths = append(paths, fmt.Sprintf("%v", i))
			})
			return paths
		},
		Func: func(c *ishell.Context) {
			if len(c.Args) > 0 {
				if path := taskrun.DirFind(c.Args); path != "." {
					os.Chdir(path)
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

// CreateWsCmd command to switch the workspaces
func CreateWsCmd(shell *ishell.Shell) {

	shell.AddCmd(&ishell.Cmd{
		Name:    "switch",
		Aliases: []string{"ws", "workspace"},
		Help:    "switch workspace for this session",
		Completer: func(args []string) []string {
			var ws []string = []string{}
			configure.WorkSpaces(func(s string) {
				ws = append(ws, s)
			})
			return ws
		},
		Func: func(c *ishell.Context) {
			if len(c.Args) > 0 {
				configure.WorkSpaces(func(ws string) {
					if c.Args[0] == ws {
						resetShell()
						configure.ChangeWorkspace(ws, taskrun.CallBackOldWs, taskrun.CallBackNewWs)
						updatePrompt(shell)
					}
				})
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
