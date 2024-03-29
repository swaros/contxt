// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Licensed under the MIT License
//
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

package taskrun

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/manout"
)

var (
	log = &logrus.Logger{
		Out:       os.Stdout,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.ErrorLevel,
	}

	// cobra stuff
	showColors    bool
	loglevel      string
	pathIndex     int
	deleteWs      string
	clearTask     bool
	setWs         string
	runAtAll      bool
	leftLen       int
	rightLen      int
	yamlIndent    int
	showInvTarget bool
	uselastIndex  bool
	showHints     bool
	preVars       map[string]string

	rootCmd = &cobra.Command{
		Use:   "contxt",
		Short: "workspaces for the shell",
		Long: `Contxt helps you to organize projects.
it helps also to execute tasks depending these projects.
this task can be used to setup and cleanup the workspace
if you enter or leave them.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)

		},
	}

	completionCmd = &cobra.Command{
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

	gotoCmd = &cobra.Command{
		Use:   "switch",
		Short: "switch workspace",
		Long: `switch the workspace to a existing ones.
all defined onEnter and onLeave task will be executed
if these task are defined
`,
		Run: func(_ *cobra.Command, args []string) {
			FindWorkspaceInfoByTemplate(nil)
			if len(args) > 0 {
				for _, arg := range args {
					doMagicParamOne(arg)
				}
			}
		},
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			targets := configure.GetGlobalConfig().ListWorkSpaces()
			return targets, cobra.ShellCompDirectiveNoFileComp
		},
	}

	workspaceCmd = &cobra.Command{
		Use:   "workspace",
		Short: "manage workspaces",
		Long: `create a new workspace 'ctx workspace new <name>'. 
Remove a workspace 'ctx workspace rm <name>'.
list all workspaces 'ctx workspace list'.
scan for new projects in the workspace 'ctx workspace scan'`,
	}

	wsNewCmd = &cobra.Command{
		Use:   "new",
		Short: "create a new workspace",
		Long: `
create a new workspace.
this will trigger any onLeave task defined in the workspace
and also onEnter task defined in the new workspace
`,
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			if len(args) > 0 {
				if err := configure.GetGlobalConfig().AddWorkSpace(args[0], CallBackOldWs, CallBackNewWs); err != nil {
					fmt.Println(err)
				} else {
					CtxOut("workspace created ", args[0])
				}

			} else {
				fmt.Println("no workspace name given")
			}
		},
	}

	wsRmCmd = &cobra.Command{
		Use:   "rm",
		Short: "remove a workspace by given name",
		Long: `
remove a workspace.
this will trigger any onLeave task defined in the workspace
and also onEnter task defined in the new workspace
`,
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			if len(args) > 0 {
				if err := configure.GetGlobalConfig().RemoveWorkspace(args[0]); err != nil {
					manout.Error("error while trying to remove workspace", err)
					systools.Exit(systools.ErrorBySystem)
				} else {
					if err := configure.GetGlobalConfig().SaveConfiguration(); err != nil {
						manout.Error("error while trying to save configuration", err)
						systools.Exit(systools.ErrorBySystem)
					}
					CtxOut("workspace removed ", args[0])
				}
			} else {
				fmt.Println("no workspace name given")
			}
		},
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			targets := configure.GetGlobalConfig().ListWorkSpaces()
			return targets, cobra.ShellCompDirectiveNoFileComp
		},
	}

	wsListCmd = &cobra.Command{
		Use:   "list",
		Short: "list all workspaces",
		Long:  "list all workspaces",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			workspacesList := configure.GetGlobalConfig().ListWorkSpaces()
			for _, ws := range workspacesList {
				fmt.Println(ws)
			}
		},
	}

	wsScanCmd = &cobra.Command{
		Use:   "scan",
		Short: "scan for new projects in the workspace",
		Long:  "scan for new projects in the workspace",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			all, updated := FindWorkspaceInfoByTemplate(func(ws string, cnt int, update bool, info configure.WorkspaceInfoV2) {
				if update {
					CtxOut(manout.ForeBlue, ws, " ", manout.ForeDarkGrey, " ", info.Path, manout.ForeGreen, "\tupdated")
				} else {
					CtxOut(manout.ForeBlue, ws, " ", manout.ForeDarkGrey, " ", info.Path, manout.ForeYellow, "\tignored. nothing to do.")
				}
			})
			CtxOut("found ", all, " projects and updated ", updated, " projects")

		},
	}

	wsUseCmd = &cobra.Command{
		Use:   "use",
		Short: "use a workspace",
		Long: `use a workspace. this is then the new active workspace
this will trigger any onLeave task defined in the workspace
and also onEnter task defined in the new workspace
`,
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			if len(args) > 0 {
				if err := configure.GetGlobalConfig().ChangeWorkspace(args[0], CallBackOldWs, CallBackNewWs); err != nil {
					fmt.Println(err)
				} else {
					CtxOut("workspace used ", args[0])
				}
			} else {
				fmt.Println("no workspace name given")
			}
		},
	}
	dirCmd = &cobra.Command{
		Use:   "dir",
		Short: "handle workspaces and assigned paths",
		Long:  "manage workspaces and paths they are assigned",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			checkDirFlags(cmd, args)
			defaulttask := true
			if pathIndex >= 0 {
				pathStr := configure.GetGlobalConfig().GetPathByIndex(strconv.Itoa(pathIndex), ".")
				fmt.Println(pathStr)
				defaulttask = false
			}

			if uselastIndex {
				GetLogger().WithField("dirIndex", configure.GetGlobalConfig().UsedV2Config.CurrentSet).Debug("current stored index")
				pathStr := configure.GetGlobalConfig().GetActivePath(".")
				fmt.Println(pathStr)
				defaulttask = false
			}

			if clearTask {
				GetLogger().Info("got clear command")
				configure.GetGlobalConfig().ClearPaths()
				defaulttask = false
			}

			if deleteWs != "" {
				GetLogger().WithField("workspace", deleteWs).Info("got remove workspace option")
				if err := configure.GetGlobalConfig().RemoveWorkspace(deleteWs); err != nil {
					manout.Error("error while trying to deleting workspace", err)
					systools.Exit(systools.ErrorBySystem)
				}
				if err := configure.GetGlobalConfig().SaveConfiguration(); err != nil {
					manout.Error("error while trying to save configuration", err)
					systools.Exit(systools.ErrorBySystem)
				}
				defaulttask = false
			}

			if setWs != "" {
				GetLogger().WithField("workspace", setWs).Info("create a new worspace")
				configure.GetGlobalConfig().ChangeWorkspace(setWs, CallBackOldWs, CallBackNewWs)
				defaulttask = false
			}

			if defaulttask {
				printInfo()
			}
		},
	}

	showPaths = &cobra.Command{
		Use:   "paths",
		Short: "show assigned paths",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			PrintCnPaths()
		},
	}

	findPath = &cobra.Command{
		Use:   "find",
		Short: "find path by a part of them",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			if len(args) < 1 {
				pathStr := configure.GetGlobalConfig().GetActivePath(".")
				fmt.Println(pathStr)
			} else {
				path, _ := DirFindApplyAndSave(args)
				fmt.Println(path) // path only as output. so cn can handle it
			}
		},
	}

	listPaths = &cobra.Command{
		Use:   "list",
		Short: "show assigned paths",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)

			for _, p := range configure.GetGlobalConfig().ListWorkSpaces() {
				fmt.Println(p)
			}
		},
	}

	addPaths = &cobra.Command{
		Use:   "add",
		Short: "add current path (pwd) to the current workspace",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			dir, err := dirhandle.Current()
			if err == nil {
				fmt.Println(manout.MessageCln("add ", manout.ForeBlue, dir))
				configure.GetGlobalConfig().AddPath(dir)
				configure.GetGlobalConfig().SaveConfiguration()
				FindWorkspaceInfoByTemplate(nil) // this is parsing all templates in all workspaces and updates the project Infos
			}
		},
	}

	removePath = &cobra.Command{
		Use:   "rm",
		Short: "remove current path (pwd) from the current workspace",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			dir, err := dirhandle.Current()
			if err == nil {
				fmt.Println(manout.MessageCln("try to remove ", manout.ForeBlue, dir, manout.CleanTag, " from workspace"))
				removed := configure.GetGlobalConfig().RemovePath(dir)
				if !removed {
					fmt.Println(manout.MessageCln(manout.ForeRed, "error", manout.CleanTag, " path is not part of the current workspace"))
					systools.Exit(1)
				} else {
					fmt.Println(manout.MessageCln(manout.ForeGreen, "success"))
					configure.GetGlobalConfig().SaveConfiguration()
				}
			}
		},
	}

	createCmd = &cobra.Command{
		Use:   "create",
		Short: "create taskfile templates",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			WriteTemplate()
		},
	}

	createImport = &cobra.Command{
		Use:   "import",
		Short: "Create importfile that can be used for templating",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			if len(args) == 0 {
				fmt.Println("No paths submitted")
				systools.Exit(1)
			}
			_, path, exists, terr := GetTemplate()
			if terr != nil {
				fmt.Println(manout.MessageCln(manout.ForeRed, "Error ", manout.CleanTag, terr.Error()))
				systools.Exit(33)
				return
			}
			if exists {
				for _, addPath := range args {
					err := CreateImport(path, addPath)
					if err != nil {
						fmt.Println("Error adding imports:", err)
						systools.Exit(1)
					}
				}
			} else {
				fmt.Println("no taskfile exists. create these first by contxt create")
				systools.Exit(1)
			}

		},
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "prints current version",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			fmt.Println("version", configure.GetVersion(), "build", configure.GetBuild())
		},
	}

	exportCmd = &cobra.Command{
		Use:   "export",
		Short: "exports the script section of an target like a bash script",
		Long: `for extracting tasks commands in a format that can be executed as a shell script.
this will be a plain export without handling dynamic generated placeholders (default placeholders will be parsed)  and contxt macros.
also go-template imports will be handled.
		`,
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			for _, target := range args {
				outStr, err := ExportTask(target)
				if err == nil {
					fmt.Println("# --- -------------- ---------- ----- ------ ")
					fmt.Println("# --- contxt export of target " + target)
					fmt.Println("# --- -------------- ---------- ----- ------ ")
					fmt.Println()
					fmt.Println(HandlePlaceHolder(outStr))
				} else {
					panic(err)
				}

			}
		},
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			//targets, found := targetsAsMap()
			targets, found := GetAllTargets()
			if !found {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return targets, cobra.ShellCompDirectiveNoFileComp
		},
	}

	lintCmd = &cobra.Command{
		Use:   "lint",
		Short: "checking the task file",
		Long: `to check if the task file contains the expected changes.
use --full to see properties they are nor used.
you will also see if a unexpected propertie found `,
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			leftLen, _ := cmd.Flags().GetInt("left")
			rightLen, _ := cmd.Flags().GetInt("right")
			showall, _ := cmd.Flags().GetBool("full")
			yamlParse, _ := cmd.Flags().GetBool("yaml")
			yamlIndent, _ := cmd.Flags().GetInt("indent")
			okay := false
			if yamlParse {
				ShowAsYaml(true, false, yamlIndent)
				okay = LintOut(leftLen, 0, false, true)
			} else {
				okay = LintOut(leftLen, rightLen, showall, false)
			}

			if !okay {
				systools.Exit(1)
			}

		},
	}

	installCmd = &cobra.Command{
		Use:   "install",
		Short: "install shell functions",
		Long: `updates shell related files to get contxt running
		as shortcut ctx. this will allow changing directories depending
		on a context switch.
		`,
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
		},
	}

	installBashRc = &cobra.Command{
		Use:   "bashrc",
		Short: "updates bashrc for using ctx alias",
		Long: `writes needed functions into the users private .bashrc file.
		This includes code completion and the ctx alias.
		`,
		Run: func(_ *cobra.Command, _ []string) {
			BashUser()
		},
	}

	installFish = &cobra.Command{
		Use:   "fish",
		Short: "create fish shell env for ctx",
		Long: `create needed fish functions, auto completion for ctx
		`,
		Run: func(cmd *cobra.Command, _ []string) {
			FishUpdate(cmd)
		},
	}

	installZsh = &cobra.Command{
		Use:   "zsh",
		Short: "create zsh shell env for ctx",
		Long: `create needed zsh functions and auto completion for zsh
		`,
		Run: func(cmd *cobra.Command, _ []string) {
			ZshUpdate(cmd)
		},
	}

	installPwrShell = &cobra.Command{
		Use:   "powershell",
		Short: "create powershell shell functions",
		Long: `create needed powershell functions and auto completion for powershell.
		for powershell the ctx shortcut is not aviable right now.
		`,
		Run: func(cmd *cobra.Command, _ []string) {
			PwrShellUpdate(cmd)
		},
	}

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "run a target in contxt.yml task file",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			checkRunFlags(cmd, args)
			GetLogger().WithField("args", args).Info("Run triggered")
			GetLogger().WithField("all", runAtAll).Info("all workspaces?")

			// set variables by argument
			for preKey, preValue := range preVars {
				GetLogger().WithFields(logrus.Fields{"key": preKey, "val": preValue}).Info("prevalue set by argument")
				SetPH(preKey, preValue)
			}

			if len(args) == 0 {
				printTargets()
			}

			for _, arg := range args {
				GetLogger().WithField("target", arg).Info("try to run target")

				path, err := dirhandle.Current()
				if err == nil {
					if runAtAll {
						configure.GetGlobalConfig().PathWorkerNoCd(func(_ string, path string) {
							GetLogger().WithField("path", path).Info("change dir")
							os.Chdir(path)
							runTargets(path, arg)
						})
					} else {
						runTargets(path, arg)
					}
				}
			}

		},
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			targets, found := GetAllTargets()
			if !found {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return targets, cobra.ShellCompDirectiveNoFileComp
		},
	}
	sharedCmd = &cobra.Command{
		Use:   "shared",
		Short: "manage shared tasks",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
		},
	}

	sharedListCmd = &cobra.Command{
		Use:   "list",
		Short: "list local shared tasks",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			sharedDirs, _ := ListUseCases(false)
			for _, sharedPath := range sharedDirs {
				fmt.Println(sharedPath)
			}
		},
	}

	sharedUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "updates shared uses if possible (git based)",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			useCases, err := ListUseCases(true)
			if err == nil {
				for _, path := range useCases {
					fmt.Println(manout.MessageCln("check usage ", manout.ForeCyan, path))
					UpdateUseCase(path)
				}
			}
		},
	}
)

func checkRunFlags(cmd *cobra.Command, _ []string) {
	runAtAll, _ = cmd.Flags().GetBool("all-paths")
	showInvTarget, _ = cmd.Flags().GetBool("all-targets")
}

func checkDirFlags(cmd *cobra.Command, _ []string) {

	if pindex, err := cmd.Flags().GetString("index"); err == nil {
		configure.GetGlobalConfig().ChangeActivePath(pindex)
	}

	clearTask, _ = cmd.Flags().GetBool("clear")
	deleteWs, _ = cmd.Flags().GetString("delete")
	setWs, _ = cmd.Flags().GetString("workspace")
	uselastIndex, _ = cmd.Flags().GetBool("last")
}

func checkDefaultFlags(cmd *cobra.Command, _ []string) {
	color, err := cmd.Flags().GetBool("coloroff")
	if err == nil && color {
		manout.ColorEnabled = false
	}

	loglevel, _ = cmd.Flags().GetString("loglevel")
	setLoggerByArg()
}

func initCobra() {
	// create dir command
	dirCmd.AddCommand(showPaths)
	dirCmd.AddCommand(addPaths)
	dirCmd.AddCommand(listPaths)
	dirCmd.AddCommand(removePath)
	dirCmd.AddCommand(findPath)

	dirCmd.Flags().IntVarP(&pathIndex, "index", "i", -1, "get path by the index in order the paths are stored")
	dirCmd.Flags().BoolP("clear", "C", false, "remove all path assigments")
	dirCmd.Flags().BoolP("last", "l", false, "get last used path index number")
	dirCmd.Flags().StringP("delete", "d", "", "remove workspace")
	dirCmd.Flags().StringP("workspace", "w", "", "set workspace. if not exists a new workspace will be created")

	runCmd.Flags().BoolP("all-paths", "a", false, "run targets in all paths in the current workspace")
	runCmd.Flags().Bool("all-targets", false, "show all targets. including invisible")
	runCmd.Flags().StringToStringVarP(&preVars, "var", "v", nil, "set variables by keyname and value.")

	createCmd.AddCommand(createImport)

	//rootCmd.PersistentFlags().BoolVarP(&Experimental, "experimental", "E", true, "enable experimental features")
	rootCmd.PersistentFlags().BoolVarP(&showColors, "coloroff", "c", false, "disable usage of colors in output")
	rootCmd.PersistentFlags().BoolVarP(&showHints, "nohints", "n", false, "disable printing hints")
	rootCmd.PersistentFlags().StringVar(&loglevel, "loglevel", "FATAL", "set loglevel")
	rootCmd.AddCommand(dirCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(exportCmd)

	lintCmd.Flags().IntVar(&leftLen, "left", 45, "set the width for the source code")
	lintCmd.Flags().IntVar(&rightLen, "right", 55, "set the witdh for the current state view")
	lintCmd.Flags().IntVar(&yamlIndent, "indent", 2, "set indent for yaml output by using lint --yaml")
	lintCmd.Flags().Bool("full", false, "print also unset properties")
	lintCmd.Flags().Bool("yaml", false, "display parsed taskfile as yaml file")
	lintCmd.Flags().Bool("parse", false, "parse second level keywords (#@...)")

	rootCmd.AddCommand(lintCmd)

	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(gotoCmd)

	installPwrShell.Flags().Bool("create-profile", false, "create a profile for powershell if not exists already")
	installCmd.AddCommand(installBashRc)
	installCmd.AddCommand(installFish)
	installCmd.AddCommand(installZsh)
	installCmd.AddCommand(installPwrShell)
	rootCmd.AddCommand(installCmd)

	workspaceCmd.AddCommand(wsNewCmd, wsRmCmd, wsListCmd, wsScanCmd, wsUseCmd)
	rootCmd.AddCommand(workspaceCmd)

	sharedCmd.AddCommand(sharedListCmd)
	sharedCmd.AddCommand(sharedUpdateCmd)
	rootCmd.AddCommand(sharedCmd)

}

func setLoggerByArg() {
	if loglevel != "" {
		lvl, err := logrus.ParseLevel(loglevel)
		if err != nil {
			log.Fatal(err)
		}
		log.SetLevel(lvl)
	}
}

func initLogger() {
	//log.Out = os.Stdout
	//log.SetLevel(logrus.DebugLevel)

}

func executeCobra() error {
	return rootCmd.Execute()
}

// GetLogger is the main Logger instance
func GetLogger() *logrus.Logger {
	return log
}

func shortcuts() bool {
	if len(os.Args) == 2 {

		switch os.Args[1] {
		case "dir", "run", "create", "version":
			return false
		default:
			foundATask := doMagicParamOne(os.Args[1])
			return foundATask

		}
	}
	return false
}

func InitDefaultVars() {
	SetPH("CTX_OS", configure.GetOs())
	SetPH("CTX_VERSION", configure.GetVersion())
	SetPH("CTX_BUILDNO", configure.GetBuild())
	// on windows we have to deal with old powershell and cmd versions, the do not
	// support ANSII.
	if configure.GetOs() == "windows" {
		manout.ColorEnabled = false         //  by default we turn off the colors.
		if os.Getenv("CTX_COLOR") == "ON" { // then lets see if this should forced for beeing enabled by env-var
			manout.ColorEnabled = true
		} else {
			// if not forced already we try to figure out, by oure own, if the powershell is able to support ANSII
			// this is since version 7 the case
			version := PwrShellExec(PWRSHELL_CMD_VERSION)
			SetPH("CTX_PS_VERSION", version) // also setup varibale to have the PS version in place
			if version >= "7" {
				manout.ColorEnabled = true // enable colors if we have powershell equals or greater then 7
			}
		}
	}
	// we checking the console support
	// and turn the color off again if we do not have an terminal
	inTerminal := "YES"
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		manout.ColorEnabled = false
		inTerminal = "NO"
	}
	SetPH("CTX_IN_TERMINAL", inTerminal) // do we have terminal running?

	if currentDir, err := os.Getwd(); err == nil { // location as var
		SetPH("CTX_PWD", currentDir)
	} else {
		CtxOut("Critical error while reading directory", err)
		systools.Exit(systools.ErrorBySystem)
	}
}

func setWorkspaceVariables() {
	SetPH("CTX_WS", configure.GetGlobalConfig().UsedV2Config.CurrentSet)
	configure.GetGlobalConfig().ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
		for _, ws2 := range cfg.Paths {
			setConfigVaribales(ws2, "WS")
		}

	})
}

// InitWsVariables is setting up variables depending the current found configuration (.contxt.yml)
func InitWsVariables() {
	setWorkspaceVariables()
}

func setConfigVaribales(wsInfo configure.WorkspaceInfoV2, varPrefix string) {
	if wsInfo.Project != "" && wsInfo.Role != "" {
		prefix := wsInfo.Project + "_" + wsInfo.Role
		SetPH(varPrefix+"0_"+prefix, wsInfo.Path) // at least XXX0 without any version. this could be overwritten by other checkouts
		if wsInfo.Version != "" {
			// if version is set, we use them for avoid conflicts with different checkouts
			if versionSan, err := systools.CheckForCleanString(wsInfo.Version); err == nil {
				prefix += "_" + versionSan
				// add it to ws1 as prefix for versionized keys
				SetPH(varPrefix+"1_"+prefix, wsInfo.Path)
			}
		}
	}
}

// FindWorkspaceInfoByTemplate is searching for a template file and if found, it will set the project and role
// from the template file to the workspace info
// this is only done if the workspace info is not set yet
// this is automatically done on each workspace, if the workspace is not set yet
// but only on the command switch and 'dir add'
func FindWorkspaceInfoByTemplate(updateFn func(workspace string, cnt int, update bool, info configure.WorkspaceInfoV2)) (allCount int, updatedCount int) {
	wsCount := 0
	wsUpdated := 0
	if currentPath, err := os.Getwd(); err != nil {
		CtxOut("Error while reading current directory", err)
		systools.Exit(systools.ErrorBySystem)
	} else {
		haveUpdate := false
		configure.GetGlobalConfig().ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
			wsCount++
			for pathIndex, ws2 := range cfg.Paths {
				if err := os.Chdir(ws2.Path); err == nil && ws2.Project == "" && ws2.Role == "" {
					template, _, found, err := GetTemplate()
					if found && err == nil {
						if template.Workspace.Project != "" && template.Workspace.Role != "" {
							ws2.Project = template.Workspace.Project
							ws2.Role = template.Workspace.Role
							cfg.Paths[pathIndex] = ws2
							CtxOut("Found template for workspace ", index, " and set project and role to ", ws2.Project, ":", ws2.Role)
							configure.GetGlobalConfig().UpdateCurrentConfig(cfg)
							haveUpdate = true
							wsUpdated++
							if updateFn != nil {
								updateFn(index, wsCount, true, ws2)
							}
						}
					} else {
						if updateFn != nil {
							updateFn(index, wsCount, false, ws2)
						}
					}
				}
			}

		})
		if haveUpdate {
			configure.GetGlobalConfig().SaveConfiguration()
		}
		os.Chdir(currentPath)
	}
	return wsCount, wsUpdated
}

// MainInit initilaize the Application.
// this is required for any entrie-point
// currently we have two of them.
// by running in interactive in ishell, and by running with parameters.
func MainInit() {
	ResetVariables()                                         // needed because we could run in a shell
	pathIndex = -1                                           // this is the path index used for the current path. -1 means unset
	initLogger()                                             // init the logger. currently there is nothing happens except sometime for local debug
	InitDefaultVars()                                        // init all the default variables first, they are independend from any configuration
	CopyPlaceHolder2Origin()                                 // doing this 1/2 to have the current variables already in palce until we parse the config
	var configErr = configure.GetGlobalConfig().InitConfig() // try to initialize current config
	if configErr != nil {
		log.Fatal(configErr)
	}
	InitWsVariables()        // like InitDefaultVars but these variables needs the configuration initalized
	CopyPlaceHolder2Origin() // make placeholders usable as golang/template again
}

// MainExecute runs main. parsing flags
func MainExecute() {
	MainInit()
	// first handle shortcuts
	// before we get cobra controll
	if !shortcuts() {
		initCobra()
		err := executeCobra()
		if err != nil {
			manout.Error("error", err)
			systools.Exit(systools.ErrorInitApp)
		}

	}

}

func CallBackOldWs(oldws string) bool {
	GetLogger().Info("OLD workspace: ", oldws)
	// get all paths first
	configure.GetGlobalConfig().PathWorkerNoCd(func(_ string, path string) {

		os.Chdir(path)
		template, templateFile, exists, _ := GetTemplate()

		GetLogger().WithFields(logrus.Fields{
			"templateFile": templateFile,
			"exists":       exists,
			"path":         path,
		}).Debug("path parsing")

		if exists && template.Config.Autorun.Onleave != "" {
			onleaveTarget := template.Config.Autorun.Onleave
			GetLogger().WithFields(logrus.Fields{
				"templateFile": templateFile,
				"target":       onleaveTarget,
			}).Info("execute leave-action")
			RunTargets(onleaveTarget, true)

		}

	})
	return true
}

func GetColorEnabled() bool {
	return showColors
}

func SetColorEnabled(enabled bool) {
	showColors = enabled
}

func CallBackNewWs(newWs string) {
	ResetVariables() // reset old variables while change the workspace. (req for shell mode)
	MainInit()       // initialize the workspace
	GetLogger().Info("NEW workspace: ", newWs)
	configure.GetGlobalConfig().PathWorker(func(_ string, path string) { // iterate any path
		template, templateFile, exists, _ := GetTemplate()

		GetLogger().WithFields(logrus.Fields{
			"templateFile": templateFile,
			"exists":       exists,
			"path":         path,
		}).Debug("path parsing")

		// try to run onEnter func at any possible target in the workspace
		if exists && template.Config.Autorun.Onenter != "" {
			onEnterTarget := template.Config.Autorun.Onenter
			GetLogger().WithFields(logrus.Fields{
				"templateFile": templateFile,
				"target":       onEnterTarget,
			}).Info("execute enter-action")
			RunTargets(onEnterTarget, true)
		}

	}, func(origin string) {
		GetLogger().WithFields(logrus.Fields{
			"current-dir": origin,
		}).Debug("done calling autoruns on sub-dirs")
	})
}

func doMagicParamOne(param string) bool {
	result := false
	if param == "show-the-rainbow" {
		systools.TestColorCombinations()
		return true
	}
	// param is a workspace ?
	configure.GetGlobalConfig().ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
		if param == index {
			configure.GetGlobalConfig().ChangeWorkspace(index, CallBackOldWs, CallBackNewWs)
			result = true
		}
	})

	return result
}

func runTargets(_ string, targets string) {
	RunTargets(targets, true)
}

func printOutHeader() {
	fmt.Println(manout.MessageCln(manout.BoldTag, manout.ForeWhite, "cont(e)xt ", manout.CleanTag, configure.GetVersion()))
	fmt.Println(manout.MessageCln(manout.Dim, " build-no [", manout.ResetDim, configure.GetBuild(), manout.Dim, "]"))
	if configure.GetOs() == "windows" {
		fmt.Println(manout.MessageCln(manout.BoldTag, manout.ForeWhite, " powershell version ", manout.CleanTag, GetPH("CTX_PS_VERSION")))
	}
}

func printInfo() {
	printOutHeader()
	printPaths()
}

func ResetVariables() {
	ClearAll()     // clears all placeholders
	ClearAllData() // clears all stored maps
}
