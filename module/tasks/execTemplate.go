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
package tasks

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/swaros/contxt/module/awaitgroup"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/systools"
)

type TaskListExec struct {
	config                 configure.RunConfig
	watch                  *Watchman
	subTasks               map[string]*targetExecuter
	args                   []interface{}
	logger                 mimiclog.Logger
	presetHardExistOnError bool
}

func NewTaskListExec(config configure.RunConfig, adds ...interface{}) *TaskListExec {
	return &TaskListExec{
		config: config,
		watch:  NewGlobalWatchman(),
		args:   adds,
	}
}

func NewStdTaskListExec(config configure.RunConfig, adds ...interface{}) *TaskListExec {
	dmc := NewCombinedDataHandler()
	req := NewDefaultRequires(dmc, mimiclog.NewNullLogger())
	if adds == nil {
		adds = make([]interface{}, 0)
	}
	adds = append(adds, dmc, req)

	return &TaskListExec{
		config: config,
		watch:  NewGlobalWatchman(),
		args:   adds,
	}
}

func (e *TaskListExec) RunTarget(target string, async bool) int {
	scopeVars := make(map[string]string)
	return e.RunTargetWithVars(target, scopeVars, async)
}

func (e *TaskListExec) RunTargetWithVars(target string, scopeVars map[string]string, async bool) int {
	tExec := e.findOrCreateTask(target, scopeVars)
	if tExec == nil {
		return systools.ExitByNoTargetExists
	}
	return tExec.executeTemplate(async, target, scopeVars)
}

func (e *TaskListExec) SetLogger(logger mimiclog.Logger) {
	e.logger = logger
}

func (e *TaskListExec) GetTask(target string) *targetExecuter {
	if e.subTasks == nil {
		e.subTasks = make(map[string]*targetExecuter)
		return nil
	}
	if tExec, found := e.subTasks[target]; found {
		return e.applyLogger(tExec)
	}
	return nil
}

func (e *TaskListExec) applyLogger(tExec *targetExecuter) *targetExecuter {
	if e.logger != nil && tExec != nil && !tExec.haveLogger() {
		tExec.SetLogger(e.logger)
	}
	return tExec
}

func (e *TaskListExec) GetWatch() *Watchman {
	return e.watch
}

func (e *TaskListExec) findOrCreateTask(target string, scopeVars map[string]string) *targetExecuter {
	if e.subTasks == nil {
		e.subTasks = make(map[string]*targetExecuter)
	}
	tExec, found := e.subTasks[target]
	if !found {
		for _, task := range e.config.Task {
			if task.ID == target {
				e.args = append(e.args, e.config) // add the config to the args
				tExec = New(target, scopeVars, e.args...)
				// take the preset also for any new task
				tExec.SetHardExitOnError(e.presetHardExistOnError)
				e.subTasks[target] = tExec
			}
		}
	}
	return e.applyLogger(tExec)
}

func (e *TaskListExec) SetHardExistToAllTasks(exitOnErr bool) {
	e.presetHardExistOnError = exitOnErr
	for _, task := range e.subTasks {
		task.SetHardExitOnError(exitOnErr)
	}
}

func (t *targetExecuter) executeTemplate(runAsync bool, target string, scopeVars map[string]string) int {
	if t == nil {
		panic("targetExecuter is nil. This should not happen. init it with New()")
	}
	if t.watch == nil {
		panic("watch is nil. This should not happen. init it with NewWatchman()")
	}
	// check if task is already running
	// this check depends on the target name.
	if !t.runCfg.Config.AllowMutliRun && t.watch.TaskRunning(target) {
		logFields := mimiclog.Fields{
			"target": target,
		}
		t.getLogger().Error("task would be triggered again while is already running. IGNORED", logFields)
		return systools.ExitAlreadyRunning
	}

	// increment task counter
	cnt := t.watch.IncTaskCount(target)
	t.out(MsgTarget{Target: target, Context: "task-start-and-count-is-now", Info: fmt.Sprintf("%v", cnt)})
	defer t.watch.IncTaskDoneCount(target) // save done count at then end

	t.getLogger().Info("executeTemplate LOOKING for target", target)

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
		// here we filter the task they is matching the ids
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
		taskCount := len(taskList)
		for curTIndex, script := range taskList {
			logFields := mimiclog.Fields{
				"target": target,
				"scope":  scopeVars,
			}
			t.getLogger().Info("executeTemplate EXECUTE target", logFields)
			targetFound = true

			//stopReason := script.Stopreasons
			/*
				var messageCmdCtrl TaskOutCtrl = TaskOutCtrl{ // define a controll hook, depending on the display comand option
					IgnoreCase: !script.Options.Displaycmd, // we ignore the message, as long the display command is NOT set
				}*/

			// check requirements
			canRun, message := t.checkRequirements(script.Requires)
			if !canRun {
				logFields := mimiclog.Fields{
					"target": target,
					"reason": message,
				}
				t.getLogger().Info("executeTemplate IGNORE because requirements not matching", logFields)
				if script.Options.Displaycmd {
					t.out(MsgTarget{Target: target, Context: "requirement-check-failed", Info: message}, MsgNumber(curTIndex+1))
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
					t.out(MsgTarget{Target: target, Context: "workspace-setting-invalid", Info: chDirError.Error()})
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
				if script.Options.Displaycmd {
					t.out(MsgTarget{Target: target, Context: "needs_required", Info: strings.Join(script.Needs, ",")}, MsgArgs(script.Needs))
				}
				t.getLogger().Debug("Needs for the script", script.Needs)
				// check if we have to run the needs in threads or not
				if runAsync {
					// here we have the "run in threads" part
					var needExecs []awaitgroup.FutureStack
					for _, needTarget := range script.Needs {

						t.watch.VerifyTaskExists(needTarget, func(taskExists bool) {
							if taskExists {
								if script.Options.Displaycmd {
									t.out(MsgTarget{Target: target, Context: "needs_ignored_runs_already", Info: needTarget})
								}
								t.getLogger().Debug("need already handled " + needTarget)
							} else {
								// task is not registered yet, so we will do it
								t.watch.CreateTask(needTarget)
								t.getLogger().Debug("need name should be added " + needTarget)
								if script.Options.Displaycmd {
									t.out(MsgTarget{Target: target, Context: "needs_execute", Info: needTarget})
								}
								needExecs = append(needExecs, awaitgroup.FutureStack{
									AwaitFunc: func(ctx context.Context) interface{} {
										argNeed := ctx.Value(awaitgroup.CtxKey{}).(string)
										_, argmap := systools.StringSplitArgs(argNeed, "arg")
										t.getLogger().Debug("add need task " + argNeed)
										return t.executeTemplate(true, argNeed, argmap)
									},
									Argument: needTarget})
							}
						})

					}
					futures := awaitgroup.ExecFutureGroup(needExecs) // create the futures and start the tasks
					results := awaitgroup.WaitAtGroup(futures)       // wait until any task is executed

					t.getLogger().Debug("needs result", results)
				} else {
					for _, syncTarget := range script.Needs {
						t.watch.VerifyTaskExists(syncTarget, func(exists bool) {
							if exists {
								t.getLogger().Debug("need already handled " + syncTarget)
								if script.Options.Displaycmd {
									t.out(MsgTarget{Target: target, Context: "needs_ignored_runs_already", Info: syncTarget})
								}
							} else {
								_, argmap := systools.StringSplitArgs(syncTarget, "arg")
								t.executeTemplate(false, syncTarget, argmap)
							}

						})
					}

				}
				if script.Options.Displaycmd {
					t.out(MsgTarget{Target: target, Context: "needs_done", Info: strings.Join(script.Needs, ",")}, MsgArgs(script.Needs))
				}
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
				lineAbort, lineExitCode := t.targetTaskExecuter(codeLine, script, t.watch)
				return lineExitCode, lineAbort
			})
			if abort {
				t.getLogger().Debug("abort reason found, or execution failed")
				// if we have a return code, we need to return it
				if returnCode == systools.ErrorCheatMacros {
					return returnCode
				}
			}

			// waitin until the any target that runns also is done
			if len(runTargetfutures) > 0 {
				t.out(MsgTarget{Target: target, Context: "wait_for_targets", Info: strings.Join(script.RunTargets, ",")}, MsgArgs(script.RunTargets))
				awaitgroup.WaitAtGroup(runTargetfutures)
				t.out(MsgTarget{Target: target, Context: "wait_targets_done"})
			}
			// next are tarets they runs afterwards the regular
			// script os done
			logFields2 := mimiclog.Fields{
				"current-target": target,
				"nexts":          script.Next,
			}
			t.getLogger().Debug("executeTemplate next definition", logFields2)

			nextfutures := t.generateFuturesByTargetListAndExec(script.Next, t.runCfg)
			awaitgroup.WaitAtGroup(nextfutures)

			t.out(MsgTarget{Target: target, Context: "wait_next_done", Info: fmt.Sprintf("(%v/%v)", curTIndex+1, taskCount)})

			//return returnCode
			// back to old dir if workpace usage was set
			if backToDir != "" {
				os.Chdir(backToDir)
			}

		}
		// we have at least none of the possible task executed.
		if !targetFound {
			//t.out(MsgTarget(target), MsgType("not_found"))
			t.out(MsgTarget{Target: target, Context: "not_found"})
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
	t.out(MsgType("target-async-group-created"), MsgArgs(RunTargets))
	return awaitgroup.ExecFutureGroup(runTargetExecs)
}
