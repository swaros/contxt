package cmdhandle

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/swaros/contxt/context/output"

	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/dirhandle"
)

//var log = logrus.New()
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

  `,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
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
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				for _, arg := range args {
					doMagicParamOne(arg)
				}
			}
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			targets, found := configure.GetWorkSpacesAsList()
			if !found {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return targets, cobra.ShellCompDirectiveNoFileComp
		},
	}

	workspaceCmd = &cobra.Command{
		Use:   "workspace",
		Short: "create new workspace if not exists, and use them",
		Long: `create a new workspace if not exists.
if the workspace is exists, we will just use them.
you need to set the name for the workspace`,
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			workspace, _ := cmd.Flags().GetString("name")
			if workspace == "" {
				output.Error("paramater missing", "name is required")
			} else {
				configure.ChangeWorkspace(workspace, callBackOldWs, callBackNewWs)
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
				dirhandle.PrintDir(pathIndex)
				defaulttask = false
			}

			if uselastIndex {
				GetLogger().WithField("dirIndex", configure.UsedConfig.LastIndex).Debug("current stored index")
				dirhandle.PrintDir(configure.UsedConfig.LastIndex)
				defaulttask = false
			}

			if clearTask {
				GetLogger().Info("got clear command")
				configure.ClearPaths()
				defaulttask = false
			}

			if deleteWs != "" {
				GetLogger().WithField("workspace", deleteWs).Info("got remove workspace option")
				configure.RemoveWorkspace(deleteWs)
				defaulttask = false
			}

			if setWs != "" {
				GetLogger().WithField("workspace", setWs).Info("create a new worspace")
				configure.ChangeWorkspace(setWs, callBackOldWs, callBackNewWs)
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
			fmt.Println(output.MessageCln("\t", "paths stored in ", output.ForeCyan, configure.UsedConfig.CurrentSet))
			dir, err := dirhandle.Current()
			if err == nil {
				count := configure.ShowPaths(dir)
				if count > 0 && !showHints {
					fmt.Println()
					fmt.Println(output.MessageCln("\t", "if you have installed the shell functions ", output.ForeDarkGrey, "(contxt install bash|zsh|fish)", output.CleanTag, " change the directory by ", output.BoldTag, "cn ", count-1))
					fmt.Println(output.MessageCln("\t", "this will be the same as ", output.BoldTag, "cd ", dirhandle.GetDir(count-1)))
				}
			}
		},
	}

	findPath = &cobra.Command{
		Use:   "find",
		Short: "find path by a part of them",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			useIndex := -1
			usePath := "."
			if len(args) == 0 {
				dirhandle.PrintDir(configure.UsedConfig.LastIndex)
			} else {
				configure.PathWorker(func(index int, path string) {
					for _, search := range args {
						found := strings.Contains(path, search)
						if found {
							useIndex = index
							usePath = path
							GetLogger().WithFields(logrus.Fields{"index": useIndex, "path": usePath}).Debug("Found match by comparing strings")
						} else {
							// this part is not found. but maybe it is a index number?
							sIndex, err := strconv.Atoi(search)
							if err == nil && index == sIndex {
								useIndex = index
								usePath = path
								GetLogger().WithFields(logrus.Fields{"index": useIndex, "path": usePath}).Debug("Found match by using param as index")
							}
						}
					}
				})

				if useIndex >= 0 && useIndex != configure.UsedConfig.LastIndex {
					configure.UsedConfig.LastIndex = useIndex
					configure.SaveDefaultConfiguration(true)
				}

				fmt.Println(usePath)
			}

		},
	}

	listPaths = &cobra.Command{
		Use:   "list",
		Short: "show assigned paths",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			configure.DisplayWorkSpaces()
		},
	}

	addPaths = &cobra.Command{
		Use:   "add",
		Short: "add current path (pwd) to the current workspace",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			dir, err := dirhandle.Current()
			if err == nil {
				fmt.Println(output.MessageCln("add ", output.ForeBlue, dir))
				configure.AddPath(dir)
				configure.SaveDefaultConfiguration(true)
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
				fmt.Println(output.MessageCln("try to remove ", output.ForeBlue, dir, output.CleanTag, " from workspace"))
				removed := configure.RemovePath(dir)
				if !removed {
					fmt.Println(output.MessageCln(output.ForeRed, "error", output.CleanTag, " path is not part of the current workspace"))
					os.Exit(1)
				} else {
					fmt.Println(output.MessageCln(output.ForeGreen, "success"))
					configure.SaveDefaultConfiguration(true)
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
				os.Exit(1)
			}
			_, path, exists := GetTemplate()
			if exists {
				for _, addPath := range args {
					err := CreateImport(path, addPath)
					if err != nil {
						fmt.Println("Error adding imports:", err)
						os.Exit(1)
					}
				}
			} else {
				fmt.Println("no taskfile exists. create these first by contxt create")
				os.Exit(1)
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
			if yamlParse {
				ShowAsYaml()
			} else {
				LintOut(leftLen, rightLen, showall)
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
		Run: func(cmd *cobra.Command, args []string) {
			BashUser()
		},
	}

	installFish = &cobra.Command{
		Use:   "fish",
		Short: "create fish shell env for ctx",
		Long: `create needed fish functions, auto completion for ctx
		`,
		Run: func(cmd *cobra.Command, args []string) {
			FishUpdate(cmd)
		},
	}

	installZsh = &cobra.Command{
		Use:   "zsh",
		Short: "create zsh shell env for ctx",
		Long: `create needed zsh functions and auto completion for zsh
		`,
		Run: func(cmd *cobra.Command, args []string) {
			ZshUpdate(cmd)
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

			//if preVars != nil {
			for preKey, preValue := range preVars {
				GetLogger().WithFields(logrus.Fields{"key": preKey, "val": preValue}).Info("prevalue set by argument")
				SetPH(preKey, preValue)
			}
			//}

			if len(args) == 0 {
				printTargets()
			}

			for _, arg := range args {
				GetLogger().WithField("target", arg).Info("try to run target")

				path, err := dirhandle.Current()
				if err == nil {
					if runAtAll {
						configure.PathWorker(func(index int, path string) {
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
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			//targets, found := targetsAsMap()
			targets, found := getAllTargets()
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
					fmt.Println(output.MessageCln("check usage ", output.ForeCyan, path))
					UpdateUseCase(path)
				}
			}
		},
	}
)

func checkRunFlags(cmd *cobra.Command, args []string) {
	runAtAll, _ = cmd.Flags().GetBool("all-paths")
	showInvTarget, _ = cmd.Flags().GetBool("all-targets")
}

func checkDirFlags(cmd *cobra.Command, args []string) {
	pindex, err := cmd.Flags().GetInt("index")
	if err == nil && pindex >= 0 {
		pathIndex = pindex
	}
	GetLogger().WithFields(logrus.Fields{"current": configure.UsedConfig.LastIndex, "index": pindex}).Trace("Index detection")
	if pindex >= 0 && pindex != configure.UsedConfig.LastIndex {
		configure.UsedConfig.LastIndex = pindex
		configure.SaveDefaultConfiguration(true)
	}

	clearTask, _ = cmd.Flags().GetBool("clear")
	deleteWs, _ = cmd.Flags().GetString("delete")
	setWs, _ = cmd.Flags().GetString("workspace")
	uselastIndex, _ = cmd.Flags().GetBool("last")
}

func checkDefaultFlags(cmd *cobra.Command, args []string) {
	color, err := cmd.Flags().GetBool("coloroff")
	if err == nil && color {
		output.ColorEnabled = false
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

	rootCmd.PersistentFlags().BoolVarP(&showColors, "coloroff", "c", false, "disable usage of colors in output")
	rootCmd.PersistentFlags().BoolVarP(&showHints, "nohints", "n", false, "disable printing hints")
	rootCmd.PersistentFlags().StringVar(&loglevel, "loglevel", "FATAL", "set loglevel")
	rootCmd.AddCommand(dirCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(versionCmd)

	lintCmd.Flags().IntVar(&leftLen, "left", 45, "set the width for the source code")
	lintCmd.Flags().IntVar(&rightLen, "right", 55, "set the witdh for the current state view")
	lintCmd.Flags().Bool("full", false, "print also unset properties")
	lintCmd.Flags().Bool("yaml", false, "display parsed taskfile as yaml file")
	lintCmd.Flags().Bool("parse", false, "parse second level keywords (#@...)")

	rootCmd.AddCommand(lintCmd)

	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(gotoCmd)

	installCmd.AddCommand(installBashRc)
	installCmd.AddCommand(installFish)
	installCmd.AddCommand(installZsh)
	rootCmd.AddCommand(installCmd)

	workspaceCmd.Flags().String("name", "", "set the name for the workspace. REQUIRED")
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
	if configure.GetOs() == "windows" {
		output.ColorEnabled = false
	}
}

// MainExecute runs main. parsing flags
func MainExecute() {
	pathIndex = -1
	initLogger()
	InitDefaultVars()
	var configErr = configure.InitConfig()
	if configErr != nil {
		log.Fatal(configErr)
	}

	currentDir, _ := dirhandle.Current()
	SetPH("CTX_PWD", currentDir)

	// first handle shortcuts
	// before we get cobra controll
	if !shortcuts() {
		initCobra()
		executeCobra()
	}

}

func callBackOldWs(oldws string) bool {
	GetLogger().Info("OLD workspace: ", oldws)
	// get all paths first
	configure.PathWorker(func(index int, path string) {

		os.Chdir(path)
		template, templateFile, exists := GetTemplate()

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

func callBackNewWs(newWs string) {
	GetLogger().Info("NEW workspace: ", newWs)
	configure.PathWorker(func(index int, path string) {

		os.Chdir(path)
		template, templateFile, exists := GetTemplate()

		GetLogger().WithFields(logrus.Fields{
			"templateFile": templateFile,
			"exists":       exists,
			"path":         path,
		}).Debug("path parsing")

		if exists && template.Config.Autorun.Onenter != "" {
			onEnterTarget := template.Config.Autorun.Onenter
			GetLogger().WithFields(logrus.Fields{
				"templateFile": templateFile,
				"target":       onEnterTarget,
			}).Info("execute enter-action")
			RunTargets(onEnterTarget, true)
		}

	})
}

func doMagicParamOne(param string) bool {
	result := false
	// param is a workspace ?
	configure.WorkSpaces(func(ws string) {
		if param == ws {
			configure.ChangeWorkspace(ws, callBackOldWs, callBackNewWs)
			result = true
		}
	})

	return result
}

func getAllTargets() ([]string, bool) {
	plainTargets, found := targetsAsMap()
	template, _, exists := GetTemplate()
	if exists {
		shareds := detectSharedTargetsAsMap(template)
		plainTargets = append(plainTargets, shareds...)
	}
	return plainTargets, exists && found
}

func detectSharedTargetsAsMap(current configure.RunConfig) []string {
	var targets []string
	SharedFolderExecuter(current, func(sharedDir, currentDir string) {
		sharedTargets, have := targetsAsMap()
		if have {
			targets = append(targets, sharedTargets...)
		}
	})

	return targets
}

func targetsAsMap() ([]string, bool) {
	var targets []string
	found := false
	template, _, exists := GetTemplate()
	if exists {
		if len(template.Task) > 0 {
			for _, tasks := range template.Task {
				if !tasks.Options.Invisible {
					found = true
					targets = append(targets, tasks.ID)
				}
			}
		}
	}
	return targets, found
}

func printTargets() {

	template, path, exists := GetTemplate()
	if exists {
		fmt.Println(output.MessageCln(output.ForeDarkGrey, "used taskfile:\t", output.CleanTag, path))
		fmt.Println(output.MessageCln(output.ForeDarkGrey, "tasks count:  \t", output.CleanTag, len(template.Task)))
		if len(template.Task) > 0 {
			fmt.Println(output.MessageCln(output.BoldTag, "existing targets:"))
			for _, tasks := range template.Task {
				if showInvTarget || !tasks.Options.Invisible {
					fmt.Println("\t", tasks.ID)
				}
			}
		} else {
			fmt.Println(output.MessageCln("that is what we gor so far:"))
			fmt.Println()
		}

		sharedTargets := detectSharedTargetsAsMap(template)
		if len(sharedTargets) > 0 {

			for _, stasks := range sharedTargets {
				fmt.Println("\t", stasks, output.MessageCln(output.ForeDarkGrey, " shared", output.CleanTag))
			}

		}
	} else {
		fmt.Println(output.MessageCln(output.ForeCyan, "no task-file exists. you can create one by ", output.CleanTag, " contxt create"))
	}
}

func runTargets(path string, targets string) {
	RunTargets(targets, true)
}

func printOutHeader() {
	fmt.Println(output.MessageCln(output.BoldTag, output.ForeWhite, "cont(e)xt ", output.CleanTag, configure.GetVersion()))
	fmt.Println(output.MessageCln(output.Dim, " build-no [", output.ResetDim, configure.GetBuild(), output.Dim, "]"))
}

func printInfo() {
	printOutHeader()
	printPaths()
}

func printPaths() {
	dir, err := dirhandle.Current()
	if err == nil {
		fmt.Println(output.MessageCln(output.ForeWhite, " current directory: ", output.BoldTag, dir))
		fmt.Println(output.MessageCln(output.ForeWhite, " current workspace: ", output.BoldTag, configure.UsedConfig.CurrentSet))
		notWorkspace := true
		pathColor := output.ForeLightBlue
		if !configure.PathMeightPartOfWs(dir) {
			pathColor = output.ForeLightMagenta
		} else {
			notWorkspace = false
		}
		fmt.Println(" contains paths:")
		configure.PathWorker(func(index int, path string) {
			template, _, exists := GetTemplate()
			add := ""
			if strings.Contains(dir, path) {
				add = output.ResetDim + output.ForeCyan
			}
			if dir == path {
				add = output.ResetDim + output.ForeGreen
			}
			if exists {
				outTasks := ""
				for _, tasks := range template.Task {
					if !tasks.Options.Invisible {
						outTasks = outTasks + " " + tasks.ID
					}
				}

				fmt.Println(output.MessageCln("       path: ", output.Dim, " no ", output.ForeYellow, index, " ", pathColor, add, path, output.CleanTag, " targets", "[", output.ForeYellow, outTasks, output.CleanTag, "]"))
			} else {
				fmt.Println(output.MessageCln("       path: ", output.Dim, " no ", output.ForeYellow, index, " ", pathColor, add, path))
			}
		})
		if notWorkspace {
			fmt.Println()
			fmt.Println(output.MessageCln(output.BackYellow, output.ForeBlue, " WARNING ! ", output.CleanTag, "\tyou are currently in none of the assigned locations."))
			fmt.Println("\t\tso maybe you are using the wrong workspace")
		}
		if !showHints {
			fmt.Println()
			fmt.Println(output.MessageCln(" targets can be executes by ", output.BoldTag, "contxt run <targetname>", output.CleanTag, "(for the current directory)"))
			fmt.Println(output.MessageCln(" a target can also be executed in all stored paths by ", output.BoldTag, "contxt run -a <targetname>", output.CleanTag, " independend from current path"))
		}

		fmt.Println()
		if !showHints {
			fmt.Println(output.MessageCln(" all workspaces:", " ... change by ", output.BoldTag, "contxt <workspace>", ""))
		} else {
			fmt.Println(output.MessageCln(" all workspaces:"))
		}
		configure.WorkSpaces(func(name string) {
			if name == configure.UsedConfig.CurrentSet {
				fmt.Println(output.MessageCln("\t[ ", output.BoldTag, name, output.CleanTag, " ]"))
			} else {
				fmt.Println(output.MessageCln("\t  ", name, "   "))
			}
		})
	}
}
