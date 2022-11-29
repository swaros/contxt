package outlaw

import (
	"fmt"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/swaros/contxt/configure"
	"github.com/swaros/contxt/dirhandle"
	"github.com/swaros/contxt/taskrun"
	"github.com/swaros/manout"
)

func RunIShell() {
	taskrun.MainInit()
	shell := ishell.New()

	// display welcome info.
	headScreen(shell)
	CreateRunCommands(shell)
	CreateDefaultComands(shell)
	CreateWsCmd(shell)
	updatePrompt(shell)
	shell.Run()
}

// updatePrompt updates the prompt
func updatePrompt(shell *ishell.Shell) {
	dir, err := dirhandle.Current()
	if err != nil {
		panic(err)
	}

	dirPrompt := manout.Message(manout.ForeLightGreen, fmt.Sprintf("%10s", dir))

	if !configure.PathMeightPartOfWs(dir) {
		dirPrompt = manout.Message(manout.ForeLightRed, dir, manout.ForeDarkGrey, " {path is out of context}")
	}
	prompt := manout.Message(manout.ForeBlue, "CTX.SHELL ", dirPrompt, manout.ForeCyan, " [", configure.UsedConfig.CurrentSet, "] ", manout.ForeLightYellow, ">> ", manout.CleanTag)
	shell.SetPrompt(prompt)
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
func CreateDefaultComands(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "lint",
		Help: "checks the current .contxt.yml",
		Func: func(c *ishell.Context) {
			taskrun.LintOut(50, 50, false, false)
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
						configure.ChangeWorkspace(ws, taskrun.CallBackOldWs, taskrun.CallBackNewWs)
						updatePrompt(shell)
					}
				})
			} else {
				manout.Error("missing workspace name")
			}
		},
	})
}

// CreateRunCommands to display any run target. without an targetname, we will display a pick list
func CreateRunCommands(shell *ishell.Shell) {
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
					taskrun.RunTargets(strings.Join(c.Args, ","), true)
				} else {
					if targets, found := taskrun.GetAllTargets(); found {
						choices := c.Checklist(targets,
							"select targets they needs to be run togehter", nil)

						if len(choices) < 1 {
							c.Println("no targets selected")
							c.Println("you have to select the targets by pressing space")
							return
						}
						var selected []string = []string{}
						for _, nr := range choices {
							selected = append(selected, targets[nr])
						}
						runs := strings.Join(selected, ",")
						c.Println("running selected targets: ", runs)
						taskrun.RunTargets(runs, true)
						c.Println("done")
					}
				}
			},
		})

	}
}
