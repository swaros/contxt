// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Licensed under the MIT License
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// collection of comands they can be used for any comand interpreter like cobra and ishell.
package taskrun

import (
	"errors"
	"fmt"
	"strings"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/manout"
)

func PrintCnPaths(hints bool) {
	fmt.Println(manout.MessageCln("\t", "paths stored in ", manout.ForeCyan, configure.UsedConfig.CurrentSet))
	dir, err := dirhandle.Current()
	if err == nil {
		count := ShowPaths(dir)
		if count > 0 && hints {
			fmt.Println()
			fmt.Println(manout.MessageCln("\t", "if you have installed the shell functions ", manout.ForeDarkGrey, "(contxt install bash|zsh|fish)", manout.CleanTag, " change the directory by ", manout.BoldTag, "cn ", count-1))
			fmt.Println(manout.MessageCln("\t", "this will be the same as ", manout.BoldTag, "cd ", dirhandle.GetDir(count-1)))
		}
	}
}

// ShowPaths : display all stored paths in the workspace
func ShowPaths(current string) int {

	configure.PathWorkerNoCd(func(index int, path string) {
		if path == current {
			fmt.Println(manout.MessageCln("\t[", manout.ForeLightYellow, index, manout.CleanTag, "]\t", manout.BoldTag, path))
		} else {
			fmt.Println(manout.MessageCln("\t ", manout.ForeLightBlue, index, manout.CleanTag, " \t", path))
		}

	})
	return len(configure.UsedConfig.Paths)
}

func GetAllTargets() ([]string, bool) {
	plainTargets, found := targetsAsMap()
	template, _, exists, terr := GetTemplate()
	if terr != nil {
		return plainTargets, found
	}
	if exists {
		shareds := detectSharedTargetsAsMap(template)
		plainTargets = append(plainTargets, shareds...)
	}
	return plainTargets, exists && found
}

func detectSharedTargetsAsMap(current configure.RunConfig) []string {
	var targets []string
	SharedFolderExecuter(current, func(_, _ string) {
		sharedTargets, have := targetsAsMap()
		if have {
			targets = append(targets, sharedTargets...)
		}
	})

	return targets
}

func ExistInStrMap(testStr string, check []string) bool {
	for _, str := range check {
		if strings.TrimSpace(str) == strings.TrimSpace(testStr) {
			return true
		}
	}
	return false
}

func targetsAsMap() ([]string, bool) {
	var targets []string
	template, _, exists, terr := GetTemplate()
	if terr != nil {
		targets = append(targets, terr.Error())
		return targets, false
	}
	if exists {
		return templateTargetsAsMap(template)
	}
	return targets, false
}

func templateTargetsAsMap(template configure.RunConfig) ([]string, bool) {
	var targets []string
	found := false

	if len(template.Task) > 0 {
		for _, tasks := range template.Task {
			if !ExistInStrMap(tasks.ID, targets) && (!tasks.Options.Invisible || showInvTarget) {
				found = true
				targets = append(targets, strings.TrimSpace(tasks.ID))
			}
		}
	}

	return targets, found
}

func printTargets() {

	template, path, exists, terr := GetTemplate()
	if terr != nil {
		return
	}
	if exists {
		fmt.Println(manout.MessageCln(manout.ForeDarkGrey, "used taskfile:\t", manout.CleanTag, path))
		fmt.Println(manout.MessageCln(manout.ForeDarkGrey, "tasks count:  \t", manout.CleanTag, len(template.Task)))
		if len(template.Task) > 0 {
			fmt.Println(manout.MessageCln(manout.BoldTag, "existing targets:"))
			taskList, _ := templateTargetsAsMap(template)
			for _, tasks := range taskList {
				fmt.Println("\t", tasks)
			}
		} else {
			fmt.Println(manout.MessageCln("that is what we gor so far:"))
			fmt.Println()
		}

		sharedTargets := detectSharedTargetsAsMap(template)
		if len(sharedTargets) > 0 {

			for _, stasks := range sharedTargets {
				fmt.Println("\t", stasks, manout.MessageCln(manout.ForeDarkGrey, " shared", manout.CleanTag))
			}

		}
	} else {
		fmt.Println(manout.MessageCln(manout.ForeCyan, "no task-file exists. you can create one by ", manout.CleanTag, " contxt create"))
	}
}

func printPaths() {
	dir, err := dirhandle.Current()
	if err == nil {
		fmt.Println(manout.MessageCln(manout.ForeWhite, " current directory: ", manout.BoldTag, dir))
		fmt.Println(manout.MessageCln(manout.ForeWhite, " current workspace: ", manout.BoldTag, configure.UsedConfig.CurrentSet))
		notWorkspace := true
		pathColor := manout.ForeLightBlue
		if !configure.PathMeightPartOfWs(dir) {
			pathColor = manout.ForeLightMagenta
		} else {
			notWorkspace = false
		}
		fmt.Println(" contains paths:")
		configure.PathWorker(func(index int, path string) {
			template, _, exists, _ := GetTemplate()
			add := ""
			if strings.Contains(dir, path) {
				add = manout.ResetDim + manout.ForeCyan
			}
			if dir == path {
				add = manout.ResetDim + manout.ForeGreen
			}
			if exists {
				outTasks := ""
				targets, _ := templateTargetsAsMap(template)
				for _, tasks := range targets {
					outTasks = outTasks + " " + tasks
				}

				fmt.Println(manout.MessageCln("       path: ", manout.Dim, " no ", manout.ForeYellow, index, " ", pathColor, add, path, manout.CleanTag, " targets", "[", manout.ForeYellow, outTasks, manout.CleanTag, "]"))
			} else {
				fmt.Println(manout.MessageCln("       path: ", manout.Dim, " no ", manout.ForeYellow, index, " ", pathColor, add, path))
			}
		}, func(origin string) {})
		if notWorkspace {
			fmt.Println()
			fmt.Println(manout.MessageCln(manout.BackYellow, manout.ForeBlue, " WARNING ! ", manout.CleanTag, "\tyou are currently in none of the assigned locations."))
			fmt.Println("\t\tso maybe you are using the wrong workspace")
		}
		if !showHints {
			fmt.Println()
			fmt.Println(manout.MessageCln(" targets can be executes by ", manout.BoldTag, "contxt run <targetname>", manout.CleanTag, "(for the current directory)"))
			fmt.Println(manout.MessageCln(" a target can also be executed in all stored paths by ", manout.BoldTag, "contxt run -a <targetname>", manout.CleanTag, " independend from current path"))
		}

		fmt.Println()
		if !showHints {
			fmt.Println(manout.MessageCln(" all workspaces:", " ... change by ", manout.BoldTag, "contxt <workspace>", ""))
		} else {
			fmt.Println(manout.MessageCln(" all workspaces:"))
		}
		configure.WorkSpaces(func(name string) {
			if name == configure.UsedConfig.CurrentSet {
				fmt.Println(manout.MessageCln("\t[ ", manout.BoldTag, name, manout.CleanTag, " ]"))
			} else {
				fmt.Println(manout.MessageCln("\t  ", name, "   "))
			}
		})
	}
}

type pathInfo struct {
	Path         string                     // the stored path
	Targets      []string                   // all existing targets
	Active       bool                       // this is the active path
	IsSubDir     bool                       // this path is the active or a subdir of current dir
	HaveTemplate bool                       // in this folder a template exists
	Project      configure.ProjectWorkspace // infos about the project could be there
}

type workspace struct {
	CurrentDir string     // the current directory
	CurrentWs  string     // the name of the current workspace
	InWs       bool       // flag if the current path is part of the workspace
	Paths      []pathInfo // all stored paths
}

func CollectWorkspaceInfos() (workspace, error) {
	var ws workspace
	if configure.UsedConfig.CurrentSet == "" {
		return ws, errors.New("no workspace loaded")
	}
	dir, err := dirhandle.Current()
	if err == nil {
		ws.CurrentDir = dir
		ws.CurrentWs = configure.UsedConfig.CurrentSet
		ws.InWs = configure.PathMeightPartOfWs(dir)

		configure.PathWorker(func(index int, path string) {
			var pInfo pathInfo
			pInfo.Path = path
			template, _, exists, _ := GetTemplate()
			pInfo.HaveTemplate = exists
			pInfo.Active = (dir == path)
			pInfo.Project = template.Workspace
			pInfo.IsSubDir = strings.Contains(dir, path)
			if exists {
				pInfo.Targets, _ = templateTargetsAsMap(template)
			}
			ws.Paths = append(ws.Paths, pInfo)

		}, func(origin string) {})
		return ws, nil
	}
	return ws, err
}
