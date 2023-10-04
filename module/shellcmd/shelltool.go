// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

 package shellcmd

import (
	"os"
	"strconv"

	"github.com/abiosoft/ishell"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/taskrun"
	"github.com/swaros/manout"
)

func autoRecoverWs() {
	// TODO: old config is gone
}

var (
	UiLogger LogOutput = *NewAutoSizeLogOutput()
)

func inWs() bool {
	dir, err := dirhandle.Current()
	if err != nil {
		panic(err)
	}
	return configure.GetGlobalConfig().PathMeightPartOfWs(dir)
}

func resetShell() {
	taskrun.ClearAll()
	taskrun.MainInit()
}

// handleWorkSpaces display a list of workspace to select one.
// it returns true, if the workspace is switched
func handleWorkSpaces(c *ishell.Context) bool {
	//var ws []string = configure.GetGlobalConfig().ListWorkSpaces()
	// adds workspaces to the list by callback iterator

	configure.GetGlobalConfig().ExecOnWorkSpaces(func(wsName string, ws configure.ConfigurationV2) {
		AddItemToSelect(selectItem{title: wsName, desc: strconv.Itoa(len(ws.Paths)) + " stored paths"})
	})

	selectedWs := uIselectItem("Select Workspace ...", false)
	if selectedWs.isSelected {
		c.Println("change to workspace: ", selectedWs.item.title)
		UiLogger.Add("change to workspace: " + selectedWs.item.title)
		configure.GetGlobalConfig().ChangeWorkspace(selectedWs.item.title, taskrun.CallBackOldWs, taskrun.CallBackNewWs)
		return true
	}
	return false
}

func handleCreateWorkspace(c *ishell.Context) bool {
	wsName := ""
	if len(c.Args) > 0 {
		wsName = c.Args[0]
	} else {
		wsName, _ = TextInput("Enter Workspace Name", "new workspace name", 128, 25)
	}
	wsName, _ = systools.CheckForCleanString(wsName)
	UiLogger.Add("try to create workspace: (" + wsName + ")")

	if wsName != "" {
		if err := configure.GetGlobalConfig().AddWorkSpace(wsName, taskrun.CallBackOldWs, taskrun.CallBackNewWs); err != nil {
			manout.Om.Println(manout.ForeRed, "Error while trying to create workspace", manout.CleanTag, err)
			UiLogger.Add("error:" + err.Error())
			return false
		}
		c.Println("workspace created: ", wsName)
		UiLogger.Add("success  (" + wsName + ") created")
		return true
	}
	manout.Om.Println(manout.ForeRed, "Error while trying to create workspace ", manout.CleanTag, " no name given")
	UiLogger.Add("error: empty workspace name")
	return false
}

func handleContexNavigation(c *ishell.Context) bool {
	configure.GetGlobalConfig().PathWorkerNoCd(func(index, path string) {

		AddItemToSelect(selectItem{title: path, desc: index})
	})

	selectedCn := uIselectItem("choose path in "+configure.GetGlobalConfig().UsedV2Config.CurrentSet, false)
	if selectedCn.isSelected {
		if err := os.Chdir(selectedCn.item.title); err != nil {
			manout.Om.Print(manout.ForeRed, "Error while trying to enter path", manout.CleanTag, err)
			UiLogger.Add("Error while trying to enter path" + err.Error())
			return false
		}
		msg := manout.MessageCln(
			manout.ForeBlue,
			"... path changed ",
			manout.CleanTag,
			selectedCn.item.title, " ",
			manout.ForeLightGrey,
			selectedCn.item.desc,
			manout.CleanTag)
		UiLogger.Add("path changed to " + selectedCn.item.title)
		taskrun.CtxOut(msg)
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
