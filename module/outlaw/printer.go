package outlaw

import (
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/swaros/contxt/configure"
	"github.com/swaros/contxt/taskrun"
)

func RunIShell() {
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
	CreateRunAsyncList(shell)
	CreateDefaultComands(shell)
	shell.SetPrompt("[ctx:]>> ")
	shell.Run()
}

func headScreen(shell *ishell.Shell) {
	shell.Println("contxt interactive shell ... " + configure.GetBuild())
}

func CreateDefaultComands(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "lint",
		Help: "checks the current .contxt.yml",
		Func: func(c *ishell.Context) {
			taskrun.ShowAsYaml(true, false, 0)
		},
	})
}

func CreateRunCommands(shell *ishell.Shell) {
	if targets, found := taskrun.GetAllTargets(); found {
		for _, target := range targets {
			shell.AddCmd(&ishell.Cmd{
				Name: "run." + target,
				Help: "run target " + target,
				Func: func(c *ishell.Context) {
					c.Println("start target " + target)
					taskrun.RunTargets(target, true)
				},
			})
		}
	}
}

func CreateRunAsyncList(shell *ishell.Shell) {
	if targets, found := taskrun.GetAllTargets(); found {
		shell.AddCmd(&ishell.Cmd{
			Name: "run",
			Help: "runs multiple targets async",
			Func: func(c *ishell.Context) {

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
			},
		})
	}

}
