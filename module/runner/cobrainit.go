package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/systools"
)

type SessionCobra struct {
	RootCmd         *cobra.Command
	ExternalCmdHndl CmdExecutor
	Options         CobraOptions
}

type CobraOptions struct {
	ShowColors      bool
	ShowHints       bool
	LogLevel        string
	DirAll          bool // dir flag for show all dirs in any workspace
	ShowFullTargets bool
}

func NewCobraCmds() *SessionCobra {
	return &SessionCobra{
		RootCmd: &cobra.Command{
			Use:   "contxt",
			Short: "organize workspaces in the shell",
			Long: `contxt is a tool to manage your projects.
it setups your shell environment to fast switch between projects
without the need to search for the right directory.
also it includes a task runner to execute commands in the right context.
`,
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Help()
			},
		},
	}
}

func (c *SessionCobra) Init(cmd CmdExecutor) error {
	c.ExternalCmdHndl = cmd
	if c.ExternalCmdHndl == nil {
		return fmt.Errorf("no command executor defined")
	}
	c.RootCmd.PersistentFlags().BoolVarP(&c.Options.ShowColors, "coloroff", "c", false, "disable usage of colors in output")
	c.RootCmd.PersistentFlags().BoolVarP(&c.Options.ShowHints, "nohints", "n", false, "disable printing hints")
	c.RootCmd.PersistentFlags().StringVar(&c.Options.LogLevel, "loglevel", "FATAL", "set loglevel")

	c.RootCmd.AddCommand(c.GetWorkspaceCmd(), c.getCompletion(), c.GetGotoCmd(), c.GetDirCmd())

	return nil
}

// til here we define the commands
// they needs to be added to the root command

// getCompletion returns the completion command for the root command.
func (c *SessionCobra) getCompletion() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

  $ source <(contxt completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ contxt completion bash > /etc/bash_completion.d/contxt
  # macOS:
  $ contxt completion bash > /usr/local/etc/bash_completion.d/contxt

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ contxt completion zsh > "${fpath[1]}/_contxt"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ contxt completion fish | source

  # To load completions for each session, execute once:
  $ contxt completion fish > ~/.config/fish/completions/contxt.fish
PowerShell:

  PS> contxt completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> contxt completion powershell > contxt.ps1
  # and source this file from your PowerShell profile.

  `,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletion(os.Stdout)
			}
		},
	}

}

func (c *SessionCobra) GetGotoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch",
		Short: "switch workspace",
		Long: `switch the workspace to a existing ones.
all defined onEnter and onLeave task will be executed
if these task are defined
`,
		Run: func(_ *cobra.Command, args []string) {
			c.ExternalCmdHndl.FindWorkspaceInfoByTemplate(nil)
			if len(args) > 0 {
				configure.CfgV1.ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
					if args[0] == index {
						configure.CfgV1.ChangeWorkspace(index, c.ExternalCmdHndl.CallBackOldWs, c.ExternalCmdHndl.CallBackNewWs)
					}
				})
			}
		},
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			targets := configure.CfgV1.ListWorkSpaces()
			return targets, cobra.ShellCompDirectiveNoFileComp
		},
	}
}

func (c *SessionCobra) GetPrintWsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "show all workspaces",
		Long:  `show all workspaces and mark the current one`,
		Run: func(cmd *cobra.Command, args []string) {
			c.checkDefaultFlags(cmd, args)
			c.ExternalCmdHndl.PrintWorkspaces()
		},
	}
}

func (c *SessionCobra) GetListWsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "list all workspaces",
		Long:  `list all workspaces`,
		Run: func(cmd *cobra.Command, args []string) {
			c.checkDefaultFlags(cmd, args)
			ws := c.ExternalCmdHndl.GetWorkspaces()
			for _, w := range ws {
				ctxout.CtxOut(c.ExternalCmdHndl.GetOuputHandler(), w)
			}
		},
	}
}

func (c *SessionCobra) GetNewWsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new",
		Short: "create a new workspace",
		Long: `
create a new workspace.
this will trigger any onLeave task defined in the workspace
and also onEnter task defined in the new workspace
`,
		Run: func(cmd *cobra.Command, args []string) {
			//checkDefaultFlags(cmd, args)
			if len(args) > 0 {
				if err := configure.CfgV1.AddWorkSpace(args[0], c.ExternalCmdHndl.CallBackOldWs, c.ExternalCmdHndl.CallBackNewWs); err != nil {
					fmt.Println(err)
				} else {
					ctxout.CtxOut(c.ExternalCmdHndl.GetOuputHandler(), "workspace created ", args[0])
				}

			} else {
				fmt.Println("no workspace name given")
			}
		},
	}
}

func (c *SessionCobra) GetRmWsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm",
		Short: "remove a workspace by given name",
		Long: `
remove a workspace.
this will trigger any onLeave task defined in the workspace
and also onEnter task defined in the new workspace
`,
		Run: func(cmd *cobra.Command, args []string) {
			//checkDefaultFlags(cmd, args)
			if len(args) > 0 {
				if err := configure.CfgV1.RemoveWorkspace(args[0]); err != nil {
					c.log().Error("error while trying to remove workspace", err)
					systools.Exit(systools.ErrorBySystem)
				} else {
					if err := configure.CfgV1.SaveConfiguration(); err != nil {
						c.log().Error("error while trying to save configuration", err)
						systools.Exit(systools.ErrorBySystem)
					}
					ctxout.CtxOut(c.ExternalCmdHndl.GetOuputHandler(), "workspace removed ", args[0])
				}
			} else {
				fmt.Println("no workspace name given")
			}
		},
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			targets := configure.CfgV1.ListWorkSpaces()
			return targets, cobra.ShellCompDirectiveNoFileComp
		},
	}
}

func (c *SessionCobra) GetScanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "scan for new projects in the workspace",
		Long:  "scan for new projects in the workspace",
		Run: func(cmd *cobra.Command, args []string) {
			c.log().Debug("scan for new projects")
			c.checkDefaultFlags(cmd, args)
			ctxout.PrintLn(c.ExternalCmdHndl.GetOuputHandler(), ctxout.BoldTag, "Scanning for new projects", ctxout.CleanTag, " ... ")
			ctxout.PrintLn(c.ExternalCmdHndl.GetOuputHandler(), "checking any known contxt project if there are information to update")
			ctxout.Print(c.ExternalCmdHndl.GetOuputHandler(), "\n")
			ctxout.Print(c.ExternalCmdHndl.GetOuputHandler(), "<table>")
			ctxout.Print(c.ExternalCmdHndl.GetOuputHandler(), "<row>", ctxout.BoldTag, "<tab size='15'> project</tab><tab size='70'>path</tab><tab size='15' origin='2'>status</tab>", ctxout.CleanTag, "</row>")
			all, updated := c.ExternalCmdHndl.FindWorkspaceInfoByTemplate(func(ws string, cnt int, update bool, info configure.WorkspaceInfoV2) {
				color := ctxout.ForeYellow
				msg := "found"
				if !update {
					msg = "nothing new"
					color = ctxout.ForeLightGreen
				}
				ctxout.Print(c.ExternalCmdHndl.GetOuputHandler(), "<row>", ctxout.ForeBlue, "<tab size='15'> ", ws, "</tab>", ctxout.ForeLightBlue, "<tab size='70'>", info.Path, color, "</tab><tab size='15' origin='2'>", msg, "</tab></row>")
			})
			ctxout.PrintLn(c.ExternalCmdHndl.GetOuputHandler(), "</table>")
			ctxout.PrintLn(c.ExternalCmdHndl.GetOuputHandler(), ctxout.CleanTag, "\n")
			ctxout.CtxOut(c.ExternalCmdHndl.GetOuputHandler(), "found ", ctxout.ForeLightBlue, all, ctxout.CleanTag, " projects and updated ", ctxout.ForeLightBlue, updated, ctxout.CleanTag, " projects")

		},
	}
}

func (c *SessionCobra) GetWorkspaceCmd() *cobra.Command {
	wsCmd := &cobra.Command{
		Use:   "workspace",
		Short: "manage workspaces",
		Long: `create a new workspace 'ctx workspace new <name>'. 
Remove a workspace 'ctx workspace rm <name>'.
list all workspaces 'ctx workspace list'.
scan for new projects in the workspace 'ctx workspace scan'`,
	}
	wsCmd.AddCommand(c.GetNewWsCmd(), c.GetRmWsCmd(), c.GetScanCmd(), c.GetPrintWsCmd(), c.GetListWsCmd())
	return wsCmd
}

// -- Dir Command

func (c *SessionCobra) GetDirCmd() *cobra.Command {
	dCmd := &cobra.Command{
		Use:   "dir",
		Short: "handle workspaces and assigned paths",
		Long:  "manage workspaces and paths they are assigned",
		Run: func(cmd *cobra.Command, args []string) {
			c.checkDefaultFlags(cmd, args)
			if len(args) == 0 {
				c.log().Debug("show all paths in any workspace")
				c.log().WithFields(logrus.Fields{"all-flag": c.Options.DirAll}).Debug("show all paths in any workspace shoul be executed")
				current := configure.CfgV1.UsedV2Config.CurrentSet
				ctxout.Print(c.ExternalCmdHndl.GetOuputHandler(), "<table>")
				if c.Options.DirAll {
					configure.CfgV1.ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
						configure.CfgV1.UsedV2Config.CurrentSet = index
						// header for each workspace
						ctxout.Print(c.ExternalCmdHndl.GetOuputHandler(), "<row>", ctxout.BoldTag, "<tab size='100' fill=' '>", index, ctxout.CleanTag, ctxout.ForeDarkGrey, ": index (", cfg.CurrentIndex, ")</tab></row>")
						c.ExternalCmdHndl.PrintPaths(true, c.Options.ShowFullTargets)
						ctxout.Print(c.ExternalCmdHndl.GetOuputHandler(), "<row>", ctxout.ForeDarkGrey, "<tab size='100' fill='─'>─</tab>", ctxout.CleanTag, "</row>")
					})
				} else {
					c.ExternalCmdHndl.PrintPaths(false, c.Options.ShowFullTargets)
				}
				ctxout.PrintLn(c.ExternalCmdHndl.GetOuputHandler(), "</table>")
				configure.CfgV1.UsedV2Config.CurrentSet = current
			}
		},
	}
	dCmd.Flags().BoolVarP(&c.Options.DirAll, "all", "a", false, "show all paths in any workspace")
	dCmd.Flags().BoolVarP(&c.Options.ShowFullTargets, "full", "f", false, "show full amount of targets")
	dCmd.AddCommand(c.GetDirFindCmd(), c.GetDirAddCmd(), c.GetDirRmCmd())
	return dCmd
}

func (c *SessionCobra) GetDirFindCmd() *cobra.Command {
	fCmd := &cobra.Command{
		Use:   "find",
		Short: "find a path in the current workspace",
		Long:  "find a path in the current workspace by the given argument combination",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				pathStr := configure.CfgV1.GetActivePath(".")
				// we use plain output here. so we can use it in the shell and is not affected by the output handler
				fmt.Println(pathStr)
			} else {
				path, _ := c.ExternalCmdHndl.DirFindApplyAndSave(args)
				fmt.Println(path) // path only as output. so cn can handle it. and again plain fmt usage
			}
		},
	}
	return fCmd
}

func (c *SessionCobra) GetDirAddCmd() *cobra.Command {
	aCmd := &cobra.Command{
		Use:   "add",
		Short: "add path(s) to the workspace",
		Long: `add current path (pwd) if no argument is set.
		else add the given paths to the workspace
		like 'ctx dir add /path/to/dir /path/to/other/dir'
		paths need to be absolute paths`,
		Run: func(cmd *cobra.Command, args []string) {
			c.checkDefaultFlags(cmd, args)
			c.print("add path(s) to workspace: ", ctxout.ForeGreen, configure.CfgV1.UsedV2Config.CurrentSet, ctxout.CleanTag)
			if len(args) == 0 {
				dir, err := os.Getwd()
				if err != nil {
					c.log().Error(err)
					return
				}
				args = append(args, dir)
			}
			c.println(" (", ctxout.ForeDarkGrey, len(args), ctxout.CleanTag, " paths)")
			for _, arg := range args {
				c.println("try ... ", ctxout.ForeLightBlue, arg, ctxout.CleanTag)
				// we need to check if the path is absolute
				if !filepath.IsAbs(arg) {
					c.println("error: ", ctxout.ForeRed, "path is not absolute", ctxout.CleanTag)
					return
				}

				if ok, err := dirhandle.Exists(arg); !ok || err != nil {
					if err != nil {
						c.println("error: ", ctxout.ForeRed, err, ctxout.CleanTag)
						return
					}
					c.println("error: ", ctxout.ForeRed, "path does not exist", ctxout.CleanTag)
					return
				}
				if err := configure.CfgV1.AddPath(arg); err == nil {
					c.println("add ", ctxout.ForeBlue, arg, ctxout.CleanTag)
					configure.CfgV1.SaveConfiguration()
					cmd := c.GetScanCmd() // we use the scan command to update the project infos
					cmd.Run(cmd, nil)     // this is parsing all templates in all workspaces and updates the project Infos
				} else {
					c.println("error: ", ctxout.ForeRed, err, ctxout.CleanTag)
				}
			}
		},
	}
	return aCmd
}

func (c *SessionCobra) GetDirRmCmd() *cobra.Command {
	rCmd := &cobra.Command{
		Use:   "rm",
		Short: "remove path(s) from the workspace",
		Long: `remove the given paths from the workspace
		like 'ctx dir rm /path/to/dir /path/to/other/dir'`,
		Run: func(cmd *cobra.Command, args []string) {
			c.checkDefaultFlags(cmd, args)
			c.print("remove path(s) from workspace: ", ctxout.ForeGreen, configure.CfgV1.UsedV2Config.CurrentSet, ctxout.CleanTag)
			if len(args) == 0 {
				dir, err := os.Getwd()
				if err != nil {
					c.log().Error(err)
					return
				}
				args = append(args, dir)
			}
			c.println(" (", ctxout.ForeDarkGrey, len(args), ctxout.CleanTag, " paths)")
			for _, arg := range args {
				c.println("try ... ", ctxout.ForeLightBlue, arg, ctxout.CleanTag)
				// we need to check if the path is absolute
				if !filepath.IsAbs(arg) {
					c.println("error: ", ctxout.ForeRed, "path is not absolute", ctxout.CleanTag)
					return
				}

				// we don not check if the path exists. we just remove it
				// it is possible that the path is not existing anymore
				if ok := configure.CfgV1.RemovePath(arg); ok {
					c.println("remove ", ctxout.ForeBlue, arg, ctxout.CleanTag)
					configure.CfgV1.SaveConfiguration()
					cmd := c.GetScanCmd() // we use the scan command to update the project infos
					cmd.Run(cmd, nil)     // this is parsing all templates in all workspaces and updates the project Infos
				} else {
					c.println("error: ", ctxout.ForeRed, "could not remove path", ctxout.CleanTag)
				}
			}
		},
	}
	return rCmd
}

// log returns the logger
func (c *SessionCobra) log() *logrus.Logger {
	return c.ExternalCmdHndl.GetLogger()
}

// print prints the given message to the output handler
func (c *SessionCobra) print(msg ...interface{}) {
	add := []interface{}{c.ExternalCmdHndl.GetOuputHandler()}
	msg = append(add, msg...)
	ctxout.Print(msg...)
}

// println prints the given message to the output handler with a new line
func (c *SessionCobra) println(msg ...interface{}) {
	add := []interface{}{c.ExternalCmdHndl.GetOuputHandler()}
	msg = append(add, msg...)
	ctxout.PrintLn(msg...)
}

func (c *SessionCobra) checkDefaultFlags(cmd *cobra.Command, _ []string) {
	color, err := cmd.Flags().GetBool("coloroff")
	if err == nil && color {
		behave := ctxout.GetBehavior()
		behave.NoColored = true
		ctxout.SetBehavior(behave)
	}

	c.Options.LogLevel, _ = cmd.Flags().GetString("loglevel")
	c.ExternalCmdHndl.SetLogLevel(c.Options.LogLevel)
}
