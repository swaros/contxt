package tasks

import (
	"strings"
	"sync"
	"time"

	"github.com/swaros/contxt/module/configure"
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
	return idStr
}

func (t *targetExecuter) GetRunnerForTask(currentTask configure.Task, callback process.ProcCallback) (runner *process.Process, err error) {
	idStr := t.getIdForTask(currentTask)
	val, ok := runners.Load(idStr)
	if ok {
		runner = val.(*process.Process)
		return
	}
	runner, err = t.createRunnerForTask(currentTask, callback)
	if err != nil {
		return
	}
	return
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
		time.Sleep(tick)
		if t.TaskRunnerIsActive(currentTask) {
			ok = true
			break
		}
		countTicks++
		if countTicks > maxTicks {
			ok = false
			break
		}
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
