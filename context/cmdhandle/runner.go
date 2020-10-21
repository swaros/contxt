package cmdhandle

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/dirhandle"
	"github.com/swaros/contxt/context/systools"
)

const version = "0.0.1-alpha"

// MainExecute runs main. parsing flags
func MainExecute() {

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
		fmt.Println(systools.White("  dir"), "\tmanaging paths in workspaces")
		fmt.Println(systools.White("  run"), "\texecuting scripts in workspaces")
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
			fmt.Println("unexpected command ", systools.Yellow(os.Args[1]))
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
					fmt.Print(systools.Purple("execute on "), systools.White(path))
					os.Chdir(path)
					_, err := ExecuteScriptLine("bash", *execute, func(output string) bool {
						fmt.Println(output)
						return true
					}, func(process *os.Process) {

					})
					if err != nil {
						errorCount++
						fmt.Println("\t", systools.Red(" Error:"), err)
					} else {
						fmt.Println(systools.Green(" OK"))
						successCount++
					}
				})
			}
			fmt.Print("execution done. ")
			if errorCount > 0 {
				fmt.Print(systools.Red(errorCount), " errors ")
			}
			if successCount > 0 {
				fmt.Print(systools.Green(successCount), " successfully ")
			}
			fmt.Println(" ...")
		}
		// non defined commands found. check shortcuts
		if someRunCmd == false {
			shrtcut := false
			if len(os.Args) > 2 {
				shrtcut = doRunShortCuts(os.Args[2])
			}
			if !shrtcut {
				printOutHeader()
				fmt.Println("to run a single target you can just type ", systools.White("contxt run <target-name>"))
				fmt.Println()
				scriptCommand.PrintDefaults()
			}
		}
	}

	if dirCommand.Parsed() {
		someDirCmd := false

		if *addCmd {
			nonParams = false
			someDirCmd = true
			dir, err := dirhandle.Current()
			if err == nil {
				fmt.Println("add ", systools.Purple(dir))
				configure.AddPath(dir)
				configure.SaveDefaultConfiguration(true)
			}

		}

		if *removeWorkSpace != "" {
			someDirCmd = true
			nonParams = false
			configure.RemoveWorkspace(*removeWorkSpace)
		}

		if *workSpace != "" {
			someDirCmd = true
			nonParams = false
			configure.ChangeWorkspace(*workSpace)
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
			fmt.Println("\t", "paths stored in ", systools.Green(configure.Config.CurrentSet))
			dir, err := dirhandle.Current()
			if err == nil {
				count := configure.ShowPaths(dir)
				if count > 0 {
					fmt.Println()
					fmt.Println("\t", "to change directory depending stored path you can write", systools.Purple("cd $(", os.Args[0], " -i ", count-1, ")"), "in bash")
					fmt.Println("\t", "this will be the same as ", systools.Magenta("cd ", dirhandle.GetDir(count-1)))
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

func doMagicParamOne(param string) bool {
	result := false
	// param is a workspace ?
	configure.WorkSpaces(func(ws string) {
		if param == ws {
			configure.ChangeWorkspace(ws)
			result = true
		}
	})
	if !result {
		fmt.Println(systools.Teal(param), "is not a workspace")
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
		fmt.Println("\t", systools.Yellow(param), "is not a valid task.")
		fmt.Println("\t", systools.White("these are the tasks they can be used as shortcut together with run"))
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
	fmt.Println(systools.White("contxt"), version)
}

func printInfo() {
	printOutHeader()
	printPaths()
	//systools.TestPrintColoredChanges()
}

func printPaths() {
	dir, err := dirhandle.Current()
	if err == nil {
		fmt.Println(systools.White(" current directory:"), dir)
		fmt.Println(systools.White(" current workspace:"), configure.Config.CurrentSet)
		fmt.Println(" contains paths:")
		configure.PathWorker(func(index int, path string) {

			template, _, exists := GetTemplate()
			if exists {
				outTasks := ""
				for _, tasks := range template.Task {
					outTasks = outTasks + " " + tasks.ID
				}
				fmt.Println(systools.White("       path:"), "no", systools.Yellow(index), path, systools.White("targets"), "[", systools.Teal(outTasks), "]")
			} else {
				fmt.Println(systools.White("       path:"), "no", systools.Yellow(index), path)
			}
		})
		fmt.Println()
		fmt.Println(" targets can be executes by ", systools.Teal("run -target <targetname>"), "(for the current directory)")
		fmt.Println(" a target can also be executed in all stored paths by ", systools.Teal("run -all-paths -target <targetname>"), "independend from current path")

		fmt.Println()
		fmt.Println(systools.White(" all workspaces:"), " ... change by ", systools.Teal("dir -w <workspace>"), "")
		configure.WorkSpaces(func(name string) {
			if name == configure.Config.CurrentSet {
				fmt.Println("\t[", systools.White(name), "]")
			} else {
				fmt.Println("\t ", name)
			}
		})
	}
}
