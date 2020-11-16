package cmdhandle

import (
	"fmt"
	"os"
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
	showColors bool
	loglevel   string
	pathIndex  int
	deleteWs   string
	clearTask  bool
	setWs      string
	runAtAll   bool

	rootCmd = &cobra.Command{
		Use:   "contxt",
		Short: "worspaces for the shell",
		Long: `Contxt helps you to organize projects.
it helps also to execute tasks depending these projects.
this task can be used to setup and cleanup the workspace 
if you enter or leave them.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)

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
				if count > 0 {
					fmt.Println()
					fmt.Println(output.MessageCln("\t", "to change directory depending stored path you can write ", output.BoldTag, "cd $(", os.Args[0], " -i ", count-1, ")", output.CleanTag, " in bash"))
					fmt.Println(output.MessageCln("\t", "this will be the same as ", output.BoldTag, "cd ", dirhandle.GetDir(count-1)))
				}
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

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "prints current version",
		Run: func(cmd *cobra.Command, args []string) {
			checkDefaultFlags(cmd, args)
			fmt.Println("version", configure.GetVersion(), "build", configure.GetBuild())
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

			if len(args) == 0 {
				printTargets()
			}

			for _, arg := range args {
				GetLogger().WithField("target", arg).Info("try to run target")

				path, err := dirhandle.Current()
				if err == nil {
					if runAtAll {
						configure.PathWorker(func(index int, path string) {
							os.Chdir(path)
							runTargets(path, arg)
						})
					} else {
						runTargets(path, arg)
					}
				}
			}

		},
	}
)

func checkRunFlags(cmd *cobra.Command, args []string) {
	runAtAll, _ = cmd.Flags().GetBool("all-workspaces")
}

func checkDirFlags(cmd *cobra.Command, args []string) {
	pindex, err := cmd.Flags().GetInt("index")
	if err == nil && pindex >= 0 {
		pathIndex = pindex
	}

	clearTask, _ = cmd.Flags().GetBool("clear")
	deleteWs, _ = cmd.Flags().GetString("delete")
	setWs, _ = cmd.Flags().GetString("workspace")

}

func checkDefaultFlags(cmd *cobra.Command, args []string) {
	color, err := cmd.Flags().GetBool("coloroff")
	if err == nil && color == true {
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

	dirCmd.Flags().IntVarP(&pathIndex, "index", "i", -1, "get path by the index in order the paths are stored")
	dirCmd.Flags().BoolP("clear", "C", false, "remove all path assigments")
	dirCmd.Flags().StringP("delete", "d", "", "remove workspace")
	dirCmd.Flags().StringP("workspace", "w", "", "set workspace. if not exists a new workspace will be created")

	runCmd.Flags().BoolP("all-workspaces", "a", false, "run targets in all workspaces")

	rootCmd.PersistentFlags().BoolVarP(&showColors, "coloroff", "c", false, "disable usage of colors in output")
	rootCmd.PersistentFlags().StringVar(&loglevel, "loglevel", "FATAL", "set loglevel")
	rootCmd.AddCommand(dirCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(versionCmd)

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

// MainExecute runs main. parsing flags
func MainExecute() {
	pathIndex = -1
	initLogger()

	var configErr = configure.InitConfig()
	if configErr != nil {
		log.Fatal(configErr)
	}

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
			RunTargets(onleaveTarget)

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
			RunTargets(onEnterTarget)
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

func printTargets() {

	template, path, exists := GetTemplate()
	if exists {
		fmt.Println(output.MessageCln(output.ForeDarkGrey, "used taskfile:\t", output.CleanTag, path))
		fmt.Println(output.MessageCln(output.ForeDarkGrey, "tasks count:  \t", output.CleanTag, len(template.Task)))
		if len(template.Task) > 0 {
			fmt.Println(output.MessageCln(output.BoldTag, "existing targets:"))
			for _, tasks := range template.Task {
				fmt.Println("\t", tasks.ID)
			}
		} else {
			fmt.Println(output.MessageCln("that is what we gor so far:"))
			fmt.Println()
			LintOut(template)
		}
	} else {
		fmt.Println(output.MessageCln(output.ForeCyan, "no task-file exists. you can create one by ", output.CleanTag, " contxt create"))
	}
}

func runTargets(path string, targets string) {
	RunTargets(targets)
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
					outTasks = outTasks + " " + tasks.ID
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

		fmt.Println()
		fmt.Println(output.MessageCln(" targets can be executes by ", output.BoldTag, "contxt run <targetname>", output.CleanTag, "(for the current directory)"))
		fmt.Println(output.MessageCln(" a target can also be executed in all stored paths by ", output.BoldTag, "contxt run -a <targetname>", output.CleanTag, " independend from current path"))

		fmt.Println()
		fmt.Println(output.MessageCln(" all workspaces:", " ... change by ", output.BoldTag, "contxt <workspace>", ""))
		configure.WorkSpaces(func(name string) {
			if name == configure.UsedConfig.CurrentSet {
				fmt.Println(output.MessageCln("\t[ ", output.BoldTag, name, output.CleanTag, " ]"))
			} else {
				fmt.Println(output.MessageCln("\t  ", name, "   "))
			}
		})
	}
}
