package outlaw

import (
	"fmt"
	"os"

	"github.com/abiosoft/ishell"
	"github.com/swaros/contxt/configure"
	"github.com/swaros/contxt/dirhandle"
	"github.com/swaros/contxt/taskrun"
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
	// adds workspaces to the list by callback iterator
	configure.WorkSpaces(func(s string) {
		desc := "other then now"
		if s == configure.UsedConfig.CurrentSet {
			desc = "this is the current workspace"
		}
		AddItemToSelect(selectItem{title: s, desc: desc})
	})
	selectedWs := uIselectItem("workspaces")
	if selectedWs.isSelected {
		c.Println("change to workspace: ", selectedWs.item.title)
		configure.ChangeWorkspace(selectedWs.item.title, taskrun.CallBackOldWs, taskrun.CallBackNewWs)
		return true
	}
	return false
}

func handleContexNavigation(c *ishell.Context) bool {
	configure.PathWorker(func(i int, s string) {
		AddItemToSelect(selectItem{title: fmt.Sprintf("CN %v", i), desc: s})
	})
	selectedCn := uIselectItem("select path in workspace ")
	if selectedCn.isSelected {
		if err := os.Chdir(selectedCn.item.desc); err != nil {
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
		for _, target := range targets {
			AddItemToSelect(selectItem{title: target, desc: ""})
		}
		selectedTarget := uIselectItem("select target to execute")
		if selectedTarget.isSelected {
			c.Println("running selected targets: ", selectedTarget.item.title)
			taskrun.RunTargets(selectedTarget.item.title, true)
			c.Println("done")
		}

	}
	return false
}
