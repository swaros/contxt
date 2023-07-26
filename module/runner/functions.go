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
	"errors"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
	"github.com/swaros/contxt/module/yaclint"
)

type CmdExecutorImpl struct {
	session *CmdSession
}

func NewCmd(session *CmdSession) *CmdExecutorImpl {
	return &CmdExecutorImpl{
		session: session,
	}
}

func (c *CmdExecutorImpl) Combine4Print(msg ...interface{}) []interface{} {
	var outInterfaces []interface{}
	outInterfaces = append(outInterfaces, c.session.OutPutHdnl)
	outInterfaces = append(outInterfaces, c.session.Printer)
	outInterfaces = append(outInterfaces, msg...)
	return outInterfaces
}

func (c *CmdExecutorImpl) MessageToString(msg ...interface{}) string {
	msg = c.Combine4Print(msg...)
	return ctxout.ToString(msg...)
}

func (c *CmdExecutorImpl) Print(msg ...interface{}) {
	ctxout.Print(c.Combine4Print(msg...)...)
}

func (c *CmdExecutorImpl) Println(msg ...interface{}) {
	ctxout.PrintLn(c.Combine4Print(msg...)...)
}

func (c *CmdExecutorImpl) doMagicParamOne(args string) {
}

func (c *CmdExecutorImpl) CallBackOldWs(oldws string) bool {
	c.session.Log.Logger.Info("OLD workspace: ", oldws)
	// get all paths first
	configure.GetGlobalConfig().PathWorkerNoCd(func(_ string, path string) {

		current := dirhandle.Pushd()
		template, exists, _ := c.session.TemplateHndl.Load()
		Fields := logrus.Fields{
			"template: ": template,
			"exists":     exists,
			"path":       path,
		}
		c.session.Log.Logger.Debug("path parsing ", Fields)

		if exists && template.Config.Autorun.Onleave != "" {
			onleaveTarget := template.Config.Autorun.Onleave
			Fields := logrus.Fields{
				"target": onleaveTarget,
			}
			c.session.Log.Logger.Info("execute leave-action", Fields)
			c.RunTargets(onleaveTarget, true)

		}
		current.Popd()

	})
	return true
}

func (c *CmdExecutorImpl) CallBackNewWs(newWs string) {
	c.ResetVariables() // reset old variables while change the workspace. (req for shell mode)
	c.session.Log.Logger.Info("NEW workspace: ", newWs)
	configure.GetGlobalConfig().PathWorker(func(_ string, path string) { // iterate any path
		template, exists, _ := c.session.TemplateHndl.Load()
		Fields := logrus.Fields{
			"template: ": template,
			"exists":     exists,
			"path":       path,
		}
		c.session.Log.Logger.Debug("path parsing", Fields)

		// try to run onEnter func at any possible target in the workspace
		if exists && template.Config.Autorun.Onenter != "" {
			Fields := logrus.Fields{
				"target": template.Config.Autorun.Onenter,
			}
			onEnterTarget := template.Config.Autorun.Onenter
			c.session.Log.Logger.Info("execute enter-action", Fields)
			c.RunTargets(onEnterTarget, true)
		}

	}, func(origin string) {
		Fields := logrus.Fields{
			"current-dir": origin,
		}
		c.session.Log.Logger.Debug("done calling autoruns on sub-dirs", Fields)
	})
}

// set the default runtime variables depeding the predefined variables from
// the main init, and the given variables depending the task and environment
func (c *CmdExecutorImpl) SetStartupVariables(dataHndl *tasks.CombinedDh, template *configure.RunConfig) {
	c.session.Log.Logger.Debug("set startup variables")
	// get the predifined variables from the MainInit function
	// and set them to the datahandler
	for k, v := range c.session.DefaultVariables {
		dataHndl.SetPH(k, v)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		c.session.Log.Logger.Error("error while getting current dir", err)
	} else {
		// we will override the current dir from the predefined ones, with the current dir
		dataHndl.SetPH("CTX_PWD", currentDir)
	}

	// template depending variables
	dataHndl.SetPH("CTX_PROJECT", template.Workspace.Project)
	dataHndl.SetPH("CTX_ROLE", template.Workspace.Role)
	dataHndl.SetPH("CTX_VERSION", template.Workspace.Version)

	dataHndl.SetPH("CTX_WS", configure.GetGlobalConfig().UsedV2Config.CurrentSet)
	keys := ""
	configure.GetGlobalConfig().ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
		for _, ws2 := range cfg.Paths {
			keys += setConfigVaribales(dataHndl, ws2, "WS") // there a space is added at the end already
			c.session.Log.Logger.Debug("set startup variables for ws2", keys)
		}

	})
	dataHndl.SetPH("CTX_WS_KEYS", keys)
	// read the imports from the template and set them to the datahandler
	c.handleImports(dataHndl, template)
}

// taking care about the imports.
// these are the imports from the template, defined in config.imports and they are
// used as variables map for the current run.
// these imports will be set as map[string]string to the datahandler as long the are json or yaml files.
// and can be used as placeholder in the tasks.
// for example:
//
//	imports:
//	  - imports.json
//	  - imports.yaml
//
// this can be used in the tasks as ${imports.json:key1} or ${imports.yaml:key1}
// but if an string is given, sperated by a space, the string will be used as key.
// for example:
//
//	imports:
//	  - imports.json jmap
//	  - imports.yaml ymap
//
// this can be used in the tasks as ${jmap:key1} or ${ymap:key1}
// any other file type will be loaded as string and assigned to the key with the whole content of the file.
// for example:
//
//	imports:
//	  - imports.txt
//
// this can be used in the tasks as ${imports.txt}
// or by using the key by the import
// for example:
//
//	imports:
//	  - imports.txt txt
//
// this can be used in the tasks as ${txt}
func (c *CmdExecutorImpl) handleImports(dataHndl *tasks.CombinedDh, template *configure.RunConfig) {
	importHndlr := NewImportHandler(c.session.Log.Logger, dataHndl, c.session.TemplateHndl)
	importHndlr.SetImports(template.Config.Imports)
	importHndlr.HandleImports()
}

func setConfigVaribales(dataHndl *tasks.CombinedDh, wsInfo configure.WorkspaceInfoV2, varPrefix string) string {
	pathStrInfo := ""
	if wsInfo.Project != "" && wsInfo.Role != "" {
		prefix := wsInfo.Project + "_" + wsInfo.Role
		pathkey := varPrefix + "0_" + prefix
		dataHndl.SetPH(pathkey, wsInfo.Path) // at least XXX0 without any version. this could be overwritten by other checkouts
		pathStrInfo += pathkey + " "
		if wsInfo.Version != "" {
			// if version is set, we use them for avoid conflicts with different checkouts
			if versionSan, err := systools.CheckForCleanString(wsInfo.Version); err == nil {
				prefix += "_" + versionSan
				// add it to ws1 as prefix for versionized keys
				dataHndl.SetPH(varPrefix+"1_"+prefix, wsInfo.Path)
			}
		}
	}
	return pathStrInfo
}

// RunTargets run the given targets
// force is used as flag for the first level targets, and is used
// to runs shared targets once in front of the regular assigned targets
func (c *CmdExecutorImpl) RunTargets(target string, force bool) error {
	if template, exists, err := c.session.TemplateHndl.Load(); err != nil {
		c.session.Log.Logger.Error("error while loading template", err)
		return err
	} else if !exists {
		c.session.Log.Logger.Error("template not exists")
		return errors.New("no contxt template found in current directory")
	} else {

		datahndl := tasks.NewCombinedDataHandler()
		c.SetStartupVariables(datahndl, &template)
		datahndl.SetPH("CTX_TARGET", target)
		datahndl.SetPH("CTX_FORCE", strconv.FormatBool(force))

		requireHndl := tasks.NewDefaultRequires(datahndl, c.session.Log.Logger)
		executer := tasks.NewTaskListExec(
			template,
			datahndl,
			requireHndl,
			c.getOutHandler(),
			tasks.ShellCmd,
		)
		executer.SetLogger(c.session.Log.Logger)
		code := executer.RunTarget(target, force)
		switch code {
		case systools.ExitByNoTargetExists:
			c.session.Log.Logger.Error("target not exists:", target)
			return errors.New("target " + target + " not exists")
		case systools.ExitAlreadyRunning:
			c.session.Log.Logger.Info("target already running")
			return nil
		case systools.ExitCmdError:
			c.session.Log.Logger.Error("error while running target ", target)
			return errors.New("error while running target:" + target)
		case systools.ExitByNothingToDo:
			c.session.Log.Logger.Info("nothing to do ", target)
			return nil
		case systools.ExitOk:
			c.session.Log.Logger.Info("target executed successfully")
			return nil
		default:
			c.session.Log.Logger.Error("unexpected exit code:", code)
			return errors.New("unexpected exit code:" + fmt.Sprintf("%d", code))
		}
	}

}

func (c *CmdExecutorImpl) GetTargets(incInvisible bool) []string {
	if template, exists, err := c.session.TemplateHndl.Load(); err != nil {
		c.session.Log.Logger.Error("error while loading template", err)
	} else if !exists {
		c.session.Log.Logger.Error("template not exists", err)
	} else {
		if res, have := TemplateTargetsAsMap(template, incInvisible); have {
			return res
		}
	}
	return nil
}

func (c *CmdExecutorImpl) ResetVariables() {
}

func (c *CmdExecutorImpl) MainInit() {
	c.initDefaultVariables()
}

// initDefaultVariables init the default variables for the current session.
// these are the varibales they should not change during the session.
func (c *CmdExecutorImpl) initDefaultVariables() {
	if currentPath, err := os.Getwd(); err != nil {
		ctxout.CtxOut("Error while reading current directory", err)
		systools.Exit(systools.ErrorBySystem)
	} else {
		c.setVariable("CTX_PWD", currentPath)
		c.setVariable("CTX_PATH", currentPath)
	}
	c.setVariable("CTX_OS", runtime.GOOS)
	c.setVariable("CTX_ARCH", runtime.GOARCH)
	c.setVariable("CTX_USER", os.Getenv("USER"))
	c.setVariable("CTX_HOST", getHostname())
	c.setVariable("CTX_HOME", os.Getenv("HOME"))
	c.setVariable("CTX_DATE", time.Now().Format("2006-01-02"))
	c.setVariable("CTX_TIME", time.Now().Format("15:04:05"))
	c.setVariable("CTX_DATETIME", time.Now().Format("2006-01-02 15:04:05"))

	c.setVariable("CTX_VERSION", configure.GetVersion())
	c.setVariable("CTX_BUILD_NO", configure.GetBuild())

	c.handleWindowsInit() // it self is testing if we are on windows
}

func getHostname() string {
	if hostname, err := os.Hostname(); err != nil {
		return ""
	} else {
		return hostname
	}
}

func (c *CmdExecutorImpl) handleWindowsInit() {
	if runtime.GOOS == "windows" {
		// we need to set the console to utf8, to be able to print utf8 chars
		// in the console
		if os.Getenv("CTX_COLOR") == "ON" { // then lets see if this should forced for beeing enabled by env-var
			c.SetColor(true)
		} else {
			// if not forced already we try to figure out, by oure own, if the powershell is able to support ANSII
			// this is since version 7 the case
			pwrShellRunner := tasks.GetShellRunnerForOs("windows")
			version, _ := pwrShellRunner.ExecSilentAndReturnLast(PWRSHELL_CMD_VERSION)
			c.setVariable("CTX_PS_VERSION", version) // also setup varibale to have the PS version in place
			if version >= "7" {
				c.SetColor(true)
			}
		}
	}
}

// updates the given variable in the current session.
// this is just for keeping the variable in the session. but this
// is not used as variables for the template.
// this is just ment, to define already variables while setting up
// the session and keep them in the session, until they get used by the template later.
// see RunTargets for the usage of the variables.
func (c *CmdExecutorImpl) setVariable(name string, value string) {
	c.session.DefaultVariables[name] = value
}

func (c *CmdExecutorImpl) GetVariable(name string) string {
	if val, have := c.session.DefaultVariables[name]; have {
		return val
	}
	return ""
}

func (c *CmdExecutorImpl) GetVariables() map[string]string {
	return c.session.DefaultVariables
}

func (c *CmdExecutorImpl) SetColor(onoff bool) {
	behave := ctxout.GetBehavior()
	behave.NoColored = onoff
	ctxout.SetBehavior(behave)
}

func (c *CmdExecutorImpl) GetOuputHandler() (ctxout.StreamInterface, ctxout.PrintInterface) {
	return c.session.OutPutHdnl, c.session.Printer
}

func (c *CmdExecutorImpl) GetWorkspaces() []string {
	ws := configure.GetGlobalConfig().ListWorkSpaces()
	sort.Strings(ws)
	return ws
}

func (c *CmdExecutorImpl) FindWorkspaceInfoByTemplate(updateFn func(workspace string, cnt int, update bool, info configure.WorkspaceInfoV2)) (allCount int, updatedCount int) {
	wsCount := 0
	wsUpdated := 0
	c.session.Log.Logger.Info("Start to find workspace info by template")

	if currentPath, err := os.Getwd(); err != nil {
		ctxout.CtxOut("Error while reading current directory", err)
		systools.Exit(systools.ErrorBySystem)
	} else {
		haveUpdate := false
		configure.GetGlobalConfig().ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
			wsCount++
			for pathIndex, savedWorkspace := range cfg.Paths {
				logFields := mimiclog.Fields{"path": savedWorkspace.Path, "project": savedWorkspace.Project, "role": savedWorkspace.Role}
				c.session.Log.Logger.Debug("parsing workspace", logFields)
				if err := os.Chdir(savedWorkspace.Path); err == nil && savedWorkspace.Project == "" && savedWorkspace.Role == "" {
					template, found, err := c.session.TemplateHndl.Load()
					if found && err == nil {
						if template.Workspace.Project != "" && template.Workspace.Role != "" {
							savedWorkspace.Project = template.Workspace.Project
							savedWorkspace.Role = template.Workspace.Role
							if template.Workspace.Version != "" {
								savedWorkspace.Version = template.Workspace.Version
							}
							cfg.Paths[pathIndex] = savedWorkspace
							logFields := mimiclog.Fields{"path": savedWorkspace.Path, "project": savedWorkspace.Project, "role": savedWorkspace.Role}
							c.session.Log.Logger.Info("found template for workspace", logFields)
							configure.GetGlobalConfig().UpdateCurrentConfig(cfg)
							haveUpdate = true
							wsUpdated++
							if updateFn != nil {
								c.session.Log.Logger.Debug("exeute update function")
								updateFn(index, wsCount, true, savedWorkspace)
							}
						}
					} else {
						if updateFn != nil {
							updateFn(index, wsCount, false, savedWorkspace)
						}
					}
				}
			}

		})
		if haveUpdate {
			c.session.Log.Logger.Info("Update configuration")
			if err := configure.GetGlobalConfig().SaveConfiguration(); err != nil {
				c.session.Log.Logger.Error("Error while saving configuration", err)
				ctxout.CtxOut("Error while saving configuration", err)
				systools.Exit(systools.ErrorBySystem)
			}
		}
		os.Chdir(currentPath)
	}
	ctxout.PrintLn("")
	return wsCount, wsUpdated
}

func (c *CmdExecutorImpl) SetLogLevel(level string) error {
	if level != "" {
		lvl, err := logrus.ParseLevel(level)
		if err != nil {
			return err
		}
		c.session.Log.Logger.SetLevel(lvl)

	}
	return nil
}

func (c *CmdExecutorImpl) GetLogger() mimiclog.Logger {
	return c.session.Log.Logger
}

func (c *CmdExecutorImpl) PrintPaths(plain bool, showFulltask bool) {
	dir, err := os.Getwd()
	logFields := mimiclog.Fields{"dir": dir, "err": err}
	c.session.Log.Logger.Debug("print paths in workspace", logFields)

	if err == nil {
		if !plain {
			c.Println(ctxout.ForeWhite, " current directory: ", ctxout.BoldTag, dir, ctxout.CleanTag)
			c.Println(ctxout.ForeWhite, " current workspace: ", ctxout.BoldTag, configure.GetGlobalConfig().UsedV2Config.CurrentSet, ctxout.CleanTag)
		}
		pathColor := ctxout.ForeLightBlue
		if !configure.GetGlobalConfig().PathMeightPartOfWs(dir) {
			pathColor = ctxout.ForeLightMagenta
		}
		if !plain {
			c.Println(" contains paths:")
		}
		//ctxout.Print(c.session.OutPutHdnl, "<table>")
		walkErr := configure.GetGlobalConfig().PathWorker(func(index string, path string) {
			template, exists, err := c.session.TemplateHndl.Load()
			if err == nil {
				add := ctxout.Dim + ctxout.ForeLightGrey
				taskDrawMode := "ignore"
				if showFulltask {
					taskDrawMode = "wordwrap"
				}
				indexColor := ctxout.ForeLightBlue
				indexStr := index
				if path == configure.GetGlobalConfig().GetActivePath("") {
					indexColor = ctxout.ForeLightCyan
					indexStr = "> " + index
					add = ctxout.ResetDim + ctxout.ForeLightGrey
				}

				if strings.Contains(dir, path) {
					add = ctxout.ResetDim + ctxout.ForeCyan
				}
				if dir == path {
					add = ctxout.ResetDim + ctxout.ForeGreen
				}
				outTasks := ""
				if exists {
					targets, _ := TemplateTargetsAsMap(template, true)
					outTasks = strings.Join(targets, " ")
				} else {
					outTasks = ctxout.ForeDarkGrey + "no tasks"
				}
				c.Print(
					"<row>",
					indexColor,
					"<tab size='5' fill=' ' draw='fixed' origin='2'>",
					indexStr, " ",
					"</tab>",
					add,
					"<tab size='65' draw='content' fill=' ' cut-add='///..' origin='1'>",
					path, " ",
					"</tab>",
					ctxout.CleanTag,
					"<tab fill=' ' prefix='<f:yellow>' suffix='</>'  overflow='"+taskDrawMode+"' draw='extend' size='30' cut-add='<f:light-blue>...more</>' origin='2'>",
					outTasks,
					"</tab>",
					"</row>",
				)
			} else {
				c.Print(ctxout.Message("       path: ", ctxout.Dim, " no ", ctxout.ForeYellow, index, " ", pathColor, path, ctxout.ForeRed, " error while loading template: ", err.Error()))
			}
		}, func(origin string) {})

		if walkErr != nil {
			c.session.Log.Logger.Error("Error while walking through paths", err)
			c.Println(ctxout.ForeRed, "Error while walking through paths: ", ctxout.CleanTag, walkErr.Error(), ctxout.CleanTag)
		}
		//c.Println("")
	}
}

func (c *CmdExecutorImpl) GetCurrentWorkSpace() string {
	return configure.GetGlobalConfig().UsedV2Config.CurrentSet
}

func (c *CmdExecutorImpl) PrintWorkspaces() {
	configure.GetGlobalConfig().ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
		if index == configure.GetGlobalConfig().UsedV2Config.CurrentSet {
			c.Println("\t[ ", ctxout.BoldTag, index, ctxout.CleanTag, " ]")
		} else {
			c.Println("\t  ", ctxout.ForeDarkGrey, index, ctxout.CleanTag)
		}
	})
}

func TemplateTargetsAsMap(template configure.RunConfig, showInvTarget bool) ([]string, bool) {
	var targets []string
	found := false

	if len(template.Task) > 0 {
		for _, tasks := range template.Task {
			if !systools.SliceContains(targets, tasks.ID) && (!tasks.Options.Invisible || showInvTarget) {
				found = true
				targets = append(targets, strings.TrimSpace(tasks.ID))
			}
		}
	}
	sort.Strings(targets)
	return targets, found
}

func (c *CmdExecutorImpl) Lint(showAll bool) error {
	c.Println("linting...")
	c.session.TemplateHndl.SetLinting(true)
	if _, exists, err := c.session.TemplateHndl.Load(); err != nil {
		c.Println(ctxout.ForeRed, "linting failed: ", ctxout.CleanTag, err.Error())
		return err
	} else {
		if exists {
			c.Println("...loading config ", ctxout.ForeGreen, "ok", ctxout.CleanTag)
			linter, lErr := c.session.TemplateHndl.GetLinter()
			if lErr != nil {
				return lErr
			}
			if linter.HasWarning() {
				if showAll {
					c.Println(" ")
					c.Println("  you see all the unset fields, which are not set in the config")
					c.Println("  these are shown as", ctxout.ForeDarkGrey, " MissingEntry: level[5]", ctxout.CleanTag)
					c.Println("  this do not mean, that you have to set them, but it is a hint, that you can set them")
					c.Println("  and how they are named")
					c.Println(" ")
					c.Println(" ")
					// we just print all warnings once per keypath
					alreadyPrinted := make(map[string]bool)

					linter.GetIssue(yaclint.IssueLevelWarn, func(token *yaclint.MatchToken) {
						//c.Println(ctxout.ForeYellow, "linting warning: ", ctxout.CleanTag, token.ToIssueString())
						propColor := ctxout.ForeYellow
						if token.Added {
							propColor = ctxout.ForeGreen
						}

						canPrint := true
						if _, ok := alreadyPrinted[token.KeyPath]; ok {
							canPrint = false
						}
						if canPrint {
							valueStr := fmt.Sprintf(" %v ", token.Value)
							c.Println(
								ctxout.Row(
									ctxout.TD(token.KeyPath, ctxout.Size(20), ctxout.Prop(ctxout.AttrPrefix, propColor)),
									ctxout.TD(valueStr, ctxout.Size(20), ctxout.Prop(ctxout.AttrPrefix, ctxout.ForeYellow)),
									ctxout.TD(token.ToIssueString(), ctxout.Size(40), ctxout.Prop(ctxout.AttrPrefix, ctxout.ForeDarkGrey)),
									ctxout.TD(token.Type, ctxout.Size(20), ctxout.Prop(ctxout.AttrPrefix, ctxout.ForeLightCyan)),
								),
							)
							alreadyPrinted[token.KeyPath] = true
						}

					})
				} else {
					if linter.HasError() {
						c.Println(ctxout.ForeRed, "...linting errors: ", ctxout.CleanTag, len(linter.Errors()))
						c.Println(" ")
						linter.GetIssue(yaclint.IssueLevelError, func(token *yaclint.MatchToken) {
							c.Println(
								ctxout.ForeRed,
								"linting error: ", ctxout.ForeYellow, token.ToIssueString(),
								ctxout.CleanTag, " check entry: ",
								ctxout.ForeBlue, token.KeyPath, ctxout.CleanTag, ":", ctxout.ForeLightBlue, token.Value, ctxout.CleanTag)
						})
						c.Println(" ")
					} else {
						c.Println(ctxout.ForeLightGreen, "...linter findings : ", ctxout.CleanTag, len(linter.Warnings()))
						c.Println(" ")
						c.Println(ctxout.ForeLightBlue,
							"linter findings are usual and expected, because there are fields not set, the ",
							ctxout.BoldTag, "could", ctxout.CleanTag, ctxout.ForeLightBlue, " be set, but it is not necessary.")
						c.Println(ctxout.ForeLightBlue, "if you like to see all of this findings, use the flag show-issues")
						c.Println(" ")
					}
				}
			}
		} else {
			c.Println(ctxout.ForeRed, "linting failed: ", ctxout.CleanTag, "no template found")
			return errors.New("no template found")
		}
	}
	return nil

}

func (c *CmdExecutorImpl) InteractiveScreen() {

	if !systools.IsStdOutTerminal() {
		c.Print("no terminal detected")
		systools.Exit(systools.ErrorInitApp)
		return
	}
	shellRunner(c)
}
