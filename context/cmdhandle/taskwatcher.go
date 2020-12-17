package cmdhandle

import (
	"sync"
	"time"
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

// WaitForTasksDone waits until all the task are done
// triggers a callback for any tick
// and one if the state DONE is reached
// there is an timeout as maximum time to wait
// if this time is reached the process will be continued
// and the timeout callback is triggered
// the callback for notStarted must return true if they handle this issue.
// on returning false it will be counted as isDone
func WaitForTasksDone(tasks []string, timeOut, tickTime time.Duration, stillWait func() bool, isDone func(), timeOutHandle func(), notStartet func(string) bool) {
	running := true
	for running {
		doneCount := 0
		for _, targetName := range tasks {
			taskInfo, found := taskList.Load(targetName)
			if found == false && notStartet(targetName) == false {
				GetLogger().WithField("task", targetName).Error("could not check against task that was never started")
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
			GetLogger().WithField("tasks", tasks).Info("Task-wait-check done regular")
			isDone()
			return
		}
		if stillWait() {
			time.Sleep(tickTime)
		} else {
			GetLogger().WithField("tasks", tasks).Info("Task-wait-check done by abort. stillWait() will no longer wait")
			isDone()
			return
		}
		time.AfterFunc(timeOut, func() {
			GetLogger().WithField("tasks", tasks).Warning("Task-wait-check done by Timeout.")
			timeOutHandle()
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
