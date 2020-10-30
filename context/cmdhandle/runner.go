package cmdhandle

import (
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/context/output"

	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/dirhandle"
)

//var log = logrus.New()
var log = &logrus.Logger{
	Out:       os.Stdout,
	Formatter: new(logrus.TextFormatter),
	Hooks:     make(logrus.LevelHooks),
	Level:     logrus.ErrorLevel,
}

func initLogger() {
	//log.Out = os.Stdout
	//log.SetLevel(logrus.DebugLevel)

}

// GetLogger is the main Logger instance
func GetLogger() *logrus.Logger {
	return log
}

// MainExecute runs main. parsing flags
func MainExecute() {

	initLogger()

	var configErr = configure.InitConfig()
	if configErr != nil {
		log.Fatal(configErr)
	}
	nonParams := true

	// Directory related commands
	dirCommand := flag.NewFlagSet("dir", flag.ExitOnError)

	addCmd := dirCommand.Bool("add", false, "register current directory in current workspace")
	pathIndex := dirCommand.Int("i", -1, "get path by index. use -paths to see index and assigned paths")
	showPaths := dirCommand.Bool("paths", false, "show current paths")
	showWorkspaces := dirCommand.Bool("list", false, "display all existing workspaces")
	workSpace := dirCommand.String("w", "", "set current workspace")
	removeWorkSpace := dirCommand.String("delete", "", "remove workspace")
	clearPaths := dirCommand.Bool("clear", false, "remove all paths from workspace")
	info := dirCommand.Bool("info", false, "show current workspace")

	// script execution releated commands
	scriptCommand := flag.NewFlagSet("run", flag.ExitOnError)

	targets := scriptCommand.String("target", "", "set target. for mutliple targets seperate by ,")
	allDirs := scriptCommand.Bool("all-paths", false, "run targets in all paths")
	execute := scriptCommand.String("exec", "", "Execute a command on all paths")
	execTemplate := scriptCommand.Bool("create-template", false, "write template for path dependeing executions in current folder")

	if len(os.Args) < 2 {
		printOutHeader()
		fmt.Println("not enough arguments")
		fmt.Println(output.MessageCln(output.ForeWhite, "  dir", output.ForeLightCyan, "\tmanaging paths in workspaces"))
		fmt.Println(output.MessageCln(output.ForeWhite, "  run", output.ForeLightCyan, "\texecuting scripts in workspaces"))
		os.Exit(1)
	}

	switch os.Args[1] {
	case "dir":
		dirCommand.Parse(os.Args[2:])
	case "run":
		scriptCommand.Parse(os.Args[2:])
	default:
		foundATask := doMagicParamOne(os.Args[1])
		if !foundATask {
			fmt.Println(output.MessageCln("unexpected command ", output.ForeLightCyan, os.Args[1]))
			flag.PrintDefaults()
			os.Exit(1)
		}

	}

	if scriptCommand.Parsed() {
		someRunCmd := false
		// run against a list of targets just in current dir
		if *targets != "" && *allDirs == false {
			nonParams = false
			someRunCmd = true
			path, _ := dirhandle.Current()
			runTargets(path, *targets)
		}

		// run targets in all paths
		if *targets != "" && *allDirs == true {
			nonParams = false
			someRunCmd = true
			_, err := dirhandle.Current()

			if err == nil {
				configure.PathWorker(func(index int, path string) {
					os.Chdir(path)
					runTargets(path, *targets)
				})
			}

		}

		// write script template
		if *execTemplate {
			someRunCmd = true
			WriteTemplate()
		}

		// run bash command over all targets
		// execute command on paths
		if *execute != "" {
			someRunCmd = true
			nonParams = false
			_, err := dirhandle.Current()
			var successCount = 0
			var errorCount = 0
			if err == nil {
				configure.PathWorker(func(index int, path string) {
					fmt.Print(output.MessageCln("execute on ", output.ForeWhite, path))
					os.Chdir(path)
					_, _, err := ExecuteScriptLine("bash", []string{"-c"}, *execute, func(output string) bool {
						fmt.Println(output)
						return true
					}, func(process *os.Process) {

					})
					if err != nil {
						errorCount++
						fmt.Println(output.MessageCln("\t", output.ForeRed, " Error:", err))
					} else {
						fmt.Println(output.MessageCln(output.ForeGreen, " OK"))
						successCount++
					}
				})
			} else {
				log.Fatal("error getting user dir", err)
			}
			fmt.Print("execution done. ")
			if errorCount > 0 {
				fmt.Print(output.MessageCln(output.ForeRed, errorCount, output.ForeWhite, " errors "))
			}
			if successCount > 0 {
				fmt.Print(output.MessageCln(output.ForeGreen, successCount, output.ForeWhite, " successes "))
			}
			fmt.Println(" ...")
		}
		// non defined commands found. check shortcuts
		if someRunCmd == false {
			shrtcut := false
			if len(os.Args) > 2 {
				log.Debug("got undefined argument. try to figure out meaning of ", os.Args[2])
				shrtcut = doRunShortCuts(os.Args[2])
			}
			if !shrtcut {
				log.Debug("no usage found for argument ", os.Args[2])
				printOutHeader()
				fmt.Println(output.MessageCln("to run a single target you can just type ", output.ForeWhite, "contxt run <target-name>"))
				fmt.Println()
				scriptCommand.PrintDefaults()
			}
		}
	}
	// DIR execution block
	if dirCommand.Parsed() {
		someDirCmd := false

		if *addCmd {
			nonParams = false
			someDirCmd = true
			dir, err := dirhandle.Current()
			if err == nil {
				fmt.Println(output.MessageCln("add ", output.ForeBlue, dir))
				configure.AddPath(dir)
				configure.SaveDefaultConfiguration(true)
			}

		}

		if *removeWorkSpace != "" {
			someDirCmd = true
			nonParams = false
			configure.RemoveWorkspace(*removeWorkSpace)
		}

		// changing worksspace
		if *workSpace != "" {
			someDirCmd = true
			nonParams = false
			configure.ChangeWorkspace(*workSpace, callBackOldWs, callBackNewWs)
		}

		if *clearPaths {
			someDirCmd = true
			nonParams = false
			configure.ClearPaths()
		}

		if *showWorkspaces {
			someDirCmd = true
			nonParams = false
			configure.DisplayWorkSpaces()
		}

		// show info
		if *info == true {
			someDirCmd = true
			nonParams = false
			printInfo()
		}

		// show paths
		if *showPaths == true {
			someDirCmd = true
			nonParams = false
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

		}

		// non of the arguments for dir was used
		if someDirCmd == false && *pathIndex == -1 {
			printOutHeader()
			dirCommand.PrintDefaults()
		}
	}

	// if nothing else happens we will change to the first path
	if nonParams == true && *pathIndex > -1 {
		dirhandle.PrintDir(*pathIndex)
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
	if !result {
		fmt.Println(output.MessageCln(output.BoldTag, param, output.CleanTag, " is not a workspace"))
	}
	return result
}

func doRunShortCuts(param string) bool {
	result := false
	template, _, exists := GetTemplate()
	if exists {
		for _, tasks := range template.Task {
			if tasks.ID == param {
				path, _ := dirhandle.Current()
				runTargets(path, tasks.ID)
				result = true
			}
		}
	}
	if !result {
		fmt.Println(output.MessageCln(output.BoldTag, param, output.CleanTag, " is not a valid task"))

		fmt.Println("\t", "these are the tasks they can be used as shortcut together with run")
		for _, tasks := range template.Task {
			fmt.Println("\t\t", tasks.ID)
		}
	}
	fmt.Println()

	return result
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

		fmt.Println(" contains paths:")
		configure.PathWorker(func(index int, path string) {

			template, _, exists := GetTemplate()
			if exists {
				outTasks := ""
				for _, tasks := range template.Task {
					outTasks = outTasks + " " + tasks.ID
				}
				fmt.Println(output.MessageCln("       path: ", output.Dim, " no ", output.ForeYellow, index, " ", output.ForeLightBlue, path, output.CleanTag, " targets", "[", output.ForeYellow, outTasks, output.CleanTag, "]"))
			} else {
				fmt.Println(output.MessageCln("       path: ", output.Dim, " no ", output.ForeYellow, index, " ", output.ForeLightBlue, path))
			}
		})
		fmt.Println()
		fmt.Println(output.MessageCln(" targets can be executes by ", "run -target <targetname>", "(for the current directory)"))
		fmt.Println(output.MessageCln(" a target can also be executed in all stored paths by ", "run -all-paths -target <targetname>", "independend from current path"))

		fmt.Println()
		fmt.Println(output.MessageCln(" all workspaces:", " ... change by ", "dir -w <workspace>", ""))
		configure.WorkSpaces(func(name string) {
			if name == configure.UsedConfig.CurrentSet {
				fmt.Println(output.MessageCln("\t[ ", output.BoldTag, name, output.CleanTag, " ]"))
			} else {
				fmt.Println(output.MessageCln("\t  ", name, "   "))
			}
		})
	}
}
