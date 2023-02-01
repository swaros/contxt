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

func (c *CmdExecutorImpl) FindWorkspaceInfoByTemplate(updateFn func(workspace string, cnt int, update bool, info configure.WorkspaceInfoV2)) (allCount int, updatedCount int) {
	wsCount := 0
	wsUpdated := 0
	if currentPath, err := os.Getwd(); err != nil {
		ctxout.CtxOut("Error while reading current directory", err)
		systools.Exit(systools.ErrorBySystem)
	} else {
		haveUpdate := false
		configure.CfgV1.ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
			wsCount++
			for pathIndex, ws2 := range cfg.Paths {
				if err := os.Chdir(ws2.Path); err == nil && ws2.Project == "" && ws2.Role == "" {
					template, found, err := c.session.TemplateHndl.Load()
					if found && err == nil {
						if template.Workspace.Project != "" && template.Workspace.Role != "" {
							ws2.Project = template.Workspace.Project
							ws2.Role = template.Workspace.Role
							cfg.Paths[pathIndex] = ws2
							//ctxout.CtxOut(c.session.OutPutHdnl, "Found template for workspace ", index, " and set project and role to ", ws2.Project, ":", ws2.Role)
							configure.CfgV1.UpdateCurrentConfig(cfg)
							haveUpdate = true
							wsUpdated++
							if updateFn != nil {
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
			configure.CfgV1.SaveConfiguration()
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

func (c *CmdExecutorImpl) PrintPaths() {
	dir, err := os.Getwd()
	if err == nil {
		ctxout.CtxOut(c.session.OutPutHdnl, ctxout.ForeWhite, " current directory: ", ctxout.BoldTag, dir, ctxout.CleanTag)
		ctxout.CtxOut(c.session.OutPutHdnl, ctxout.ForeWhite, " current workspace: ", ctxout.BoldTag, configure.CfgV1.UsedV2Config.CurrentSet, ctxout.CleanTag)
		notWorkspace := true
		pathColor := ctxout.ForeLightBlue
		if !configure.CfgV1.PathMeightPartOfWs(dir) {
			pathColor = ctxout.ForeLightMagenta
		} else {
			notWorkspace = false
		}
		ctxout.CtxOut(c.session.OutPutHdnl, " contains paths:")
		ctxout.CtxOut(c.session.OutPutHdnl, "<table>")
		configure.CfgV1.PathWorker(func(index string, path string) {
			template, exists, err := c.session.TemplateHndl.Load()
			if err == nil {
				add := ctxout.ResetDim + ctxout.ForeLightMagenta
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
					ctxout.ForeLightBlue,
					"<tab size='5' fill=' ' draw='fixed' origin='2'>",
					index, " ",
					"</tab>",
					add,
					"<tab size='65' draw='content' fill=' ' cut-add='///..' origin='1'>",
					path, " ",
					"</tab>",
					ctxout.ForeYellow,
					"<tab fill=' ' draw='extend' size='30' cut-add='<f:light-blue>...more</>' origin='2'>",
					outTasks,
					"</tab>",
					ctxout.CleanTag,
					"</row>",
				)
			} else {
				ctxout.CtxOut(c.session.OutPutHdnl, ctxout.Message("       path: ", ctxout.Dim, " no ", ctxout.ForeYellow, index, " ", pathColor, path, ctxout.ForeRed, " error while loading template: ", err.Error()))
			}
		}, func(origin string) {})
		if notWorkspace {
			ctxout.CtxOut(c.session.OutPutHdnl, "</table>")
			ctxout.CtxOut(c.session.OutPutHdnl, "\n")
			ctxout.CtxOut(c.session.OutPutHdnl, ctxout.BackYellow, ctxout.ForeBlue, " WARNING ! ", ctxout.CleanTag, "\tyou are currently in none of the assigned locations.")
			ctxout.CtxOut(c.session.OutPutHdnl, "\t\tso maybe you are using the wrong workspace")
		} else {
			ctxout.CtxOut(c.session.OutPutHdnl, "</table>")
		}

		ctxout.CtxOut(c.session.OutPutHdnl, "\n")

		ctxout.CtxOut(c.session.OutPutHdnl, " all workspaces:")

		configure.CfgV1.ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
			if index == configure.CfgV1.UsedV2Config.CurrentSet {
				ctxout.CtxOut(c.session.OutPutHdnl, "\t[ ", ctxout.BoldTag, index, ctxout.CleanTag, " ]")
			} else {
				ctxout.CtxOut(c.session.OutPutHdnl, "\t  ", ctxout.ForeDarkGrey, index, ctxout.CleanTag)
			}
		})
	}
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
