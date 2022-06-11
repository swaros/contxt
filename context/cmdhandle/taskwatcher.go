package cmdhandle

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var taskList sync.Map

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
	taskList.Store(target, taskInfo)
}

func incTaskDoneCount(target string) bool {
	taskInfo, exists := getTask(target)
	if !exists {
		GetLogger().Fatal("can not handle task they never started")
		return false
	}
	taskInfo.doneCount++
	taskInfo.done = taskInfo.doneCount == taskInfo.count
	taskList.Store(target, taskInfo)
	return taskInfo.done

}

// ResetAllTaskInfos resets all task infos
func ResetAllTaskInfos() {
	taskList.Range(func(key, value interface{}) bool {
		taskList.Delete(key)
		return true
	})
}

// TaskExists checks if a tak is already created
func TaskExists(target string) bool {
	_, found := taskList.Load(target)
	return found
}

// TaskRunning checks if a task is already running
func TaskRunning(target string) bool {
	info, found := taskList.Load(target)
	return found && info != nil && info.(TaskDef).count > 0 && info.(TaskDef).count != info.(TaskDef).doneCount
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
			taskInfo, found := taskList.Load(targetName)
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
	taskInfo, found := taskList.Load(target)
	if found && taskInfo != nil {
		return taskInfo.(TaskDef), true
	}
	nwTask := TaskDef{
		count:     0,
		done:      false,
		doneCount: 0,
		started:   false,
	}
	taskList.Store(target, nwTask)
	return nwTask, false

}
