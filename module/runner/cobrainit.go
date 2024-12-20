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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/systools"
)

type SessionCobra struct {
	RootCmd         *cobra.Command // the root command
	ExternalCmdHndl CmdExecutor    // the command executor that is used to execute the commands logic
	Options         CobraOptions   // all flags for the root command
}

type CobraOptions struct {
	ShowColors           bool // flag for show colors in output
	DisableTable         bool // flag for disable table output
	RunOnceTimeOutMillis int  // timeout for run once mode
	ShowHints            bool
	LogLevel             string
	DirAll               bool // dir flag for show all dirs in any workspace
	ShowFullTargets      bool
	ShowBuild            bool
	LastDirIndex         bool
	UseContext           string            // flag to switch to a workspace in the context of another workspace
	InContext            bool              // flag to use the current workspace as context
	PreVars              map[string]string // preset variables they will set variables from commandline. they will be overwritten by the template
	ShowVars             bool              // show the variables
	ShowVarsPattern      string            // show the variables with a pattern
	OutputHandler        string            // set the output handler by name
}

// this is the main entry point for the cobra command
func NewCobraCmds() *SessionCobra {
	return &SessionCobra{
		RootCmd: &cobra.Command{
			Use:   configure.GetBinaryName(),
			Short: "organize workspaces in the shell",
			Long: `contxt is a tool to manage your projects.
it setups your shell environment to fast switch between projects
without the need to search for the right directory.
also it includes a task runner to execute commands in the right context.
`,
			Run: func(cmd *cobra.Command, args []string) {
				if len(args) == 0 {
					cmd.Help()
				}
			},
		},
	}
}

// init the cobra command tree
func (c *SessionCobra) Init(cmd CmdExecutor) error {
	c.ExternalCmdHndl = cmd
	if c.ExternalCmdHndl == nil {
		return fmt.Errorf("no command executor defined")
	}
	c.RootCmd.PersistentFlags().BoolVarP(&c.Options.ShowColors, "coloroff", "c", false, "disable usage of colors in output")
	c.RootCmd.PersistentFlags().BoolVarP(&c.Options.ShowHints, "nohints", "n", false, "disable printing hints")
	c.RootCmd.PersistentFlags().StringVar(&c.Options.LogLevel, "loglevel", "FATAL", "set loglevel")
	c.RootCmd.PersistentFlags().BoolVar(&c.Options.DisableTable, "notable", false, "disable table format output")
	c.RootCmd.PersistentFlags().StringVar(&c.Options.UseContext, "usecontext", "", "use a different workspace as context")
	c.RootCmd.PersistentFlags().BoolVarP(&c.Options.InContext, "incontext", "I", false, "use the current workspace as context")
	c.RootCmd.PersistentFlags().StringToStringVarP(&c.Options.PreVars, "var", "v", nil, "set variables by keyname and value.")
	c.RootCmd.PersistentFlags().BoolVarP(&c.Options.ShowVars, "showvars", "s", false, "show all variables")
	c.RootCmd.PersistentFlags().StringVarP(&c.Options.OutputHandler, "output", "o", "table", "set output format handler. (table)")

	c.RootCmd.AddCommand(
		c.GetWorkspaceCmd(),
		c.getCompletion(),
		c.GetGotoCmd(),
		c.GetDirCmd(),
		c.GetInteractiveCmd(),
		c.GetRunCmd(),
		c.GetLintCmd(),
		c.GetInstallCmd(),
		c.GetVersionCmd(),
		c.getSharedCmd(),
		c.GetVariablesCmd(),
		c.GetCreateCmd(),
		c.GetAnkoRunCmd(),
	)
	c.RootCmd.SilenceUsage = true
	return nil
}

func (c *SessionCobra) GetCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "helper for create some config entries or files for contxt",
		Long:  `help command to create template files, imports and other.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c.checkDefaultFlags(cmd, args)

			if len(args) == 0 {
				return c.ExternalCmdHndl.CreateContxtFile()
			}
			return nil
		},
	}
	createCmd.AddCommand(c.GetCreateImportCmd())
	return createCmd
}

func (c *SessionCobra) GetCreateImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "create a new import",
		Long: `create a new import of a file or directory that will be parsed as a value file.
		this will be used to set variables in the template`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c.checkDefaultFlags(cmd, args)
			if len(args) == 0 {
				return errors.New("no path given")
			}
			for _, arg := range args {
				if err := c.ExternalCmdHndl.AddIncludePath(arg); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func (c *SessionCobra) GetAnkoRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "anko",
		Short: "run anko commands",
		Long: `executes the given anko commands from input.
like: ctx anko 'println("hello world")'

use -f <filename> to set a file, that contains the anko commands
like: ctx anko -f my-anko-script.ank`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c.checkDefaultFlags(cmd, args)
			if len(args) == 0 {
				if incFile, err := cmd.Flags().GetString("file"); err != nil {
					return err
				} else {
					if exists, err := systools.Exists(incFile); err != nil {
						return err
					} else {
						if exists {
							if content, err := os.ReadFile(incFile); err != nil {
								return err
							} else {
								args = []string{string(content)}
								if err := c.ExternalCmdHndl.RunAnkoScript(args); err != nil {
									return err
								}
							}
						} else {
							return fmt.Errorf("file %s not found", incFile)
						}
					}
				}
			} else {

				if err := c.ExternalCmdHndl.RunAnkoScript(args); err != nil {
					return err
				}
			}

			return nil
		},
	}
	cmd.Flags().StringP("file", "f", "", "set file path to execute")
	return cmd
}

func (c *SessionCobra) getSharedCmd() *cobra.Command {
	shared := &cobra.Command{
		Use:   "shared",
		Short: "manage shared libraries",
		Long:  "manage shared libraries",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
			}
		},
	}
	shared.AddCommand(
		c.sharedCmdList(),
		c.getSharedUpdateCmd(),
	)
	return shared
}

func (c *SessionCobra) sharedCmdList() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "show all shared libraries",
		Long:  "show all shared libraries",
		Run: func(cmd *cobra.Command, args []string) {
			c.checkDefaultFlags(cmd, args)
			c.ExternalCmdHndl.PrintShared()
		},
	}
}

func (c *SessionCobra) getSharedUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "update shared libraries",
		Long:  "update shared libraries",
		Run: func(cmd *cobra.Command, args []string) {
			c.checkDefaultFlags(cmd, args)
			shared := NewSharedHelper()
			useCases, err := shared.ListUseCases(true)
			if err == nil {
				for _, path := range useCases {
					c.println("check usage ", ctxout.ForeBlue, path, ctxout.CleanTag)
					shared.UpdateUseCase(path)
				}
			}
		},
	}
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
		RunE: func(_ *cobra.Command, args []string) error {
			current := dirhandle.Pushd()
			current.Popd()
			var cmderr error
			found := false
			if len(args) > 0 {
				configure.GetGlobalConfig().ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
					if args[0] == index {
						found = true
						if err := configure.GetGlobalConfig().ChangeWorkspace(index, c.ExternalCmdHndl.CallBackOldWs, c.ExternalCmdHndl.CallBackNewWs); err != nil {
							cmderr = err
						}
					}
				})
			}
			if !found {
				cmderr = fmt.Errorf("workspace %s not found", args[0])
			}
			return cmderr
		},
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			targets := configure.GetGlobalConfig().ListWorkSpaces()
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

// prints the current workspace
func (c *SessionCobra) PrintCurrentWs() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "prints the current workspace",
		Long:  `prints the current active workspace. This is the workspace which is used for the current session`,
		Run: func(cmd *cobra.Command, args []string) {
			c.checkDefaultFlags(cmd, args)
			c.println(c.ExternalCmdHndl.GetCurrentWorkSpace())
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
				out, printer := c.ExternalCmdHndl.GetOuputHandler()
				ctxout.CtxOut(out, printer, w)
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
		RunE: func(cmd *cobra.Command, args []string) error {
			//checkDefaultFlags(cmd, args)
			if len(args) == 1 {
				if err := configure.GetGlobalConfig().AddWorkSpace(args[0], c.ExternalCmdHndl.CallBackOldWs, c.ExternalCmdHndl.CallBackNewWs); err != nil {
					return err
				} else {
					c.print("workspace created ", args[0])
				}

			} else {
				if len(args) == 0 {
					return errors.New("no workspace name given")
				} else {
					return errors.New("to many arguments")
				}
			}
			return nil
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
		RunE: func(cmd *cobra.Command, args []string) error {
			//checkDefaultFlags(cmd, args)
			if len(args) > 0 {
				if err := configure.GetGlobalConfig().RemoveWorkspace(args[0]); err != nil {
					c.log().Error("error while trying to remove workspace", err)
					return err
				} else {
					if err := configure.GetGlobalConfig().SaveConfiguration(); err != nil {
						c.log().Error("error while trying to save configuration", err)
						return err
					}
					c.print("workspace removed ", args[0])
				}
			} else {
				return errors.New("no workspace name given")
			}
			return nil
		},
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			targets := configure.GetGlobalConfig().ListWorkSpaces()
			return targets, cobra.ShellCompDirectiveNoFileComp
		},
	}
}

func (c *SessionCobra) GetScanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "scan all projects in the workspace",
		Long:  "looking for project information, like names and roles, in all the workspaces, and update the global project list",
		Run: func(cmd *cobra.Command, args []string) {
			c.log().Debug("scan for new projects")
			c.checkDefaultFlags(cmd, args)
			c.println(ctxout.BoldTag, "Scanning for new projects", ctxout.CleanTag, " ... ")
			c.println("checking any known contxt project if there are information to update")
			c.print("\n")
			c.print("<table>")
			c.print("<row>", ctxout.BoldTag, "<tab size='15'> project</tab><tab size='70'>path</tab><tab size='15' origin='2'>status</tab>", ctxout.CleanTag, "</row>")
			current := dirhandle.Pushd()
			all, updated := c.ExternalCmdHndl.FindWorkspaceInfoByTemplate(func(ws string, cnt int, update bool, info configure.WorkspaceInfoV2) {
				color := ctxout.ForeYellow
				msg := "found"
				if !update {
					msg = "nothing new"
					color = ctxout.ForeLightGreen
				}
				c.print("<row>", ctxout.ForeBlue, "<tab size='15'> ", ws, "</tab>", ctxout.ForeLightBlue, "<tab size='70'>", info.Path, color, "</tab><tab size='15' origin='2'>", msg, "</tab></row>")
			})
			current.Popd()
			c.print("</table>")
			c.println(ctxout.CleanTag, "")
			c.println("found ", ctxout.ForeLightBlue, all, ctxout.CleanTag, " projects and updated ", ctxout.ForeLightBlue, updated, ctxout.CleanTag, " projects")

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
	wsCmd.AddCommand(
		c.GetNewWsCmd(),
		c.GetRmWsCmd(),
		c.GetScanCmd(),
		c.GetPrintWsCmd(),
		c.GetListWsCmd(),
		c.PrintCurrentWs(),
	)
	return wsCmd
}

// -- lint cmd

func (c *SessionCobra) GetLintCmd() *cobra.Command {
	var showAll bool
	lCmd := &cobra.Command{
		Use:   "lint",
		Short: "lint the .contxt.yaml file",
		Long:  "lint the .contxt.yaml and shows unexpected fields",
		RunE: func(cmd *cobra.Command, args []string) error {
			c.checkDefaultFlags(cmd, args)

			return c.ExternalCmdHndl.Lint(showAll)
		},
	}

	lCmd.Flags().BoolVarP(&showAll, "show-issues", "i", false, "show all issues")
	lCmd.AddCommand(c.GetLintTemplateCmd())

	return lCmd
}

// a lint sub command to display the current loaded template
// as yaml
func (c *SessionCobra) GetLintTemplateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "yaml",
		Short: "show the current loaded template",
		Long:  "show the current loaded template as yaml",
		Run: func(cmd *cobra.Command, args []string) {
			c.checkDefaultFlags(cmd, args)
			c.ExternalCmdHndl.PrintTemplate()
		},
	}
}

// -- Run cmd

func (c *SessionCobra) GetRunAtAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "run all commands in any assigned path",
		Long: `run all commands in any assigned path in the workspace.

		so if you have a workspace that have multiple paths, you can execute the targets on them.
		as example: run all build
		... will run the build target in all assigned paths in the workspace. if this target is defined.
		if the target is not defined in this path, it is just ignored.
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c.checkDefaultFlags(cmd, args)
			if len(args) < 1 {
				return errors.New("no target given")
			}
			c.log().Debug("run all commands in context of project")
			currentDir := dirhandle.Pushd()
			defer currentDir.Popd()
			var runErr error
			for _, runTargetName := range args {
				configure.GetGlobalConfig().PathWorker(func(index, path string) {
					targets := c.ExternalCmdHndl.GetTargets(true)

					c.log().Debug("run command in context of project", args)

					if systools.SliceContains(targets, runTargetName) {
						c.println(ctxout.ForeLightBlue, "running target ", ctxout.ForeLightYellow, runTargetName, ctxout.ResetCode, " in context of path ", ctxout.ForeBlue, path, ctxout.ResetCode)

						if err := c.ExternalCmdHndl.InitExecuter(); err != nil {
							runErr = err
						} else {
							if err := c.ExternalCmdHndl.RunTargets(runTargetName, true); err != nil {
								runErr = err
							}
						}
					} else {
						c.println(ctxout.ForeDarkGrey, "no target ", ctxout.ForeBlue, runTargetName, ctxout.ForeDarkGrey, " in path ", ctxout.ResetCode, path)
					}

				}, func(origin string) {
					// we do not need to do anything here
				})
			}
			return runErr
		},
	}
}

func (c *SessionCobra) GetRunCmd() *cobra.Command {
	rCmd := &cobra.Command{
		Use:   "run",
		Short: "run a command in the context of a project",
		Long:  "run a command in the context of a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			c.checkDefaultFlags(cmd, args)
			if len(args) > 0 {
				c.log().Debug("run command in context of project", args)
				if err := c.ExternalCmdHndl.InitExecuter(); err != nil {
					return err
				}
				for _, p := range args {
					if err := c.ExternalCmdHndl.RunTargets(p, true); err != nil {
						return err
					}
				}
			} else {
				targets := c.ExternalCmdHndl.GetTargets(false)
				for _, p := range targets {
					c.println(p)
				}
				return nil
			}
			return nil
		},
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			targets := c.ExternalCmdHndl.GetTargets(false)
			return targets, cobra.ShellCompDirectiveNoFileComp
		},
	}
	rCmd.AddCommand(c.GetRunAtAllCmd())
	return rCmd
}

// -- Dir Command

func (c *SessionCobra) GetDirCmd() *cobra.Command {
	dCmd := &cobra.Command{
		Use:   "dir",
		Short: "handle workspaces and assigned paths",
		Long:  "manage workspaces and paths they are assigned",
		Run: func(cmd *cobra.Command, args []string) {
			c.checkDefaultFlags(cmd, args)

			// if we do not have any workspace, we do not need to do anything.
			// without an workspace we also have no paths to show
			if len(configure.GetGlobalConfig().ListWorkSpaces()) == 0 {
				c.log().Debug("no workspace found, nothing to do")
				c.println("no workspace found, nothing to do. create a new workspace with 'ctx workspace new <name>'")
				return
			}

			if c.Options.LastDirIndex {
				pathStr := configure.GetGlobalConfig().GetActivePath(".")
				fmt.Println(pathStr)
				c.Options.LastDirIndex = false // one time usage
				return
			}

			if len(args) == 0 {
				c.log().Debug("show all paths in any workspace", c.Options.DirAll)
				current := configure.GetGlobalConfig().UsedV2Config.CurrentSet
				c.print("<table>")
				if c.Options.DirAll {
					configure.GetGlobalConfig().ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
						configure.GetGlobalConfig().UsedV2Config.CurrentSet = index
						// header for each workspace
						c.print("<row>", ctxout.BoldTag, "<tab size='100' fill=' '>", index, ctxout.CleanTag, ctxout.ForeDarkGrey, ": index (", cfg.CurrentIndex, ")</tab></row>")
						c.ExternalCmdHndl.PrintPaths(true, c.Options.ShowFullTargets)
						c.print("<row>", ctxout.ForeDarkGrey, "<tab size='100' fill='─'> </tab>", ctxout.CleanTag, "</row>")

					})
				} else {
					c.ExternalCmdHndl.PrintPaths(false, c.Options.ShowFullTargets)
				}
				c.println("</table>")
				configure.GetGlobalConfig().UsedV2Config.CurrentSet = current
			}
		},
	}
	dCmd.Flags().BoolVarP(&c.Options.DirAll, "all", "a", false, "show all paths in any workspace")
	dCmd.Flags().BoolVarP(&c.Options.ShowFullTargets, "full", "f", false, "show full amount of targets")
	dCmd.Flags().BoolVarP(&c.Options.LastDirIndex, "last", "l", false, "show last used path")
	dCmd.AddCommand(c.GetDirFindCmd(), c.GetDirAddCmd(), c.GetDirRmCmd(), c.GetDirLsCmd())
	return dCmd
}

func (c *SessionCobra) GetDirFindCmd() *cobra.Command {
	fCmd := &cobra.Command{
		Use:   "find",
		Short: "find a path in the current workspace",
		Long:  "find a path in the current workspace by the given argument combination",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				pathStr := configure.GetGlobalConfig().GetActivePath(".")
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

// GetDirLsCmd returns the command to list all paths in the current workspace
func (c *SessionCobra) GetDirLsCmd() *cobra.Command {
	fCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all paths in the current workspace",
		Long:    "list all paths in the current workspace in a simple list",
		Run: func(cmd *cobra.Command, args []string) {
			paths := configure.GetGlobalConfig().ListPaths()
			for _, path := range paths {
				c.println(path)
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
		RunE: func(cmd *cobra.Command, args []string) error {
			c.checkDefaultFlags(cmd, args)
			c.print("add path(s) to workspace: ", ctxout.ForeGreen, configure.GetGlobalConfig().UsedV2Config.CurrentSet, ctxout.CleanTag)
			if len(args) == 0 {
				dir, err := os.Getwd()
				if err != nil {
					c.log().Error(err)
					return err
				}
				args = append(args, dir)
			}
			c.println(" (", ctxout.ForeDarkGrey, len(args), ctxout.CleanTag, " paths)")
			for _, arg := range args {
				if arg == "" {
					return errors.New("empty path is not allowed")
				}
				c.println("try ... ", ctxout.ForeLightBlue, arg, ctxout.CleanTag)
				// we need to check if the path is absolute
				if !filepath.IsAbs(arg) {
					c.println("error: ", ctxout.ForeRed, "path is not absolute", ctxout.CleanTag)
					currentDir, _ := filepath.Abs(".")
					return errors.New("given path for adding is not absolute [" + arg + "], current path:" + currentDir)
				}

				if ok, err := dirhandle.Exists(arg); !ok || err != nil {
					if err != nil {
						c.println("error: ", ctxout.ForeRed, err, ctxout.CleanTag)
						return err
					}
					c.println("error: ", ctxout.ForeRed, "path does not exist", ctxout.CleanTag)
					return errors.New("path does not exist")
				}
				if err := configure.GetGlobalConfig().AddPath(arg); err == nil {
					c.println("add ", ctxout.ForeBlue, arg, ctxout.CleanTag)
					configure.GetGlobalConfig().SaveConfiguration()
					cmd := c.GetScanCmd() // we use the scan command to update the project infos
					cmd.Run(cmd, nil)     // this is parsing all templates in all workspaces and updates the project Infos
				} else {
					c.println("error: ", ctxout.ForeRed, err, ctxout.CleanTag)
					return err
				}
			}
			return nil
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
		RunE: func(cmd *cobra.Command, args []string) error {
			c.checkDefaultFlags(cmd, args)
			c.print("remove path(s) from workspace: ", ctxout.ForeGreen, configure.GetGlobalConfig().UsedV2Config.CurrentSet, ctxout.CleanTag)
			if len(args) == 0 {
				dir, err := os.Getwd()
				if err != nil {
					c.log().Error(err)
					return err
				}
				args = append(args, dir)
			}
			c.println(" (", ctxout.ForeDarkGrey, len(args), ctxout.CleanTag, " paths)")
			for _, arg := range args {
				c.println("try ... ", ctxout.ForeLightBlue, arg, ctxout.CleanTag)
				// we need to check if the path is absolute
				if !filepath.IsAbs(arg) {
					c.println("error: ", ctxout.ForeRed, "path is not absolute", ctxout.CleanTag)
					return errors.New("path is not absolute for dir cmd")
				}

				// we don not check if the path exists. we just remove it
				// it is possible that the path is not existing anymore
				if ok := configure.GetGlobalConfig().RemovePath(arg); ok {
					c.println("remove ", ctxout.ForeBlue, arg, ctxout.CleanTag)
					configure.GetGlobalConfig().SaveConfiguration()
					cmd := c.GetScanCmd() // we use the scan command to update the project infos
					cmd.Run(cmd, nil)     // this is parsing all templates in all workspaces and updates the project Infos
				} else {
					c.println("error: ", ctxout.ForeRed, "could not remove path", ctxout.CleanTag)
					return errors.New("could not remove path")
				}
			}
			return nil
		},
	}
	return rCmd
}

func (c *SessionCobra) GetInstallCmd() *cobra.Command {
	iCmd := &cobra.Command{
		Use:   "install",
		Short: "install shell support",
		Long: `install shell support for different shells.
		supported shells are: bash, zsh, fish, powershell`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c.checkDefaultFlags(cmd, args)
			c.print("install shell support")
			return nil
		},
	}
	iCmd.AddCommand(c.GetInstallBashCmd())
	iCmd.AddCommand(c.GetInstallZshCmd())
	iCmd.AddCommand(c.GetInstallFishCmd())
	iCmd.AddCommand(c.GetInstallPowershellCmd())
	return iCmd
}

func (c *SessionCobra) GetInstallBashCmd() *cobra.Command {
	iCmd := &cobra.Command{
		Use:   "bash",
		Short: "install shell support for bash",
		Long: `install shell support for bash.
		this is done by adding some functions and a source command to the bashrc file.
		Afterwards you can use the ctx command in your bash shell instead of contxt.
		So after an switch command, you will automatically change the dir to the new workspace.
		You can also use the cn command to change one of the assigned paths to the current workspace.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			c.ExternalCmdHndl.MainInit()
			installer := NewShellInstall(c.log())
			if err := installer.BashUserInstall(); err != nil {
				c.log().Error(err)
				return err
			}
			return nil
		},
	}
	return iCmd
}

func (c *SessionCobra) GetInstallZshBaseFunc() *cobra.Command {
	iCmd := &cobra.Command{
		Use:   "base",
		Short: "get the zsh base functions for completion",
		Long: `get the zsh base functions for completion.
		this is sometimes helpfull to get completion in zsh running, because of some
		restrictions in zsh.
		`,

		Run: func(cmd *cobra.Command, args []string) {
			c.ExternalCmdHndl.MainInit()
			// check the shortcut flag
			shortcut, _ := cmd.Flags().GetBool("shortcut")
			installer := NewShellInstall(c.log())
			fmt.Println(installer.GetBinFunc(c.RootCmd, shortcut))

		},
	}
	iCmd.Flags().BoolP("shortcut", "s", false, "show the shortcut functions")
	return iCmd
}

func (c *SessionCobra) GetInstallZshShellScript() *cobra.Command {
	iCmd := &cobra.Command{
		Use:   "compscript",
		Short: "creates a script for zsh completion",
		Long: `creates a script for zsh completion.
		this script try to create the completions for zsh locally.
		and uses then sudo to copy the file to the zsh completion dir.
		`,

		Run: func(cmd *cobra.Command, args []string) {
			c.ExternalCmdHndl.MainInit()

			installer := NewShellInstall(c.log())
			fmt.Println(installer.GetScript(c.RootCmd))

		},
	}
	return iCmd
}

func (c *SessionCobra) GetInstallZshCmd() *cobra.Command {
	iCmd := &cobra.Command{
		Use:   "zsh",
		Short: "install shell support for zsh",
		Long: `install shell support for zsh.
		this is done by adding some functions and a source command to the zshrc file.
		Afterwards you can use the ctx command in your zsh shell instead of contxt.
		So after an switch command, you will automatically change the dir to the new workspace.
		You can also use the cn command to change one of the assigned paths to the current workspace.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			c.ExternalCmdHndl.MainInit()
			installer := NewShellInstall(c.log())
			if err := installer.ZshUpdate(c.RootCmd); err != nil {
				return err
			}
			return nil
		},
	}
	iCmd.AddCommand(c.GetInstallZshBaseFunc())
	iCmd.AddCommand(c.GetInstallZshShellScript())
	return iCmd
}

func (c *SessionCobra) GetInstallFishCmd() *cobra.Command {
	iCmd := &cobra.Command{
		Use:   "fish",
		Short: "install shell support for fish",
		Long: `install shell support for fish.
		this is done by adding some functions and a source command to the fish config file.
		Afterwards you can use the ctx command in your fish shell instead of contxt.
		So after an switch command, you will automatically change the dir to the new workspace.
		You can also use the cn command to change one of the assigned paths to the current workspace.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			c.ExternalCmdHndl.MainInit()
			installer := NewShellInstall(c.log())
			if err := installer.FishUpdate(c.RootCmd); err != nil {
				c.log().Error(err)
				return err
			}
			return nil
		},
	}
	return iCmd
}

func (c *SessionCobra) GetInstallPowershellCmd() *cobra.Command {
	iCmd := &cobra.Command{
		Use:   "powershell",
		Short: "install shell support for powershell",
		Long: `install shell support for powershell.
		this is done by adding some functions and a source command to the powershell profile file.
		Afterwards you can use the ctx command in your powershell shell instead of contxt.		
		You can also use the cn command to change one of the assigned paths to the current workspace.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c.ExternalCmdHndl.MainInit()
			installer := NewShellInstall(c.log())
			if err := installer.PwrShellUpdate(c.RootCmd); err != nil {
				c.log().Error(err)
				return err
			}
			return nil
		},
	}
	return iCmd
}

func (c *SessionCobra) GetVariablesCmd() *cobra.Command {
	vCmd := &cobra.Command{
		Use:   "vars",
		Short: "shows variables",
		Long:  `list all defined variables that are used in the current workspace.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c.checkDefaultFlags(cmd, args)
			c.ExternalCmdHndl.PrintVariables(c.Options.ShowVarsPattern)
			return nil
		},
	}
	vCmd.Flags().StringVarP(&c.Options.ShowVarsPattern, "format", "f", "", "format string for the output. use [nl] for new line and 2x %s for the key and the value like  --format=\"%s=%s[nl]\"")
	return vCmd
}

func (c *SessionCobra) GetVersionCmd() *cobra.Command {
	vCmd := &cobra.Command{
		Use:   "version",
		Short: "print the version number of contxt",
		Long:  `All software has versions. This is contxt's`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c.checkDefaultFlags(cmd, args)
			if c.Options.ShowBuild {
				c.println("contxt version: ", configure.GetVersion(), " build ", configure.GetBuild())
			} else {
				c.println("contxt version: ", configure.GetVersion())
			}
			return nil
		},
	}
	vCmd.Flags().BoolVarP(&c.Options.ShowBuild, "build", "b", false, "show build information")
	return vCmd
}

// log returns the logger
func (c *SessionCobra) log() mimiclog.Logger {
	return c.ExternalCmdHndl.GetLogger()
}

// print prints the given message to the output handler
func (c *SessionCobra) print(msg ...interface{}) {
	c.ExternalCmdHndl.Print(msg...)
}

// println prints the given message to the output handler with a new line
func (c *SessionCobra) println(msg ...interface{}) {
	c.ExternalCmdHndl.Println(msg...)
}

func (c *SessionCobra) checkDefaultFlags(cmd *cobra.Command, _ []string) {
	envColorOff := os.Getenv("CTX_COLOROFF")
	// TODO: why this way? is the global flag not working?
	color, err := cmd.Flags().GetBool("coloroff")
	if err == nil && (color || envColorOff == "true") {
		behave := ctxout.GetBehavior()
		behave.NoColored = true
		ctxout.SetBehavior(behave)
	} else if err != nil {
		c.log().Error(err)
		systools.Exit(systools.ErrorInitApp)
	}

	c.Options.LogLevel, _ = cmd.Flags().GetString("loglevel")
	if err := c.ExternalCmdHndl.SetLogLevel(c.Options.LogLevel); err != nil {
		c.log().Error(err)
		systools.Exit(systools.ErrorInitApp)
	}

	// set the output handler by env or flag
	envOutput := os.Getenv("CTX_OUTPUT")
	if envOutput != "" {
		c.Options.OutputHandler = envOutput
	}
	if c.Options.OutputHandler != "" {
		if err := c.ExternalCmdHndl.SetOutputHandlerByName(c.Options.OutputHandler); err != nil {
			ctxout.PrintLn("error while setting output handler:", err)
			systools.Exit(systools.ErrorInitApp)
		}
	}

	// force the log level by env var
	envLogLevel := os.Getenv("CTX_LOGLEVEL")
	if envLogLevel != "" {
		c.Options.LogLevel = envLogLevel
	}

	// force the disable table flag by env var
	envDisableTable := os.Getenv("CTX_DISABLE_TABLE")
	if envDisableTable == "true" {
		c.Options.DisableTable = true
	}
	if c.Options.DisableTable {
		// overwrite the table plugin with a disabled one
		ctxout.UpdateFilterByRef(ctxout.NewTabOut(), ctxout.PostFilterInfo{Disabled: true})
	}

	// preset all variables from command line
	// set variables by argument.
	for preKey, preValue := range c.Options.PreVars {
		c.log().Debug("preset Value", preKey, preValue)
		c.ExternalCmdHndl.SetPreValue(preKey, preValue)
	}

	// forces to change to the current used workspace and the assosiated path
	if c.Options.InContext {
		if c.Options.UseContext != "" {
			c.log().Critical("use of --incontext and --usecontext is not allowed")
			systools.Exit(systools.ErrorInitApp)
		}
		c.log().Info("force to change to the current workspace and path", configure.GetGlobalConfig().CurrentWorkspace())
		path := configure.GetGlobalConfig().GetActivePath("")
		if path != "" {
			c.log().Info("change to path", path)
			os.Chdir(path)
			// important to set the default variables in the context of the used workspace
			c.ExternalCmdHndl.MainInit()
		}
	}

	// if this flag is set, we do change into the workspace and using the current directory there.
	// so any command will be executed in the context of the workspace
	if c.Options.UseContext != "" {
		c.log().Info("working in context of ", c.Options.UseContext)
		configure.GetGlobalConfig().ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
			if c.Options.UseContext == index {

				if err := configure.GetGlobalConfig().ChangeWorkspaceNotSaved(index); err != nil {
					c.log().Error("error while trying to change workspace", err)
					systools.Exit(systools.ErrorWhileLoadCfg)
				}

				path := configure.GetGlobalConfig().GetActivePath("")
				if path != "" {
					c.println("going to path: ", ctxout.ForeLightBlue, path, ctxout.CleanTag)
					os.Chdir(path)
					// important to set the default variables in the context of the used workspace
					c.ExternalCmdHndl.MainInit()
				}
			}
		})

	}
}

func (c *SessionCobra) GetInteractiveCmd() *cobra.Command {
	iCmd := &cobra.Command{
		Use:   "interactive",
		Short: "start the interactive mode",
		Long:  `start the interactive mode`,
		Run: func(cmd *cobra.Command, args []string) {
			c.checkDefaultFlags(cmd, args)
			c.println("start interactive mode")
			c.ExternalCmdHndl.InteractiveScreen()
		},
	}
	iCmd.AddCommand(c.GetInteractiveOnceCmd())
	return iCmd
}

func (c *SessionCobra) GetInteractiveOnceCmd() *cobra.Command {
	iCmd := &cobra.Command{
		Use:   "once",
		Short: "run a command in the context ot the contxt shell",
		Long: `run a command in the contxt shell. so you have access to all variables and functions 
		they are only available in the contxt shell`,
		Run: func(cmd *cobra.Command, args []string) {
			c.checkDefaultFlags(cmd, args)
			c.println("...using shell to execute commands:")
			for _, arg := range args {
				c.println("... ", ctxout.ForeLightBlue, arg, ctxout.CleanTag)
			}
			c.println("... timeout is ", ctxout.ForeLightBlue, c.Options.RunOnceTimeOutMillis, ctxout.CleanTag, " milliseconds")
			c.ExternalCmdHndl.ShellWithComands(args, c.Options.RunOnceTimeOutMillis)
		},
	}
	iCmd.PersistentFlags().IntVarP(&c.Options.RunOnceTimeOutMillis, "timeout", "t", 10000, "timeout for the interactive once commands in milliseconds")
	return iCmd
}
