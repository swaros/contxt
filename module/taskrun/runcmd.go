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
	"fmt"
	"strings"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/manout"
)

func PrintCnPaths() {
	fmt.Println(manout.MessageCln("\t", "paths stored in ", manout.ForeCyan, configure.GetGlobalConfig().UsedV2Config.CurrentSet))
	dir, err := dirhandle.Current()
	if err == nil {
		ShowPaths(dir)
	}
}

// ShowPaths : display all stored paths in the workspace
func ShowPaths(current string) {
	configure.GetGlobalConfig().PathWorkerNoCd(func(index string, path string) {
		if path == current {
			fmt.Println(manout.MessageCln("\t[", manout.ForeLightYellow, index, manout.CleanTag, "]\t", manout.BoldTag, path))
		} else {
			fmt.Println(manout.MessageCln("\t ", manout.ForeLightBlue, index, manout.CleanTag, " \t", path))
		}

	})
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
		fmt.Println(manout.MessageCln(manout.ForeWhite, " current workspace: ", manout.BoldTag, configure.GetGlobalConfig().UsedV2Config.CurrentSet))
		notWorkspace := true
		pathColor := manout.ForeLightBlue
		if !configure.GetGlobalConfig().PathMeightPartOfWs(dir) {
			pathColor = manout.ForeLightMagenta
		} else {
			notWorkspace = false
		}
		fmt.Println(" contains paths:")
		configure.GetGlobalConfig().PathWorker(func(index string, path string) {
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
		configure.GetGlobalConfig().ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
			if cfg.Name == configure.GetGlobalConfig().UsedV2Config.CurrentSet {
				fmt.Println(manout.MessageCln("\t[ ", manout.BoldTag, cfg.Name, manout.CleanTag, " ]"))
			} else {
				fmt.Println(manout.MessageCln("\t  ", cfg.Name, "   "))
			}
		})
	}
}
