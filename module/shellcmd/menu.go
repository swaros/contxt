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
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/taskrun"
	"github.com/swaros/manout"
)

func mainMenu(c *ishell.Context) {

	for {
		ApplyLogOut(UiLogger)
		currentPath := configure.GetGlobalConfig().GetActivePath("[path not set]")

		AddItemToSelect(selectItem{title: "Workspace", desc: "[" + configure.GetGlobalConfig().UsedV2Config.CurrentSet + "] change the active workspace"})
		AddItemToSelect(selectItem{title: "Contxt Navigation", desc: "[" + currentPath + "] change the active path in the current workspace "})

		AddItemToSelect(selectItem{title: "Show Variables", desc: "display the current variables"})
		if ok, err := systools.Exists("./.contxt.yml"); ok && err == nil {
			AddItemToSelect(selectItem{title: "verify .contxt.yml", desc: "display the current variables"})
			AddItemToSelect(selectItem{title: "Run Task", desc: "runs task in the current path (if exists)"})
		}

		AddItemToSelect(selectItem{title: "close", desc: "close the menu and go back to shell"})
		AddItemToSelect(selectItem{title: "exit", desc: "exit contxt"})
		menuOption := uIselectItem("Contxt Main menu @ "+configure.GetGlobalConfig().UsedV2Config.CurrentSet, true)
		switch menuOption.item.title {
		case "Workspace":
			workspaceMenu(c)
		case "Contxt Navigation":
			handleContexNavigation(c)
		case "Run Task":
			//handleRunCmds(c)
			//WaitForResponse()
			taskMenu()
		case "verify .contxt.yml":
			if w, _, err := systools.GetStdOutTermSize(); err == nil {
				taskrun.LintOut(w/2, w/2, false, false)
			} else {
				taskrun.LintOut(50, 50, false, false)
			}
			WaitForResponse()

		case "Show Variables":
			taskrun.GetPlaceHoldersFnc(func(phKey, phValue string) {
				UiLogger.Add(fmt.Sprintf("%s: %s", phKey, phValue))
				manout.Om.Println(manout.Message(manout.ForeCyan, phKey, ":", manout.ForeYellow, phValue))
			})
		case "close":
			manout.Om.Println("closing menu")
			return
		case "exit":
			manout.Om.Println("closing menu...and application")
			forceExit = true
			manout.Om.Println("bye bye...")
			systools.Exit(0)

			return
		default:
			manout.Om.Println("closing menu")
			return
		}

	}
}

func taskMenu() {
	if taskList, have := taskrun.GetAllTargets(); !have {
		UiLogger.Add("no tasks found")
		return
	} else {
		taskMenu := NewRunMenu(taskList, UiLogger)
		taskMenu.Run()
	}
}

func workspaceMenu(c *ishell.Context) {

	for {
		ApplyLogOut(UiLogger)
		AddItemToSelect(selectItem{title: "create new workspace", desc: "create a new workspace"})
		AddItemToSelect(selectItem{title: "select workspace", desc: "select an existing workspace"})
		AddItemToSelect(selectItem{title: "... back", desc: "close the menu and go back to shell"})
		menuOption := uIselectItem("Workspace menu", true)
		switch menuOption.item.title {
		case "create new workspace":
			handleCreateWorkspace(c)
		case "select workspace":
			handleWorkSpaces(c)
		case "... back":
			manout.Om.Println("closing menu")
			return
		default:
			manout.Om.Println("closing menu")
			return
		}

	}
}

func WaitForResponse() {
	taskrun.CtxOut("<f:white>   ------------------------")
	taskrun.CtxOut("</>press <f:yellow>enter</> to continue")
	fmt.Scanln()
}
