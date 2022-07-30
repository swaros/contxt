// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Licensed under the MIT License
//
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
package cmdhandle

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var watchTaskList sync.Map

// TaskDef holds information about running
// and finished tasks
type TaskDef struct {
	started   bool
	count     int
	done      bool
	doneCount int
}

func incTaskCount(target string) {
	taskInfo, _ := getTask(target)
	taskInfo.count++
	watchTaskList.Store(target, taskInfo)
}

func incTaskDoneCount(target string) bool {
	taskInfo, exists := getTask(target)
	if !exists {
		GetLogger().Fatal("can not handle task they never started")
		return false
	}
	taskInfo.doneCount++
	taskInfo.done = taskInfo.doneCount == taskInfo.count
	watchTaskList.Store(target, taskInfo)
	return taskInfo.done

}

// ResetAllTaskInfos resets all task infos
func ResetAllTaskInfos() {
	watchTaskList.Range(func(key, _ interface{}) bool {
		watchTaskList.Delete(key)
		return true
	})
}

// TaskExists checks if a task is already created
func TaskExists(target string) bool {
	_, found := watchTaskList.Load(target)
	return found
}

// TaskRunning checks if a task is already running
func TaskRunning(target string) bool {
	info, found := watchTaskList.Load(target)
	return found && info != nil && info.(TaskDef).count > 0 && info.(TaskDef).count != info.(TaskDef).doneCount
}

// checks if a task was at least started X times
func TaskRunsAtLeast(target string, atLeast int) bool {
	if info, found := watchTaskList.Load(target); found {
		return info.(TaskDef).count >= atLeast
	}
	return false
}

// WaitForTasksDone waits until all the task are done
// triggers a callback for any tick
// and one if the state DONE is reached
// there is an timeout as maximum time to wait
// if this time is reached the process will be continued
// and the timeout callback is triggered
// the callback for notStarted must return true if they handle this issue.
// on returning false it will be counted as isDone
func WaitForTasksDone(tasks []string, timeOut, tickTime time.Duration, stillWait func() bool, isDone func(), timeOutHandle func(), notStartet func(string, string, map[string]string) bool) {
	running := true
	allDone := false
	for running {
		doneCount := 0
		for _, targetFullName := range tasks {
			targetName, args := StringSplitArgs(targetFullName, "arg")
			taskInfo, found := watchTaskList.Load(targetName)
			GetLogger().WithField("task", targetFullName).Trace("checking taskWait for needs")
			if !found && !notStartet(targetFullName, targetName, args) {
				GetLogger().WithField("task", targetFullName).Error("could not check against task that was never started")
				doneCount++
			}
			if found && taskInfo.(TaskDef).done {
				doneCount++
			}
		}

		if doneCount > len(tasks) {
			GetLogger().WithField("tasks", tasks).Fatal("Unexpected count of task reported. taskWatcher seems broken")
		}

		if doneCount == len(tasks) {
			GetLogger().WithFields(logrus.Fields{"tasks": tasks, "doneCount": doneCount, "taskCount": len(tasks)}).Info("Task-wait-check done regular")
			allDone = true
			isDone()
			return
		}
		if stillWait() {
			GetLogger().WithField("sleep", tickTime).Info("Task-wait-check waiting")
			time.Sleep(tickTime)
		} else {
			GetLogger().WithField("tasks", tasks).Info("Task-wait-check done by abort. stillWait() will no longer wait")
			allDone = true
			isDone()
			return
		}
		time.AfterFunc(timeOut, func() {
			if !allDone {
				GetLogger().WithField("tasks", tasks).Warning("Task-wait-check running in Timeout-Check.")
				GetLogger().WithFields(logrus.Fields{
					"timeOut":   timeOut,
					"doneFlag":  allDone,
					"tasks":     tasks,
					"doneCount": doneCount,
					"taskCount": len(tasks)}).Info("timeout variables")
				timeOutHandle()
			}
			running = false
		})
	}
}

func getTask(target string) (TaskDef, bool) {
	taskInfo, found := watchTaskList.Load(target)
	if found && taskInfo != nil {
		return taskInfo.(TaskDef), true
	}
	nwTask := TaskDef{
		count:     0,
		done:      false,
		doneCount: 0,
		started:   false,
	}
	watchTaskList.Store(target, nwTask)
	return nwTask, false

}
