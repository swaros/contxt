package tasks

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/awaitgroup"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/manout"
)

type taskListExec struct {
	config   configure.RunConfig
	watch    *Watchman
	subTasks map[string]*targetExecuter
	args     []interface{}
}

func NewTaskListExec(config configure.RunConfig, adds ...interface{}) *taskListExec {
	return &taskListExec{
		config: config,
		watch:  NewWatchman(),
		args:   adds,
	}
}

func (e *taskListExec) RunTarget(target string, async bool) int {
	scopeVars := make(map[string]string)
	return e.RunTargetWithVars(target, scopeVars, async)
}

func (e *taskListExec) RunTargetWithVars(target string, scopeVars map[string]string, async bool) int {
	tExec := e.findOrCreateTask(target, scopeVars)
	return tExec.executeTemplate(async, target, scopeVars)
}

func (e *taskListExec) GetTask(target string) *targetExecuter {
	if e.subTasks == nil {
		e.subTasks = make(map[string]*targetExecuter)
		return nil
	}
	if tExec, found := e.subTasks[target]; found {
		return tExec
	}
	return nil
}

func (e *taskListExec) GetWatch() *Watchman {
	return e.watch
}

/*
	func (e *taskListExec) SetTask(tExec *targetExecuter) {
		if e.subTasks == nil {
			e.subTasks = make(map[string]*targetExecuter)
		}
		e.subTasks[tExec.target] = tExec
	}
*/
func (e *taskListExec) findOrCreateTask(target string, scopeVars map[string]string) *targetExecuter {
	if e.subTasks == nil {
		e.subTasks = make(map[string]*targetExecuter)
	}
	tExec, found := e.subTasks[target]
	if !found {
		for _, task := range e.config.Task {
			if task.ID == target {
				e.args = append(e.args, e.config) // add the config to the args
				tExec = New(target, scopeVars, e.args...)
				e.subTasks[target] = tExec
			}
		}
	}
	return tExec
}

func (t *targetExecuter) executeTemplate(runAsync bool, target string, scopeVars map[string]string) int {

	// check if task is already running
	// this check depends on the target name.
	if !t.runCfg.Config.AllowMutliRun && t.watch.TaskRunning(target) {
		t.getLogger().WithField("task", target).Warning("task would be triggered again while is already running. IGNORED")
		return systools.ExitAlreadyRunning
	}

	// increment task counter
	t.watch.incTaskCount(target)
	defer t.watch.incTaskDoneCount(target) // save done count at then end

	t.getLogger().WithFields(logrus.Fields{
		"target": target,
	}).Info("executeTemplate LOOKING for target")

	// Checking if the Tasklist have something
	// to handle
	if len(t.runCfg.Task) > 0 {
		returnCode := systools.ExitOk

		// the main variables will be set at first
		// but only if the they not already exists
		// from other task or by start argument
		if t.phHandler != nil {
			for keyName, variable := range t.runCfg.Config.Variables {
				t.phHandler.SetIfNotExists(keyName, t.phHandler.HandlePlaceHolder(variable))
			}
		}
		// set the colorcodes for the labels on left side of screen
		//colorCode, bgCode := systools.CreateColorCode()

		// updates global variables
		t.setPh("RUN.TARGET", target)

		// this flag is only used
		// for a "target not found" message later
		targetFound := false

		// this flag is used to check if the target
		// was executed at least once
		targetExecuted := false

		// oure tasklist that will use later
		var taskList []configure.Task
		for _, script := range t.runCfg.Task {
			if strings.EqualFold(target, script.ID) {
				taskList = append(taskList, script)
			}
		}

		// depending on the config
		// we merge the tasks and handle them as one task,
		// or we keep them as a list of tasks what would
		// keep more flexibility.
		// by merging task we can loose runtime definitions
		/*
			if runCfg.Config.MergeTasks {
				mergedScript := mergeTargets(target, runCfg)
				taskList = append(taskList, mergedScript)
			} else {
				for _, script := range runCfg.Task {
					if strings.EqualFold(target, script.ID) {
						taskList = append(taskList, script)
					}
				}
			}*/

		// check if we have found the target
		for curTIndex, script := range taskList {

			t.getLogger().WithFields(logrus.Fields{
				"target":    target,
				"scopeVars": scopeVars,
			}).Info("executeTemplate EXECUTE target")
			targetFound = true

			//stopReason := script.Stopreasons

			var messageCmdCtrl TaskOutCtrl = TaskOutCtrl{ // define a controll hook, depending on the display comand option
				IgnoreCase: !script.Options.Displaycmd, // we ignore the message, as long the display command is NOT set
			}

			// check requirements
			canRun, message := t.checkRequirements(script.Requires)
			if !canRun {
				t.getLogger().WithFields(logrus.Fields{
					"target": target,
					"reason": message,
				}).Info("executeTemplate IGNORE because requirements not matching")
				if script.Options.Displaycmd {
					ctxout.CtxOut(messageCmdCtrl, ctxout.LabelFY("require"), ctxout.ValF(message), ctxout.InfoF("Task-Block "), curTIndex+1, " of ", len(taskList), " skipped")
				}
				// ---- return ExitByRequirement
				continue
			}
			// at least one target was executed. this menas not all targets
			// and it is not necessary to run script lines
			targetExecuted = true

			// get the task related variables
			if t.phHandler != nil {
				for keyName, variable := range script.Variables {
					t.setPh(keyName, t.phHandler.HandlePlaceHolder(variable))
					scopeVars[keyName] = variable
				}
			}
			backToDir := ""
			// if working dir is set change to them
			if script.Options.WorkingDir != "" {
				backToDir, _ = dirhandle.Current()
				wDir := script.Options.WorkingDir
				if t.phHandler != nil {
					wDir = t.phHandler.HandlePlaceHolderWithScope(script.Options.WorkingDir, scopeVars)
				}
				chDirError := os.Chdir(wDir)
				if chDirError != nil {
					manout.Error("Workspace setting seems invalid ", chDirError)
					systools.Exit(systools.ErrorBySystem)
				}
			}

			// just the abort flag.
			abort := false

			// experimental usage of needs

			// -- NEEDS
			// needs are task, the have to be startet once
			// before we continue.
			// any need can have his own needs they needs to
			// be executed
			if len(script.Needs) > 0 {

				ctxout.CtxOut(messageCmdCtrl, ctxout.LabelFY("target"), ctxout.ValF(target), ctxout.InfoF("require"), ctxout.ValF(len(script.Needs)), ctxout.InfoF("needs. async?"), ctxout.ValF(runAsync))
				t.getLogger().WithField("needs", script.Needs).Debug("Needs for the script")
				if runAsync {
					var needExecs []awaitgroup.FutureStack
					for _, needTarget := range script.Needs {
						if t.watch.TaskRunsAtLeast(needTarget, 1) {
							ctxout.CtxOut(messageCmdCtrl, ctxout.LabelFY("need check"), ctxout.ValF(target), ctxout.InfoRed("already executed"), ctxout.ValF(needTarget))
							t.getLogger().Debug("need already handled " + needTarget)
						} else {
							t.getLogger().Debug("need name should be added " + needTarget)
							ctxout.CtxOut(messageCmdCtrl, ctxout.LabelFY("need check"), ctxout.ValF(target), ctxout.InfoF("executing"), ctxout.ValF(needTarget))
							needExecs = append(needExecs, awaitgroup.FutureStack{
								AwaitFunc: func(ctx context.Context) interface{} {
									argNeed := ctx.Value(awaitgroup.CtxKey{}).(string)
									_, argmap := systools.StringSplitArgs(argNeed, "arg")
									t.getLogger().Debug("add need task " + argNeed)
									return t.executeTemplate(true, argNeed, argmap)
								},
								Argument: needTarget})
						}
					}
					futures := awaitgroup.ExecFutureGroup(needExecs) // create the futures and start the tasks
					results := awaitgroup.WaitAtGroup(futures)       // wait until any task is executed

					t.getLogger().WithField("result", results).Debug("needs result")
				} else {
					for _, needTarget := range script.Needs {
						if t.watch.TaskRunsAtLeast(needTarget, 1) { // do not run needs the already runs
							t.getLogger().Debug("need already handled " + needTarget)
						} else {
							_, argmap := systools.StringSplitArgs(needTarget, "arg")
							t.executeTemplate(false, needTarget, argmap)
						}
					}
				}
				ctxout.CtxOut(ctxout.LabelFY("target"), ctxout.ValF(target), ctxout.InfoF("needs done"))
			}

			// targets that should be started as well
			// these targets running at the same time
			// so different to scope, we dont need to wait
			// right now until they ends
			runTargetfutures := t.generateFuturesByTargetListAndExec(script.RunTargets, t.runCfg)

			// check if we have script lines.
			// if not, we need at least to check
			// 'now' listener
			if len(script.Script) < 1 {
				t.getLogger().Debug("no script lines defined. run listener anyway")
				t.listenerWatch("", nil, &script)
				// workaround til the async runnig is refactored
				// now we need to give the subtask time to run and update the waitgroup
				duration := time.Second
				time.Sleep(duration)
			}

			// preparing codelines by execute second level commands
			// that can affect the whole script
			abort, returnCode, _ = t.TryParse(script.Script, func(codeLine string) (bool, int) {
				lineAbort, lineExitCode := t.lineExecuter(codeLine, script)
				return lineExitCode, lineAbort
			})
			if abort {
				t.getLogger().Debug("abort reason found, or execution failed")
			}

			// waitin until the any target that runns also is done
			if len(runTargetfutures) > 0 {
				ctxout.CtxOut(messageCmdCtrl, ctxout.LabelFY("wait targets"), "waiting until beside running targets are done")
				trgtRes := awaitgroup.WaitAtGroup(runTargetfutures)
				ctxout.CtxOut(messageCmdCtrl, ctxout.LabelFY("wait targets"), "waiting done", trgtRes)
			}
			// next are tarets they runs afterwards the regular
			// script os done
			t.getLogger().WithFields(logrus.Fields{
				"current-target": target,
				"nexts":          script.Next,
			}).Debug("executeTemplate next definition")

			nextfutures := t.generateFuturesByTargetListAndExec(script.Next, t.runCfg)
			nextRes := awaitgroup.WaitAtGroup(nextfutures)
			ctxout.CtxOut(messageCmdCtrl, ctxout.LabelFY("wait next"), "waiting done", nextRes)

			//return returnCode
			// back to old dir if workpace usage was set
			if backToDir != "" {
				os.Chdir(backToDir)
			}

		}
		// we have at least none of the possible task executed.
		if !targetFound {
			ctxout.CtxOut(manout.MessageCln(manout.ForeYellow, "target not defined or matching any requirement: ", manout.ForeWhite, target))
			t.getLogger().Error("Target can not be found: ", target)
			return systools.ExitByNoTargetExists
		}

		if !targetExecuted {
			return systools.ExitByNothingToDo
		}
		return returnCode
	}
	return systools.ExitNoCode
}

func (t *targetExecuter) generateFuturesByTargetListAndExec(RunTargets []string, runCfg configure.RunConfig) []awaitgroup.Future {
	if len(RunTargets) < 1 {
		return []awaitgroup.Future{}
	}
	var runTargetExecs []awaitgroup.FutureStack
	for _, needTarget := range RunTargets {
		t.getLogger().Debug("runTarget name should be added " + needTarget)
		runTargetExecs = append(runTargetExecs, awaitgroup.FutureStack{
			AwaitFunc: func(ctx context.Context) interface{} {
				argTask := ctx.Value(awaitgroup.CtxKey{}).(string)
				_, argmap := systools.StringSplitArgs(argTask, "arg")
				t.getLogger().Debug("add runTarget task " + argTask)
				return t.executeTemplate(true, argTask, argmap)
			},
			Argument: needTarget})

	}
	ctxout.CtxOut(ctxout.LabelFY("async targets"), ctxout.InfoF("count"), len(runTargetExecs), ctxout.InfoF(" targets"))
	return awaitgroup.ExecFutureGroup(runTargetExecs)
}
