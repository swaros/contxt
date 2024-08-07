// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

package tasks

import (
	"strings"
	"sync"
	"time"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/process"
)

var (
	runners sync.Map
)

func (t *targetExecuter) getIdForTask(currentTask configure.Task) string {
	idStr := currentTask.ID
	if currentTask.Options.Maincmd != "" {
		idStr += "_" + currentTask.Options.Maincmd
	}
	if len(currentTask.Options.Mainparams) > 0 {
		idStr += "_" + strings.Join(currentTask.Options.Mainparams, "_")
	}
	// if the working dir is set, add it to the id because elsewhere the shell would
	// have the wong working dir
	if currentTask.Options.WorkingDir != "" {
		idStr += "_" + currentTask.Options.WorkingDir
	}
	t.getLogger().Trace("tasks.Runner: id for task:", mimiclog.Fields{"id": idStr, "task": currentTask.ID})
	return idStr
}

func (t *targetExecuter) GetRunnerForTask(currentTask configure.Task, callback process.ProcCallback) (*RunnerCtrl, error) {
	idStr := t.getIdForTask(currentTask)
	var runCtl RunnerCtrl
	runCtl.currentTask = currentTask
	runCtl.parentTask = t

	val, ok := runners.Load(idStr)
	if ok {
		t.getLogger().Debug("tasks.Runner: found runner for task:", mimiclog.Fields{"id": idStr, "task": currentTask.ID})
		runner := val.(*process.Process)
		runCtl.runner = runner
		return &runCtl, nil
	}
	runner, err := t.createRunnerForTask(currentTask, callback)
	if err != nil {
		return &runCtl, err
	}
	runCtl.runner = runner
	return &runCtl, nil
}

func (t *targetExecuter) createRunnerForTask(currentTask configure.Task, callback process.ProcCallback) (runner *process.Process, err error) {
	idStr := t.getIdForTask(currentTask)

	if currentTask.Options.Maincmd == "" {
		runner = process.NewTerminal()
	} else {
		runner = process.NewProcess(currentTask.Options.Maincmd, currentTask.Options.Mainparams...)
	}
	runner.SetKeepRunning(true)
	runner.SetOnOutput(callback)
	if _, _, err := runner.Exec(); err != nil {
		return nil, err
	}

	runners.Store(idStr, runner)
	t.getLogger().Debug("tasks.Runner: created runner for task:", mimiclog.Fields{"id": idStr, "task": currentTask.ID})
	return
}

func (t *targetExecuter) WaitTilAllRunnersAreDone(tick time.Duration) {
	for {
		if !t.RunnersActive() {
			time.Sleep(tick)
			break
		}
	}
}

func (t *targetExecuter) StopAllTaskRunner() {
	runners.Range(func(key, value interface{}) bool {
		process := value.(*process.Process)
		t.getLogger().Debug("tasks.Runner: stopping runner:", mimiclog.Fields{"id": key})
		process.Stop()
		return true
	})
	runners = sync.Map{}
}

func (t *targetExecuter) WaitTilTaskRunnerIsDone(currentTask configure.Task, tick time.Duration) {
	for {
		time.Sleep(tick)
		if !t.TaskRunnerIsActive(currentTask) {
			break
		}
	}
}

func (t *targetExecuter) WaitTilTaskRunnerIsRunning(currentTask configure.Task, tick time.Duration, maxTicks int) bool {
	countTicks := 0
	ok := false
	for {
		if t.TaskRunnerIsActive(currentTask) {
			ok = true
			t.getLogger().Debug("tasks.Runner: task is active. get out because of finding:", mimiclog.Fields{"task": currentTask.ID})
			break
		}
		countTicks++
		if countTicks > maxTicks {
			t.getLogger().Debug("tasks.Runner: task is not active (yet). get out because of timeout:", mimiclog.Fields{"task": currentTask.ID})
			ok = false
			break
		}
		time.Sleep(tick)
	}
	return ok
}

func (t *targetExecuter) StopAndRemoveTaskRunner(currentTask configure.Task) {
	idStr := t.getIdForTask(currentTask)
	val, ok := runners.Load(idStr)
	if ok {
		runner := val.(*process.Process)
		runner.Stop()
		runners.Delete(idStr)
		t.getLogger().Debug("tasks.Runner: stopped and removed runner for task:", mimiclog.Fields{"id": idStr, "task": currentTask.ID})
	}
}

func (t *targetExecuter) TaskRunnerIsActive(currentTask configure.Task) bool {
	idStr := t.getIdForTask(currentTask)
	val, ok := runners.Load(idStr)
	if ok {
		runner := val.(*process.Process)
		processWatcher, err := runner.GetProcessWatcher()
		if err != nil {
			return false
		}
		if processWatcher != nil {
			if processWatcher.CountChildsAll() > 0 {
				return true
			}
		}
	}
	return false
}

func (t *targetExecuter) RunnersActive() bool {
	foundAtLeastOnRunning := false
	runners.Range(func(key, value interface{}) bool {
		process := value.(*process.Process)
		procHandl, err := process.GetProcessWatcher()
		if err != nil {
			return true
		}
		if procHandl != nil {
			if procHandl.CountChildsAll() > 0 {
				foundAtLeastOnRunning = true
				return false
			}
		}
		return true
	})
	return foundAtLeastOnRunning
}

type RunnerCtrl struct {
	runner      *process.Process
	currentTask configure.Task
	parentTask  *targetExecuter
	cmdLock     sync.Mutex
}

func (r *RunnerCtrl) GetTask() configure.Task {
	return r.currentTask
}

func (r *RunnerCtrl) GetRunner() *process.Process {
	return r.runner
}

func (r *RunnerCtrl) Cmd(cmd string) error {
	// no race condition please
	r.cmdLock.Lock()
	defer r.cmdLock.Unlock()

	if err := r.runner.Command(cmd); err != nil {
		return err
	}
	// give the task a chance to start and excute the command
	r.parentTask.WaitTilTaskRunnerIsRunning(r.currentTask, 1*time.Millisecond, 50)
	// wait til the task is done. this is done by watching the childs of the process
	r.parentTask.WaitTilTaskRunnerIsDone(r.currentTask, 10*time.Millisecond)
	return nil
}
