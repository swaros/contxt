package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/swaros/contxt/context/cmdhandle"
	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/dirhandle"
	"github.com/swaros/contxt/context/systools"
)

func main() {

	savePtr := flag.Bool("add", false, "add current dir to workspace")
	showPaths := flag.Bool("paths", false, "show current paths")
	showWorkspaces := flag.Bool("l", false, "display all existing workspaces")
	pathIndex := flag.Int("i", 0, "get path by index. use -paths to see index and assigned paths")
	workSpace := flag.String("w", "UNSET", "set current workspace")
	removeWorkSpace := flag.String("delete", "UNSET", "remove workspace")
	clearPaths := flag.Bool("clear", false, "remove all paths from workspace")
	info := flag.Bool("info", false, "show current workspace")
	execute := flag.String("exec", "UNSET", "Execute a command on all paths")
	execTemplate := flag.Bool("create-template", false, "write template for path dependeing executions in current folder")

	scriptRun := flag.Bool("script", false, "run template in current folder")
	scriptRunAll := flag.Bool("script-all", false, "run template in all Folders")

	// targets
	targetInit := flag.Bool("init", false, "set script target to init")
	targetClean := flag.Bool("clean", false, "set script target to clean")
	targetTest := flag.Bool("test", false, "set script target to test")

	nonParams := true

	runTargetName := cmdhandle.TargetScript

	flag.Parse()

	if *targetInit {
		runTargetName = cmdhandle.InitScript
	}

	if *targetClean {
		runTargetName = cmdhandle.ClearScript
	}

	if *targetTest {
		runTargetName = cmdhandle.TestScript
	}

	// check on target commands
	if runTargetName != cmdhandle.TargetScript {
		if *scriptRun == false && *scriptRunAll == false {
			*scriptRun = true
		}
	}

	var configErr = configure.InitConfig()
	if configErr != nil {
		log.Fatal(configErr)
	}

	if *scriptRun {
		nonParams = false
		cmdhandle.ExecCurrentPathTemplate(runTargetName)
	}

	if *execTemplate {
		cmdhandle.WriteTemplate()
	}

	if *removeWorkSpace != "UNSET" {
		configure.RemoveWorkspace(*removeWorkSpace)
	}

	if *workSpace != "UNSET" {
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

	if *scriptRunAll {
		nonParams = false
		dir, err := dirhandle.Current()

		if err == nil {
			configure.PathWorker(dir, func(index int, path string) {
				os.Chdir(path)
				cmdhandle.ExecPathFile(path+cmdhandle.DefaultExecFile, runTargetName)
			})
		}
	}

	// execute command on paths
	if *execute != "UNSET" {
		nonParams = false
		dir, err := dirhandle.Current()
		var successCount = 0
		var errorCount = 0
		if err == nil {
			configure.PathWorker(dir, func(index int, path string) {
				fmt.Print(systools.Purple("execute on "), systools.White(path))
				os.Chdir(path)
				out, errout, err := cmdhandle.Shellout("bash", *execute)
				if err != nil {
					errorCount++
					fmt.Println("\t", systools.Red(" Error:"), systools.Yellow(errout))
				} else {
					fmt.Println(systools.Green(" OK"))
					successCount++
				}
				fmt.Printf("%s", out)
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

	// save current path
	if *savePtr == true {
		nonParams = false
		dir, err := dirhandle.Current()
		if err == nil {
			fmt.Println("add ", systools.Purple(dir))
			configure.AddPath(dir)
			configure.SaveDefaultConfiguration(true)
		}
	}

	// if nothing else happens we will change to the first path
	if nonParams == true {
		dirhandle.PrintDir(*pathIndex)
	}
}
