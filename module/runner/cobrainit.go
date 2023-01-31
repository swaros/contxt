package runner

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/systools"
)

type SessionCobra struct {
	RootCmd         *cobra.Command
	ExternalCmdHndl CmdExecutor
	Options         CobraOptions
}

type CobraOptions struct {
	ShowColors bool
	ShowHints  bool
	LogLevel   string
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
			all, updated := c.ExternalCmdHndl.FindWorkspaceInfoByTemplate(func(ws string, cnt int, update bool, info configure.WorkspaceInfoV2) {
				if update {
					ctxout.CtxOut(c.ExternalCmdHndl.GetOuputHandler(), ctxout.ForeBlue, ws, " ", ctxout.ForeDarkGrey, " ", info.Path, ctxout.ForeGreen, "\tupdated")
				} else {
					ctxout.CtxOut(c.ExternalCmdHndl.GetOuputHandler(), ctxout.ForeBlue, ws, " ", ctxout.ForeDarkGrey, " ", info.Path, ctxout.ForeYellow, "\tignored. nothing to do.")
				}
			})
			ctxout.CtxOut(c.ExternalCmdHndl.GetOuputHandler(), "found ", all, " projects and updated ", updated, " projects")

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
	wsCmd.AddCommand(c.GetNewWsCmd(), c.GetRmWsCmd(), c.GetScanCmd())
	return wsCmd
}

// -- Dir Command

func (c *SessionCobra) GetDirCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dir",
		Short: "handle workspaces and assigned paths",
		Long:  "manage workspaces and paths they are assigned",
		Run: func(cmd *cobra.Command, args []string) {
			c.checkDefaultFlags(cmd, args)

			c.ExternalCmdHndl.PrintPaths()

		},
	}
}

// -- Cobra Tools

func (c *SessionCobra) log() *logrus.Logger {
	return c.ExternalCmdHndl.GetLogger()
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