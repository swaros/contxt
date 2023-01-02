package shellcmd

import (
	"os"
	"strconv"

	"github.com/abiosoft/ishell"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/taskrun"
	"github.com/swaros/manout"
)

func autoRecoverWs() {
	// TODO: old config is gone
}

func inWs() bool {
	dir, err := dirhandle.Current()
	if err != nil {
		panic(err)
	}
	return configure.CfgV1.PathMeightPartOfWs(dir)
}

func resetShell() {
	taskrun.ClearAll()
	taskrun.MainInit()
}

// handleWorkSpaces display a list of workspace to select one.
// it returns true, if the workspace is switched
func handleWorkSpaces(c *ishell.Context) bool {
	//var ws []string = configure.CfgV1.ListWorkSpaces()
	// adds workspaces to the list by callback iterator

	configure.CfgV1.ExecOnWorkSpaces(func(wsName string, ws configure.ConfigurationV2) {
		AddItemToSelect(selectItem{title: wsName, desc: strconv.Itoa(len(ws.Paths)) + " stored paths"})
	})

	selectedWs := uIselectItem("Select Workspace ...", false)
	if selectedWs.isSelected {
		c.Println("change to workspace: ", selectedWs.item.title)
		configure.CfgV1.ChangeWorkspace(selectedWs.item.title, taskrun.CallBackOldWs, taskrun.CallBackNewWs)
		return true
	}
	return false
}

func handleContexNavigation(c *ishell.Context) bool {
	configure.CfgV1.PathWorkerNoCd(func(index, path string) {

		AddItemToSelect(selectItem{title: path, desc: index})
	})

	selectedCn := uIselectItem("choose path in "+configure.CfgV1.UsedV2Config.CurrentSet, false)
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
