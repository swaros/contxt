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
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/swaros/contxt/module/mimiclog"
)

const GlobalName = "global"

var (
	instances     = make(map[string]*Watchman)
	instanceMutex sync.Mutex
)

// Track watchman instances
// in a global sync map

func storeNewInstance(wm *Watchman) {
	uuidString := uuid.New().String()
	storeNamedInstance(wm, uuidString)
}

func storeNamedInstance(wm *Watchman, name string) {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()
	instances[name] = wm
}

func GetWatcherInstance(uuidString string) *Watchman {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()
	if wm, found := instances[uuidString]; found {
		return wm
	}
	return nil
}

func ListWatcherInstances() []string {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()
	var inst []string
	for k := range instances {
		inst = append(inst, k)
	}
	return inst
}

func ShutDownProcesses(reportFn func(target string, time int, succeed bool)) {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()
	for _, wm := range instances {
		wm.StopAllTasks(reportFn)
	}
}

// the watchman implementation
type Watchman struct {
	// contains filtered or unexported fields
	watchTaskList sync.Map
	mu            sync.Mutex
	logger        mimiclog.Logger
}

func NewGlobalWatchman() *Watchman {
	wm := GetWatcherInstance(GlobalName)
	if wm == nil {
		wm := &Watchman{
			watchTaskList: sync.Map{},
			logger:        mimiclog.NewNullLogger(),
		}
		storeNamedInstance(wm, GlobalName)
		return wm
	}
	return wm
}

func NewWatchman() *Watchman {
	wm := &Watchman{
		logger:        mimiclog.NewNullLogger(),
		watchTaskList: sync.Map{},
	}
	storeNewInstance(wm)
	return wm
}

func (w *Watchman) GetTask(target string) (TaskDef, bool) {
	taskInfo, found := w.watchTaskList.Load(target)
	if found && taskInfo != nil {
		return taskInfo.(TaskDef), true
	}
	return TaskDef{}, false
}

func (w *Watchman) SetLogger(logger mimiclog.Logger) {
	w.logger = logger
}

func (w *Watchman) StopAllTasks(reportFn func(target string, time int, succeed bool)) {
	w.watchTaskList.Range(func(key, _ interface{}) bool {
		target := key.(string)
		done, timeNeeded := w.WaitForStopProcess(target, 100*time.Millisecond, 10)
		if reportFn != nil {
			reportFn(target, timeNeeded, done)
		}
		return true
	})
}

// WaitForProcessStart waits until the process is started
// or the timeout is reached
// the timeout is defined by the tickDuration multiplied by the maxTicks
func (w *Watchman) WaitForProcessStart(target string, tickDuration time.Duration, maxTicks int) (bool, int) {
	// repeat until the process is started
	// or the timeout is reached
	// or the process is not running anymore
	currentTick := 0
	for {
		if wtask, found := w.GetTask(target); found {
			if wtask.IsProcessRunning() {
				return true, currentTick * int(tickDuration)
			}
		}
		time.Sleep(tickDuration)
		currentTick++
		if currentTick >= maxTicks {
			return false, currentTick * int(tickDuration)
		}
	}
}

func (w *Watchman) WaitForStopProcess(target string, tickDuration time.Duration, maxTicks int) (bool, int) {
	// repeat until the process is started
	// or the timeout is reached
	// or the process is not running anymore
	currentTick := 0
	alreadyTryToStop := false
	for {
		if wtask, found := w.GetTask(target); found {
			if !wtask.IsProcessRunning() {
				w.logger.Debug("process is not running anymore", target)
				return true, currentTick * int(tickDuration)
			}
			// send the stop signal once.
			// we do not want to spam the process with signals
			// instead we will ry to kill it if it is still running
			if !alreadyTryToStop && wtask.IsProcessRunning() {
				w.logger.Debug("try to stop process with os.Interrupt ", target)
				if err := wtask.StopProcessIfRunning(); err != nil {
					w.logger.Warn("failed to stop process with os.Interrupt ", target, err)
					alreadyTryToStop = true
				}
			}
			time.Sleep(tickDuration)
			currentTick++
			if currentTick == maxTicks {
				// in the last tick we try to kill the process
				if wtask.IsProcessRunning() {
					w.logger.Debug("try to stop process with os.Kill ", target)
					if err := wtask.KillProcess(); err != nil {
						w.logger.Warn("failed to stop process with os.Kill ", target, err)
						return false, currentTick * int(tickDuration)
					} else {
						w.logger.Debug("process stopped with os.Kill ", target)
						return true, currentTick * int(tickDuration)
					}
				}
			}
			if currentTick > maxTicks {
				// overtime, so we we just report the process as stopped or not
				isRunning := wtask.IsProcessRunning()
				w.logger.Debug("process is running(?)", target, isRunning)
				return isRunning, currentTick * int(tickDuration)
			}
		} else {
			// no task found, so we can stop
			w.logger.Debug("no task found ", target)
			return true, currentTick * int(tickDuration)
		}
	}
}

// do not sync this function.
func (w *Watchman) CreateTask(target string) {
	w.watchTaskList.Store(target, TaskDef{
		uuid:      uuid.New().String(),
		count:     0,
		done:      false,
		doneCount: 0,
		started:   false,
	})
}

func (w *Watchman) getTaskOrCreate(target string) (TaskDef, bool) {
	taskInfo, found := w.watchTaskList.Load(target)
	if found && taskInfo != nil {
		return taskInfo.(TaskDef), true
	}
	nwTask := TaskDef{
		uuid:      uuid.New().String(),
		count:     0,
		done:      false,
		doneCount: 0,
		started:   false,
	}
	w.watchTaskList.Store(target, nwTask)
	return nwTask, false

}

func (w *Watchman) UpdateTask(target string, task TaskDef) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	// we need to make sure that the task is already done.
	// the update is not allowed to create a new task
	// or update an task with a different uuid
	taskInfo, found := w.watchTaskList.Load(target)
	if found && taskInfo != nil {
		if taskInfo.(TaskDef).uuid == task.uuid {
			w.watchTaskList.Store(target, task)
			return nil
		} else {
			return fmt.Errorf("can not update task %q, because the uuid is different", target)
		}
	}
	return fmt.Errorf("can not update task %q, because it does not exists", target)

}

func (w *Watchman) IncTaskCount(target string) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	taskInfo, _ := w.getTaskOrCreate(target)
	taskInfo.started = true
	taskInfo.count++
	w.watchTaskList.Store(target, taskInfo)
	return taskInfo.count
}

func (w *Watchman) IncTaskDoneCount(target string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	taskInfo, exists := w.getTaskOrCreate(target)
	if !exists {
		return false
	}
	taskInfo.doneCount++
	taskInfo.done = taskInfo.doneCount == taskInfo.count
	w.watchTaskList.Store(target, taskInfo)
	return taskInfo.done

}

// ResetAllTaskInfos resets all task infos
func (w *Watchman) ResetAllTaskInfos() {
	w.watchTaskList.Range(func(key, _ interface{}) bool {
		w.WaitForStopProcess(key.(string), 100*time.Millisecond, 10)
		w.watchTaskList.Delete(key)
		return true
	})

}

// TaskRunning checks if a task is already running
func (w *Watchman) TaskRunning(target string) bool {
	info, found := w.watchTaskList.Load(target)
	return found && info != nil && info.(TaskDef).count > 0 && info.(TaskDef).count != info.(TaskDef).doneCount
}

// returns the list of all running tasks as string slice by the task name
func (w *Watchman) GetAllRunningTasks() []string {
	var tasks []string
	w.watchTaskList.Range(func(key, _ interface{}) bool {
		if w.TaskRunning(key.(string)) {
			tasks = append(tasks, key.(string))
		}
		return true
	})
	return tasks
}

func (w *Watchman) ResetAllTasksIfPossible() error {
	if len(w.GetAllRunningTasks()) == 0 {
		w.ResetAllTaskInfos()
		return nil
	}
	return fmt.Errorf("can not reset watchman, because there are still %v running tasks", len(w.GetAllRunningTasks()))
}

// checks if a task was at least started X times
func (w *Watchman) TaskRunsAtLeast(target string, atLeast int) bool {
	w.mu.Lock()
	instanceMutex.Lock() // we also need to lock the instanceMutex
	defer instanceMutex.Unlock()
	defer w.mu.Unlock()
	if info, found := w.watchTaskList.Load(target); found {
		return info.(TaskDef).count >= atLeast
	}
	return false
}

func (w *Watchman) TaskIsRegisteredCallBack(target string, handleFn func(bool)) bool {
	w.mu.Lock()
	instanceMutex.Lock() // we also need to lock the instanceMutex
	defer instanceMutex.Unlock()
	defer w.mu.Unlock()
	_, found := w.watchTaskList.Load(target)
	if handleFn != nil {
		handleFn(found)
	}
	return found

}

func (w *Watchman) IncTaskIfCounterIsLessThen(target string, lessThen int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if info, found := w.watchTaskList.Load(target); found {
		task := info.(TaskDef)
		if task.count < lessThen {
			task.count++
			w.watchTaskList.Store(target, task)
			return nil
		}
		return fmt.Errorf("can not update task %q, because the count is greater or equal %d", target, lessThen)
	}
	return fmt.Errorf("can not update task %q, because it does not exists", target)
}

func (w *Watchman) GetTaskCount(target string) int {
	if info, found := w.watchTaskList.Load(target); found {
		return info.(TaskDef).count
	}
	return 0
}

func (w *Watchman) GetTaskDoneCount(target string) int {
	if info, found := w.watchTaskList.Load(target); found {
		return info.(TaskDef).doneCount
	}
	return 0
}

func (w *Watchman) GetTaskDone(target string) bool {
	if info, found := w.watchTaskList.Load(target); found {
		return info.(TaskDef).done
	}
	return false
}

func (w *Watchman) GetTaskStarted(target string) bool {
	if info, found := w.watchTaskList.Load(target); found {
		return info.(TaskDef).started
	}
	return false
}

func (w *Watchman) ListTasks() []string {
	var tasks []string
	w.watchTaskList.Range(func(key, _ interface{}) bool {
		tasks = append(tasks, key.(string))
		return true
	})
	return tasks
}
