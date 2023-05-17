package runner

import (
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/systools"
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

		os.Chdir(path)
		template, exists, _ := c.session.TemplateHndl.Load()

		c.session.Log.Logger.WithFields(logrus.Fields{
			"exists": exists,
			"path":   path,
		}).Debug("path parsing")

		if exists && template.Config.Autorun.Onleave != "" {
			onleaveTarget := template.Config.Autorun.Onleave
			c.session.Log.Logger.WithFields(logrus.Fields{
				"target": onleaveTarget,
			}).Info("execute leave-action")
			c.RunTargets(onleaveTarget, true)

		}

	})
	return true
}

func (c *CmdExecutorImpl) CallBackNewWs(newWs string) {
	c.ResetVariables() // reset old variables while change the workspace. (req for shell mode)
	c.MainInit()       // initialize the workspace
	c.session.Log.Logger.Info("NEW workspace: ", newWs)
	configure.GetGlobalConfig().PathWorker(func(_ string, path string) { // iterate any path
		template, exists, _ := c.session.TemplateHndl.Load()

		c.session.Log.Logger.WithFields(logrus.Fields{

			"exists": exists,
			"path":   path,
		}).Debug("path parsing")

		// try to run onEnter func at any possible target in the workspace
		if exists && template.Config.Autorun.Onenter != "" {
			onEnterTarget := template.Config.Autorun.Onenter
			c.session.Log.Logger.WithFields(logrus.Fields{
				"target": onEnterTarget,
			}).Info("execute enter-action")
			c.RunTargets(onEnterTarget, true)
		}

	}, func(origin string) {
		c.session.Log.Logger.WithFields(logrus.Fields{
			"current-dir": origin,
		}).Debug("done calling autoruns on sub-dirs")
	})
}

func (c *CmdExecutorImpl) RunTargets(target string, force bool) {
}

func (c *CmdExecutorImpl) ResetVariables() {
}

func (c *CmdExecutorImpl) MainInit() {
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
			for pathIndex, ws2 := range cfg.Paths {
				c.session.Log.Logger.WithFields(logrus.Fields{"path": ws2.Path, "project": ws2.Project, "role": ws2.Role}).Debug("parsing workspace")
				if err := os.Chdir(ws2.Path); err == nil && ws2.Project == "" && ws2.Role == "" {
					template, found, err := c.session.TemplateHndl.Load()
					if found && err == nil {
						if template.Workspace.Project != "" && template.Workspace.Role != "" {
							ws2.Project = template.Workspace.Project
							ws2.Role = template.Workspace.Role
							cfg.Paths[pathIndex] = ws2
							c.session.Log.Logger.WithFields(logrus.Fields{"path": ws2.Path, "project": ws2.Project, "role": ws2.Role}).Info("found template for workspace")
							configure.GetGlobalConfig().UpdateCurrentConfig(cfg)
							haveUpdate = true
							wsUpdated++
							if updateFn != nil {
								c.session.Log.Logger.WithFields(logrus.Fields{"path": ws2.Path, "project": ws2.Project, "role": ws2.Role}).Debug("exeute update function")
								updateFn(index, wsCount, true, ws2)
							}
						}
					} else {
						if updateFn != nil {
							updateFn(index, wsCount, false, ws2)
						}
					}
				}
			}

		})
		if haveUpdate {
			c.session.Log.Logger.Info("Update configuration")
			if err := configure.GetGlobalConfig().SaveConfiguration(); err != nil {
				c.session.Log.Logger.WithFields(logrus.Fields{"err": err}).Error("Error while saving configuration")
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

func (c *CmdExecutorImpl) GetLogger() *logrus.Logger {
	return c.session.Log.Logger
}

func (c *CmdExecutorImpl) PrintPaths(plain bool, showFulltask bool) {
	dir, err := os.Getwd()
	c.session.Log.Logger.WithFields(logrus.Fields{
		"dir": dir,
		"err": err,
	}).Debug("print paths in workspace")

	if err == nil {
		if !plain {
			c.Print(ctxout.ForeWhite, " current directory: ", ctxout.BoldTag, dir, ctxout.CleanTag)
			c.Print(ctxout.ForeWhite, " current workspace: ", ctxout.BoldTag, configure.GetGlobalConfig().UsedV2Config.CurrentSet, ctxout.CleanTag)
		}
		notWorkspace := true
		pathColor := ctxout.ForeLightBlue
		if !configure.GetGlobalConfig().PathMeightPartOfWs(dir) {
			pathColor = ctxout.ForeLightMagenta
		} else {
			notWorkspace = false
		}
		if !plain {
			c.Println(" contains paths:")
		}
		//ctxout.Print(c.session.OutPutHdnl, "<table>")
		configure.GetGlobalConfig().PathWorker(func(index string, path string) {
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
		if notWorkspace && !plain {

			c.Println("<row><tab size='20' origin='2'>", ctxout.ForeYellow, " WARNING ! </tab>", ctxout.CleanTag, "<tab size='80'>you are currently in none of the assigned locations.<tab></row>")
			c.Println("<row><tab size='20'> </tab><tab=size='80'>so maybe you are using the wrong workspace</tab></row>")
		}
		// end table
		// print a new line
		c.Println("")
	}
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

func (c *CmdExecutorImpl) InteractiveScreen() {

	if !systools.IsStdOutTerminal() {
		c.Print("no terminal detected")
		systools.Exit(systools.ErrorInitApp)
		return
	}
	shellRunner(c)
}
