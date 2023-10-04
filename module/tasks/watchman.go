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

import "sync"

type Watchman struct {
	// contains filtered or unexported fields
	watchTaskList sync.Map
	mu            sync.Mutex
}

// TaskDef holds information about running
// and finished tasks
type TaskDef struct {
	started   bool
	count     int
	done      bool
	doneCount int
}

func NewWatchman() *Watchman {
	return &Watchman{}
}

func (w *Watchman) getTask(target string) (TaskDef, bool) {
	taskInfo, found := w.watchTaskList.Load(target)
	if found && taskInfo != nil {
		return taskInfo.(TaskDef), true
	}
	nwTask := TaskDef{
		count:     0,
		done:      false,
		doneCount: 0,
		started:   false,
	}
	w.watchTaskList.Store(target, nwTask)
	return nwTask, false

}

func (w *Watchman) IncTaskCount(target string) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	taskInfo, _ := w.getTask(target)
	taskInfo.started = true
	taskInfo.count++
	w.watchTaskList.Store(target, taskInfo)
	return taskInfo.count
}

func (w *Watchman) IncTaskDoneCount(target string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	taskInfo, exists := w.getTask(target)
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
		w.watchTaskList.Delete(key)
		return true
	})
}

// TaskExists checks if a task is already created
func (w *Watchman) TaskExists(target string) bool {
	_, found := w.watchTaskList.Load(target)
	return found
}

// TaskRunning checks if a task is already running
func (w *Watchman) TaskRunning(target string) bool {
	info, found := w.watchTaskList.Load(target)
	return found && info != nil && info.(TaskDef).count > 0 && info.(TaskDef).count != info.(TaskDef).doneCount
}

// checks if a task was at least started X times
func (w *Watchman) TaskRunsAtLeast(target string, atLeast int) bool {
	if info, found := w.watchTaskList.Load(target); found {
		return info.(TaskDef).count >= atLeast
	}
	return false
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
