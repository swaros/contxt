package outlaw

import (
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/swaros/contxt/configure"
	"github.com/swaros/contxt/taskrun"
	"github.com/swaros/manout"
)

func RunIShell() {
	taskrun.MainInit()
	shell := ishell.New()

	// display welcome info.
	headScreen(shell)
	// register a function for "greet" command.
	shell.AddCmd(&ishell.Cmd{
		Name: "greet",
		Help: "greet user",
		Func: func(c *ishell.Context) {
			c.Println("Hello", strings.Join(c.Args, " "))
		},
	})
	CreateRunCommands(shell)
	CreateDefaultComands(shell)
	updatePrompt(shell)
	shell.Run()
}

func updatePrompt(shell *ishell.Shell) {
	prompt := manout.Message(manout.ForeBlue, "CTX.SHELL", manout.ForeCyan, " [", configure.UsedConfig.CurrentSet, "] ", manout.ForeLightYellow, ">> ", manout.CleanTag)
	shell.SetPrompt(prompt)
}

func headScreen(shell *ishell.Shell) {
	shell.Println("contxt interactive shell ... " + configure.GetBuild())
}

func CreateDefaultComands(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "lint",
		Help: "checks the current .contxt.yml",
		Func: func(c *ishell.Context) {
			taskrun.LintOut(50, 50, false, false)
		},
	})
}

func CreateRunCommands(shell *ishell.Shell) {
	if _, found := taskrun.GetAllTargets(); found {

		shell.AddCmd(&ishell.Cmd{
			Name: "run",
			Help: "run one target ",
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
