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
	"gopkg.in/yaml.v2"
)

type CmdExecutorImpl struct {
	session     *CmdSession
	executer    *tasks.TaskListExec
	dataHandl   *tasks.CombinedDh
	outHandlers map[string]*OutputHandler
	usedHandler string
}

func NewCmd(session *CmdSession) *CmdExecutorImpl {
	return &CmdExecutorImpl{
		session:     session,
		outHandlers: make(map[string]*OutputHandler),
	}
}

func (c *CmdExecutorImpl) SetOutputHandlerByName(name string) error {
	c.usedHandler = name
	return nil // right now we only set the name of the handler. so this will not fail. it may fail in the future
}

func (c *CmdExecutorImpl) PrintVariables(format string) {
	format = strings.TrimSpace(format)
	format = strings.ReplaceAll(format, "[nl]", "\n")
	checkOdd := 0
	iterMap := systools.StrStr2StrAny(c.GetVariables())
	systools.MapRangeSortedFn(iterMap, func(key string, value any) {
		checkOdd++
		if strings.Contains(format, "%") {
			c.Print(fmt.Sprintf(format, key, value))
		} else {
			fColorLeft := ctxout.ForeWhite
			fColorRight := ctxout.ForeWhite
			backCol := ctxout.BackBlue
			if checkOdd%2 == 0 {
				fColorLeft = ctxout.ForeBlack
				fColorRight = ctxout.ForeBlue
				backCol = ctxout.BackLightBlue
			}
			ctxout.PrintLn(
				c.session.OutPutHdnl,
				c.session.Printer,
				ctxout.Row(
					ctxout.TD(
						key,
						ctxout.Prop(ctxout.AttrSize, 20),
						ctxout.Prop(ctxout.AttrOrigin, ctxout.OriginRight),
						ctxout.Prop(ctxout.AttrPrefix, fColorLeft+backCol),
						ctxout.Prop(ctxout.AttrSuffix, ctxout.ResetCode),
					),
					ctxout.TD(
						" -> ",
						ctxout.Prop(ctxout.AttrSize, 2),
						ctxout.Prop(ctxout.AttrOrigin, ctxout.OriginRight),
						ctxout.Prop(ctxout.AttrPrefix, ctxout.ForeLightCyan+backCol),
						ctxout.Prop(ctxout.AttrSuffix, ctxout.ResetCode),
					),
					ctxout.TD(
						value,
						ctxout.Prop(ctxout.AttrSize, 70),
						ctxout.Prop(ctxout.AttrOrigin, ctxout.OriginLeft),
						ctxout.Prop(ctxout.AttrPrefix, fColorRight+backCol),
						ctxout.Prop(ctxout.AttrSuffix, ctxout.ResetCode),
					),
				),
			)
		}
	})

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
			c.InitExecuter()
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
			c.InitExecuter()
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
	// first apply logger if poosible
	mimiclog.ApplyLogger(c.session.Log.Logger, dataHndl)

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
		dataHndl.SetPH("BASEPATH", currentDir)
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
		}
	})
	c.session.Log.Logger.Debug("set startup variables for ws2", keys)
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
	c.session.Log.Logger.Info("handle imports")
	importHndlr := NewImportHandler(c.session.Log.Logger, dataHndl, c.session.TemplateHndl)
	importHndlr.SetImports(template.Config.Imports)
	if err := importHndlr.HandleImports(); err != nil {
		c.Println(ctxout.ForeRed, "error while handling imports", ctxout.ForeYellow, err)
		c.session.Log.Logger.Error("error while handling imports", err)
		systools.Exit(systools.ErrorBySystem)
	}
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

func (c *CmdExecutorImpl) AddIncludePath(path string) error {
	if path == "" {
		return errors.New("empty path is not allowed")
	}
	fileContent, err := AddPathToIncludeImports(c.session.TemplateHndl.GetIncludeConfig(), path)
	if err != nil {
		return err
	}
	if fileContent == "" {
		return errors.New("internal error. failed by creating the include file content")
	}
	return systools.WriteFile(c.session.TemplateHndl.GetIncludeFile(), fileContent)
}

func (c *CmdExecutorImpl) CreateContxtFile() error {
	// Define the content of the file
	return CreateContxtFile()
}

func (c *CmdExecutorImpl) runAsyncTargets(targets []string, force bool) error {
	// using channel to sync the goroutines
	ch := make(chan bool)
	// run the targets in goroutines
	c.InitExecuter()
	for _, target := range targets {
		go func(t string) {
			c.RunTargets(t, force)
			ch <- true
		}(target)
	}
	// wait for all goroutines to finish
	for range targets {
		<-ch
	}
	return nil
}

func (c *CmdExecutorImpl) setDefaultOutHandlers() {
	c.addOutHandler(NewTableOutput())
	c.addOutHandler(NewPlainOutput())
}

func (c *CmdExecutorImpl) InitExecuter() error {
	if template, exists, err := c.session.TemplateHndl.Load(); err != nil {
		c.session.Log.Logger.Error("error while loading template", err)
		c.tryExplainError(err)
		return err
	} else if !exists {
		c.session.Log.Logger.Error("template not exists")
		return errors.New("no contxt template found in current directory")
	} else {

		c.dataHandl = tasks.NewCombinedDataHandler()
		c.SetStartupVariables(c.dataHandl, &template)

		c.setDefaultOutHandlers() // register any outputhandler
		outputHndl, err := c.setOutHandler(c.usedHandler)
		if err != nil {
			return err
		}

		requireHndl := tasks.NewDefaultRequires(c.dataHandl, c.session.Log.Logger)
		c.executer = tasks.NewTaskListExec(
			template,
			c.dataHandl,
			requireHndl,
			outputHndl,
			tasks.ShellCmd,
		)
		c.executer.SetLogger(c.session.Log.Logger)
	}
	return nil
}

// RunTargets run the given targets
// force is used as flag for the first level targets, and is used
// to runs shared targets once in front of the regular assigned targets
func (c *CmdExecutorImpl) RunTargets(target string, force bool) error {

	// the executer needs to be initialized first
	if c.executer == nil {
		return errors.New("executer not initialized")
	}

	// the datahandler needs to be initialized first
	if c.dataHandl == nil {
		return errors.New("datahandler not initialized")
	}

	// first we need to check if there is a commas separated list of targets
	// if so, we need to split them and run them one by one
	if strings.Contains(target, ",") {
		targets := strings.Split(target, ",")
		return c.runAsyncTargets(targets, force)
	}

	c.dataHandl.SetPH("CTX_TARGET", target)
	c.dataHandl.SetPH("CTX_FORCE", strconv.FormatBool(force))

	c.executer.SetLogger(c.session.Log.Logger)
	code := c.executer.RunTarget(target, force)
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
	case systools.ExitByUnsupportedVersion:
		c.session.Log.Logger.Error("unsupported version")
		return errors.New("unsupported version")
	default:
		c.session.Log.Logger.Error("unexpected exit code:", code)
		return errors.New("unexpected exit code:" + fmt.Sprintf("%d", code))
	}

}

func (c *CmdExecutorImpl) GetTargets(incInvisible bool) []string {
	if template, exists, err := c.session.TemplateHndl.Load(); err != nil {
		c.session.Log.Logger.Error("error while loading template", err)
		c.tryExplainError(err)
	} else if !exists {
		c.session.Log.Logger.Debug("template not exists", err)
	} else {
		if res, have := TemplateTargetsAsMap(template, incInvisible); have {
			return res
		}
	}
	return nil
}

func (c *CmdExecutorImpl) ResetVariables() {
}

func (c *CmdExecutorImpl) MainInit() error {
	return c.initDefaultVariables()
}

// initDefaultVariables init the default variables for the current session.
// these are the varibales they should not change during the session.
func (c *CmdExecutorImpl) initDefaultVariables() error {
	if currentPath, err := os.Getwd(); err != nil {
		ctxout.CtxOut("Error while reading current directory", err)
		systools.Exit(systools.ErrorBySystem)
	} else {
		c.setVariable("CTX_PWD", currentPath)
		c.setVariable("CTX_PATH", currentPath)
		c.setVariable("BASEPATH", currentPath)
	}
	exec, err := os.Executable()
	if err != nil {
		c.session.Log.Logger.Error("error while getting executable", err)
	} else {
		c.setVariable("CTX_BINARY", exec)
	}
	c.setVariable("CTX_BIN_NAME", configure.GetBinaryName())
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

	if err := c.SetProjectVariables(); err != nil {
		return err
	}

	c.handleWindowsInit() // it self is testing if we are on windows
	return nil
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
func (c *CmdExecutorImpl) setVariable(name string, value string) error {
	if name, err := systools.CheckForCleanString(name); err == nil {
		c.session.DefaultVariables[name] = value
		return nil
	} else {
		c.session.Log.Logger.Error("error while setting variable", err)
		return err
	}
}

func (c *CmdExecutorImpl) GetVariable(name string) string {

	if c.dataHandl != nil {
		if val, have := c.dataHandl.GetPHExists(name); have {
			return val
		}
	}

	if val, have := c.session.DefaultVariables[name]; have {
		return val
	}
	return ""
}

func (c *CmdExecutorImpl) GetVariables() map[string]string {

	returns := c.session.DefaultVariables
	if c.dataHandl != nil {
		c.dataHandl.GetPlaceHoldersFnc(func(key string, value string) {
			returns[key] = value
		})
	}
	return returns
}

func (c *CmdExecutorImpl) SetColor(onoff bool) {
	behave := ctxout.GetBehavior()
	behave.NoColored = onoff
	ctxout.SetBehavior(behave)
}

func (c *CmdExecutorImpl) GetOuputHandler() (ctxout.StreamInterface, ctxout.PrintInterface) {
	return c.session.OutPutHdnl, c.session.Printer
}

// tryExplainError try to explain the error by parsing the error message
// and give a hint to the user, what could be the reason for the error.
// if possible it shows also the source code, where the error is located.
func (c *CmdExecutorImpl) tryExplainError(err error) {
	if err != nil {
		errExplain := NewErrParse(err, c.session)
		c.Println(ctxout.ForeYellow, "error explanation: ", ctxout.ForeLightBlue, errExplain.Explain(), ctxout.CleanTag)
		if errExplain.code != nil {
			for _, code := range errExplain.code {
				codeColor := ctxout.ForeBlue + ctxout.BackWhite
				errMsg := ""
				if code.IsError {
					codeColor = ctxout.ForeRed + ctxout.BackLightYellow
					errMsg = "« " + errExplain.Explain()
				}
				c.Println(ctxout.Row(
					ctxout.TD(code.LineNr, ctxout.Fixed(), ctxout.Prop(ctxout.AttrPrefix, ctxout.ForeDarkGrey), ctxout.Right(), ctxout.Size(5)),
					ctxout.TD("|", ctxout.Fixed(), ctxout.Prop(ctxout.AttrPrefix, ctxout.ForeDarkGrey), ctxout.Right(), ctxout.Size(1)),
					ctxout.TD(code.Line, ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag), ctxout.Prop(ctxout.AttrPrefix, codeColor), ctxout.Left(), ctxout.Size(55)),
					ctxout.TD(errMsg, ctxout.Prop(ctxout.AttrPrefix, ctxout.ForeLightYellow), ctxout.Left(), ctxout.Size(30)),
				))
			}
			msq := `
			by investigating the source code, keep in mind that the line numbers
			could be different, since the error is shown in the parsed code and 
			the line count could be changed by the parser.
			`
			c.Println(ctxout.ForeBlue, msq, ctxout.CleanTag)
		} else {
			c.Println(ctxout.ForeDarkGrey, "code could not be shown. Why should be shown in the error above.", ctxout.CleanTag)
		}

	}
}

// SetProjectVariables set the project variables for the current session.
// what includes any workspaces and there paths.
// we ignore any errors, since we are not able to do anything with them.
// this would just be a log entry and ignored as variable.
func (c *CmdExecutorImpl) SetProjectVariables() error {
	if template, exists, err := c.session.TemplateHndl.Load(); err != nil {
		c.session.Log.Logger.Error("error while loading template for Setting Default Vatiables", err)
		c.tryExplainError(err)
		return err
	} else if !exists {
		c.session.Log.Logger.Debug("template not exists", err)
	} else {
		c.setVariable("WS_PROJECT", template.Workspace.Project)
		c.setVariable("WS_ROLE", template.Workspace.Role)
		c.setVariable("WS_VERSION", template.Workspace.Version)
	}

	configure.GetGlobalConfig().ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
		for key, ws := range cfg.Paths {
			upperIndex := strings.ToUpper(index)
			if ws.Role != "" {
				c.setVariable("WS_PATH_"+upperIndex+"_"+strings.ToUpper(ws.Role), ws.Path)
			} else {
				c.setVariable("WS_PATH_"+upperIndex+"_"+key, ws.Path)
			}
			if ws.Version != "" {
				c.setVariable("WS_VERSION_"+upperIndex, ws.Version)
			}

		}
	})
	return nil
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
					indexStr+" ",
					"</tab>",
					add,
					"<tab size='65' draw='content' fill=' ' cut-add='///..' origin='1'>",
					path, " ",
					"</tab>",
					ctxout.CleanTag,
					"<tab size='29' fill=' ' prefix='<f:yellow>' suffix='</>'  overflow='"+taskDrawMode+"' draw='extend' cut-add='<f:light-blue> ..<f:yellow>.' origin='2'>",
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
		c.tryExplainError(err)
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
	shellRunner(c).runAsShell()
}
func (c *CmdExecutorImpl) ShellWithComands(cmds []string, timeout int) {
	if err := shellRunner(c).runWithCmds(cmds, timeout); err != nil {
		c.Println(ctxout.ForeRed, "error while running shell", ctxout.CleanTag, err.Error())
	}
}

// PrintShared print all shared paths in a simple list
func (c *CmdExecutorImpl) PrintShared() {
	sharedRun := NewSharedHelper()
	sharedDirs, _ := sharedRun.ListUseCases(false)
	for _, sharedPath := range sharedDirs {
		c.Println(sharedPath)
	}
}

// displays the current version of contxt template as a yaml string
func (c *CmdExecutorImpl) PrintTemplate() {
	if template, exists, err := c.session.TemplateHndl.Load(); err != nil {
		c.tryExplainError(err)
		c.Println(ctxout.ForeRed, "yaml export failed: ", ctxout.CleanTag, err.Error())
		c.Print(ctxout.ForeRed, "error while loading template: ", ctxout.CleanTag, err.Error())
	} else {
		if exists {
			c.Println("...loading config ", ctxout.ForeGreen, "ok", ctxout.CleanTag)
			// map the template to a yaml string
			if yamlStr, err := yaml.Marshal(template); err != nil {
				c.Println(ctxout.ForeRed, "yaml export failed: ", ctxout.CleanTag, err.Error())
			} else {
				c.Println(string(yamlStr))
			}
		} else {
			c.Println(ctxout.ForeRed, "yaml export failed: ", ctxout.CleanTag, "no template found")
		}
	}
}

// Set the given variable to the current session Default Variables.
// this will end up by using them as variables for the template, and are reset for any run.
// this is also an different Behavior to V1 where the variables are set for the wohle runtime, and if
// they changed by a task, they are changed for the whole runtime.
// this is not happen anymore, and the variables are just set for the current run.
func (c *CmdExecutorImpl) SetPreValue(name string, value string) {
	c.setVariable(name, value)
}

func (c *CmdExecutorImpl) RunAnkoScript(args []string) error {
	runner := tasks.NewAnkoRunner()
	runner.SetLogger(c.session.Log.Logger)

	script := strings.Join(args, "\n")
	_, err := runner.RunAnko(script)
	return err
}
