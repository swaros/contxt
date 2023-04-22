package runner

import (
	"os"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
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

func (c *CmdExecutorImpl) doMagicParamOne(args string) {
}

func (c *CmdExecutorImpl) CallBackOldWs(oldws string) bool {
	c.session.Log.Logger.Info("OLD workspace: ", oldws)
	// get all paths first
	configure.CfgV1.PathWorkerNoCd(func(_ string, path string) {

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
	configure.CfgV1.PathWorker(func(_ string, path string) { // iterate any path
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

func (c *CmdExecutorImpl) GetOuputHandler() ctxout.PrintInterface {
	return c.session.OutPutHdnl
}

func (c *CmdExecutorImpl) GetWorkspaces() []string {
	ws := configure.CfgV1.ListWorkSpaces()
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
		configure.CfgV1.ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
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
							configure.CfgV1.UpdateCurrentConfig(cfg)
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
			if err := configure.CfgV1.SaveConfiguration(); err != nil {
				c.session.Log.Logger.WithFields(logrus.Fields{"err": err}).Error("Error while saving configuration")
				ctxout.CtxOut("Error while saving configuration", err)
				systools.Exit(systools.ErrorBySystem)
			}
		}
		os.Chdir(currentPath)
	}
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
			ctxout.CtxOut(c.session.OutPutHdnl, ctxout.ForeWhite, " current directory: ", ctxout.BoldTag, dir, ctxout.CleanTag)
			ctxout.CtxOut(c.session.OutPutHdnl, ctxout.ForeWhite, " current workspace: ", ctxout.BoldTag, configure.CfgV1.UsedV2Config.CurrentSet, ctxout.CleanTag)
		}
		notWorkspace := true
		pathColor := ctxout.ForeLightBlue
		if !configure.CfgV1.PathMeightPartOfWs(dir) {
			pathColor = ctxout.ForeLightMagenta
		} else {
			notWorkspace = false
		}
		if !plain {
			ctxout.CtxOut(c.session.OutPutHdnl, " contains paths:")
		}
		//ctxout.Print(c.session.OutPutHdnl, "<table>")
		configure.CfgV1.PathWorker(func(index string, path string) {
			template, exists, err := c.session.TemplateHndl.Load()
			if err == nil {
				add := ctxout.Dim + ctxout.ForeLightGrey
				taskDrawMode := "ignore"
				if showFulltask {
					taskDrawMode = "wordwrap"
				}
				indexColor := ctxout.ForeLightBlue
				indexStr := index
				if path == configure.CfgV1.GetActivePath("") {
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
				ctxout.Print(
					c.session.OutPutHdnl,
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
				ctxout.Print(c.session.OutPutHdnl, ctxout.Message("       path: ", ctxout.Dim, " no ", ctxout.ForeYellow, index, " ", pathColor, path, ctxout.ForeRed, " error while loading template: ", err.Error()))
			}
		}, func(origin string) {})
		if notWorkspace && !plain {

			ctxout.PrintLn(c.session.OutPutHdnl, "<row><tab size='20' origin='2'>", ctxout.ForeYellow, " WARNING ! </tab>", ctxout.CleanTag, "<tab size='80'>you are currently in none of the assigned locations.<tab></row>")
			ctxout.PrintLn(c.session.OutPutHdnl, "<row><tab size='20'> </tab><tab=size='80'>so maybe you are using the wrong workspace</tab></row>")
		}

	}
}

func (c *CmdExecutorImpl) PrintWorkspaces() {
	configure.CfgV1.ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
		if index == configure.CfgV1.UsedV2Config.CurrentSet {
			ctxout.CtxOut(c.session.OutPutHdnl, "\t[ ", ctxout.BoldTag, index, ctxout.CleanTag, " ]")
		} else {
			ctxout.CtxOut(c.session.OutPutHdnl, "\t  ", ctxout.ForeDarkGrey, index, ctxout.CleanTag)
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
	tc := initTcellScreen(c)
	tc.Run()
}

func initTcellScreen(c *CmdExecutorImpl) *ctCell {
	tc := NewTcell()
	tc.SetMouse(true).SetNoClearScreen(false)
	// then first submenu
	menu := tc.NewMenu()

	// top bar
	contxtTopMenu := tc.ActiveText("contxt")
	contxtTopMenu.SetPos(1, 0).SetStyle(tcell.StyleDefault.Foreground(tcell.ColorGoldenrod).Background(tcell.ColorBlack))
	contxtTopMenu.OnSelect = func(selected bool) {
		menu.SetVisible(!menu.IsVisible())
	}
	tc.AddElement(contxtTopMenu)

	exitTopMenu := tc.ActiveText("exit")
	exitTopMenu.SetPosProcentage(100, 0).
		SetStyle(tcell.StyleDefault.Foreground(tcell.ColorGoldenrod).Background(tcell.ColorBlack))

	exitTopMenu.GetPos().SetMargin(0, 0, -5, 0)
	exitTopMenu.OnSelect = func(selected bool) {
		menu.SetVisible(false)
	}
	tc.AddElement(exitTopMenu)

	menu.SetTopLeft(1, 1).SetBottomRight(20, 10)
	menu.AddItem("PrintPaths", func(itm *MenuElement) {
		itm.text.text = "PrintPaths RUNS"
		c.PrintPaths(false, false)
		itm.text.text = "PrintPaths done"
	})

	menu.SetVisible(false)
	tc.AddElement(menu)

	return tc

}
