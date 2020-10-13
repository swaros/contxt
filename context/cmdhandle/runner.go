package cmdhandle

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/dirhandle"
	"github.com/swaros/contxt/context/systools"
)

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
	pathIndex := dirCommand.Int("i", 0, "get path by index. use -paths to see index and assigned paths")
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
		fmt.Println("not enough arguments")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "dir":
		dirCommand.Parse(os.Args[2:])
	case "run":
		scriptCommand.Parse(os.Args[2:])
	default:
		fmt.Println("unexpected command ", systools.Yellow(os.Args[1]))
		flag.PrintDefaults()
		os.Exit(1)
	}

	if scriptCommand.Parsed() {
		// run against a list of targets just in current dir
		if *targets != "" && *allDirs == false {
			nonParams = false
			path, _ := dirhandle.Current()
			runTargets(path, *targets)
		}

		// run targets in all paths
		if *targets != "" && *allDirs == true {
			nonParams = false
			dir, err := dirhandle.Current()

			if err == nil {
				configure.PathWorker(dir, func(index int, path string) {
					os.Chdir(path)
					runTargets(path, *targets)
				})
			}

		}

		// write script template
		if *execTemplate {
			WriteTemplate()
		}

		// run bash command over all targets
		// execute command on paths
		if *execute != "" {
			nonParams = false
			dir, err := dirhandle.Current()
			var successCount = 0
			var errorCount = 0
			if err == nil {
				configure.PathWorker(dir, func(index int, path string) {
					fmt.Print(systools.Purple("execute on "), systools.White(path))
					os.Chdir(path)
					err := ExecuteScriptLine("bash", *execute, func(output string) bool {
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
	}

	if dirCommand.Parsed() {
		if *addCmd {
			nonParams = false
			dir, err := dirhandle.Current()
			if err == nil {
				fmt.Println("add ", systools.Purple(dir))
				configure.AddPath(dir)
				configure.SaveDefaultConfiguration(true)
			}

		}

		if *removeWorkSpace != "" {
			configure.RemoveWorkspace(*removeWorkSpace)
		}

		if *removeWorkSpace != "" {
			configure.RemoveWorkspace(*removeWorkSpace)
		}

		if *workSpace != "" {
			configure.ChangeWorkspace(*workSpace)
		}

		if *clearPaths {
			configure.ClearPaths()
		}

		if *showWorkspaces {
			configure.DisplayWorkSpaces()
		}

		// show info
		if *info == true {
			nonParams = false
			fmt.Println("\t", "current workspace", systools.Green(configure.Config.CurrentSet))
		}

		// show paths
		if *showPaths == true {
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
	}

	// if nothing else happens we will change to the first path
	if nonParams == true {
		dirhandle.PrintDir(*pathIndex)
	}
}

func runTargets(path string, targets string) {
	allTargets := strings.Split(targets, ",")
	for _, runTarget := range allTargets {
		//ExecCurrentPathTemplate(runTarget)
		ExecPathFile(path+DefaultExecYaml, runTarget)
	}
}
