package shellcmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/taskrun"
	"github.com/swaros/manout"
)

func autoRecoverWs() {
	if !inWs() {
		configure.WorkSpaces(func(ws string) {
			if configure.UsedConfig.CurrentSet == ws {
				configure.ChangeWorkspace(ws, taskrun.CallBackOldWs, taskrun.CallBackNewWs)
			}
		})
	}
}

func inWs() bool {
	dir, err := dirhandle.Current()
	if err != nil {
		panic(err)
	}
	return configure.PathMeightPartOfWs(dir)
}

func resetShell() {
	taskrun.ClearAll()
	taskrun.MainInit()
}

// handleWorkSpaces display a list of workspace to select one.
// it returns true, if the workspace is switched
func handleWorkSpaces(c *ishell.Context) bool {
	var ws []string
	// adds workspaces to the list by callback iterator
	configure.WorkSpaces(func(s string) {
		ws = append(ws, s)
	})
	selectedWs := simpleSelect("workspaces", ws)
	if selectedWs.isSelected {
		c.Println("change to workspace: ", selectedWs.item.title)
		configure.ChangeWorkspace(selectedWs.item.title, taskrun.CallBackOldWs, taskrun.CallBackNewWs)
		return true
	}
	return false
}

func handleContexNavigation(c *ishell.Context) bool {
	workspace, err := taskrun.CollectWorkspaceInfos() // get workspace meta-info
	if err != nil {
		manout.Om.Print(manout.ForeRed, "Error parsing workspace", manout.CleanTag, err)
		return false
	}

	for _, wsPath := range workspace.Paths { // iterate the path infos
		if wsPath.HaveTemplate {
			// build description by the tasks the beeing used
			label := fmt.Sprintf("%d tasks:  ", len(wsPath.Targets))
			AddItemToSelect(selectItem{title: wsPath.Path, desc: label + strings.Join(wsPath.Targets, "|")})
		} else {
			// plain added path. no template there
			AddItemToSelect(selectItem{title: wsPath.Path, desc: "no tasks in this path"})
		}

	}

	selectedCn := uIselectItem("choose path in " + workspace.CurrentWs)
	if selectedCn.isSelected {
		if err := os.Chdir(selectedCn.item.title); err != nil {
			manout.Om.Print(manout.ForeRed, "Error while trying to enter path", manout.CleanTag, err)
			return false
		}
		c.Println(
			manout.MessageCln(
				manout.ForeBlue,
				"... path changed ",
				manout.CleanTag,
				selectedCn.item.title, " ",
				manout.ForeLightGrey,
				selectedCn.item.desc,
				manout.CleanTag))
		return true
	}
	return false
}

func handleRunCmds(c *ishell.Context) bool {
	if targets, found := taskrun.GetAllTargets(); found {
		// commented. maybe we add an description for run targets so the other list would makes sense
		/*
			for _, target := range targets {
				AddItemToSelect(selectItem{title: target, desc: ""})
			}*/
		selectedTarget := simpleSelect("select target to execute", targets)
		if selectedTarget.isSelected {
			c.Println("running selected targets: ", selectedTarget.item.title)
			taskrun.RunTargets(selectedTarget.item.title, true)
			c.Println("done")
		}

	}
	return false
}
