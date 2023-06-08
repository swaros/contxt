// Copyright (c) 2023 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// # Licensed under the MIT License
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
package runner

import (
	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
)

type CmdExecutor interface {
	Print(msg ...interface{})
	Println(msg ...interface{})
	PrintPaths(plain bool, showFulltask bool)                         // print out all paths
	GetLogger() *logrus.Logger                                        // get logger
	GetOuputHandler() (ctxout.StreamInterface, ctxout.PrintInterface) // get output handlers
	SetLogLevel(level string) error                                   // set log level
	ResetVariables()                                                  // reset old variables while change the workspace. (req for shell mode)
	MainInit()                                                        // initialize the workspace
	doMagicParamOne(string)
	RunTargets(string, bool)               // run targets
	GetTargets(incInvisible bool) []string // return all targets. optional include invisible targets
	CallBackNewWs(string)                  // callback for new workspace
	CallBackOldWs(string) bool             // callback for old workspace
	FindWorkspaceInfoByTemplate(updateFn func(workspace string, cnt int, update bool, info configure.WorkspaceInfoV2)) (allCount int, updatedCount int)
	PrintWorkspaces()                                  // print out all workspaces
	GetWorkspaces() []string                           // print out all workspaces as a list
	DirFindApplyAndSave(args []string) (string, error) // find pathbay arguments,save the current path print the path
	InteractiveScreen()                                // interactive screen
	GetCurrentWorkSpace() string                       // get current workspace
}
