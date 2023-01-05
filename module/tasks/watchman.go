package tasks

import "sync"

type Watchman struct {
	// contains filtered or unexported fields
	watchTaskList sync.Map
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

func (w *Watchman) incTaskCount(target string) int {
	taskInfo, _ := w.getTask(target)
	taskInfo.count++
	w.watchTaskList.Store(target, taskInfo)
	return taskInfo.count
}

func (w *Watchman) incTaskDoneCount(target string) bool {
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
