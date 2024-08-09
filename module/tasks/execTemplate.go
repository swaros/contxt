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
	"strings"
	"time"

	"github.com/swaros/contxt/module/awaitgroup"
	"github.com/swaros/contxt/module/configure"
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

// findOrCreateTask returns the taskExecuter for the given target.
// if the task is not found, it will be created.
func (e *TaskListExec) findOrCreateTask(target string, scopeVars map[string]string) *targetExecuter {
	// first create the tasklist if not exists
	if e.subTasks == nil {
		e.subTasks = make(map[string]*targetExecuter)
	}
	// check if the task is already created
	tExec, found := e.subTasks[target]
	if !found { // task not found, so we need to create it
		for _, task := range e.config.Task { // check if the task is defined in the config
			if task.ID == target { // task found
				e.args = append(e.args, e.config) // add the config to the args
				tExec = New(target, scopeVars, e.args...)
				// take the preset also for any new task
				tExec.SetHardExitOnError(e.presetHardExistOnError)
				e.subTasks[target] = tExec // add the task to the tasklist
				if e.logger != nil {       // if we have a logger, we will set it to the task
					tExec.SetLogger(e.logger)
				}
			}
		}
	}
	return e.applyLogger(tExec) // return the task with the logger. we are doing this here again because the logger can be set after the task was created
}

func (e *TaskListExec) SetHardExistToAllTasks(exitOnErr bool) {
	e.presetHardExistOnError = exitOnErr
	for _, task := range e.subTasks {
		task.SetHardExitOnError(exitOnErr)
	}
}

func (t *targetExecuter) verifiedKeyname(keyName string) (string, bool) {
	// just trim spaces
	keyName = strings.TrimSpace(keyName)
	// some weird target name? we will not allow this
	if clTarget, err := systools.CheckForCleanString(keyName); err != nil {
		t.getLogger().Error("invalid key-name", err)
		return "", false
	} else {
		keyName = clTarget
	}
	return keyName, true
}

func (t *targetExecuter) verifyVersion() bool {
	if t.runCfg.Version == "" || configure.GetVersion() == "" {
		return true
	}
	return configure.CheckVersion(t.runCfg.Version, configure.GetVersion())
}

func (t *targetExecuter) executeTemplate(runAsync bool, target string, scopeVars map[string]string) int {

	// check the version of the task
	if !t.verifyVersion() {
		t.getLogger().Error("unsupported version", t.runCfg.Version, " current version is ", configure.GetVersion())
		return systools.ExitByUnsupportedVersion
	}
	// some weird target name? we will not allow this
	if clTarget, err := systools.CheckForCleanString(target); err != nil {
		t.getLogger().Error("invalid target name", err)
		return systools.ErrorTemplateReading
	} else {
		target = clTarget
	}

	// just trim spaces
	target = strings.TrimSpace(target)

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
	t.watch.IncTaskCount(target)
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

			// just the abort flag.
			abort := false

			// -- NEEDS
			// needs are task, the have to be startet once
			// before we continue.
			// any need can have his own needs they also needs to
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
						// check if the task is already registered
						if !t.watch.TryCreate(needTarget) {
							// task is already registered, so we will not do it
							if script.Options.Displaycmd {
								t.out(MsgTarget{Target: target, Context: "needs_ignored_runs_already", Info: needTarget})
							}
							t.getLogger().Debug("need already handled " + needTarget)
						} else {
							// task is not registered, so it never runs. we need to run it
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
					}

					futures := awaitgroup.ExecFutureGroup(needExecs) // create the futures and start the tasks
					results := awaitgroup.WaitAtGroup(futures)       // wait until any task is executed

					t.getLogger().Debug("needs result", results)
				} else {
					// here we have the "run in sequence" part
					// we need to run the needs in sequence
					// so there is no syncronisation needed for this part of needs,
					// but others can be run in parallel the same needs, so we still need
					// to use the watchman to check if the task is already running
					for _, syncTarget := range script.Needs {
						if !t.watch.TryCreate(syncTarget) {
							// task is already registered, so we will not do it
							t.getLogger().Debug("need already handled " + syncTarget)
							if script.Options.Displaycmd {
								t.out(MsgTarget{Target: target, Context: "needs_ignored_runs_already", Info: syncTarget})
							}
						} else {
							_, argmap := systools.StringSplitArgs(syncTarget, "arg")
							t.executeTemplate(false, syncTarget, argmap)
						}
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
			runTargetfutures := t.generateFuturesByTargetListAndExec(script.RunTargets)

			// check if we have script lines.
			// if not, we need at least to check
			// 'now' listener
			if len(script.Script) < 1 && len(script.Cmd) < 1 {
				t.getLogger().Debug("no script lines defined. run listener anyway")
				t.listenerWatch("", nil, &script)
				// workaround til the async runnig is refactored
				// now we need to give the subtask time to run and update the waitgroup
				// UPDATE: set from 1 second to 15 milliseconds
				// this is a workaround for the async running, but right now it is no
				// longer clear if we need this workaround. need to investigate
				duration := time.Millisecond * time.Duration(15)
				time.Sleep(duration)
			}

			// execute the ank commands if exists
			if len(script.Cmd) > 0 {
				if returnCode, err := t.runAnkCmd(&script); err != nil {
					t.getLogger().Error("error while executing ank commands", err)
					return returnCode
				}
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

			nextfutures := t.generateFuturesByTargetListAndExec(script.Next)
			awaitgroup.WaitAtGroup(nextfutures)

			t.out(MsgTarget{Target: target, Context: "wait_next_done", Info: fmt.Sprintf("(%v/%v)", curTIndex+1, taskCount)})
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

func (t *targetExecuter) generateFuturesByTargetListAndExec(RunTargets []string) []awaitgroup.Future {
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
