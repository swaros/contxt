package cmdhandle

import (
	"errors"
	"strings"
	"sync"
	"time"
)

var procTracker sync.Map

type TaskRuntimeState struct {
	// flag that will be true if the task is started already. or start is initiated
	Started bool
	// flag that is true if the task is timed out. not equals to Done
	TimedOut bool
	// flag if task is done
	Done bool
	// the runId. equals to the taskname
	RunId string
}

// simple container used to get the result
// from task
type TaskResult struct {
	Error   error
	Content interface{}
}

type TaskWatched struct {
	// current runtime states
	task TaskRuntimeState
	// taskname that should be unique
	taskName string
	// arguments for the task
	TaskArgs map[string]string
	// if true, no error will be raised, if the same task will tried to be executed twice
	NoErrorIfBlocked bool
	// maximum of allowed runtime for the task
	TimeOutTiming time.Duration
	// optional callback. if set, these function is called. that is ment to be informed and for doing some cleanups.
	// but it can not interrupt the timeeout.
	TimeOutHandler func()
	// optional callback. this method can decide, if this task could be started again.
	// so this method will be executed only on the second try to run this task.
	CanRunAgain func(*TaskWatched) bool
	// can be set to define a custom taskid by some logic
	GetRunId func() string
	// the main function that contains the logic
	Exec func(*TaskWatched) TaskResult
	// simple callback that can be used for regular output or loggings
	LoggerFnc func(...interface{})
	// if true, the task will be started async.
	Async bool
	// callback for the execution result
	ResultFnc func(TaskResult)
}

type TaskGroup struct {
	tasks []TaskWatched
	// simple callback that can be used for regular output or loggings
	LoggerFnc func(...interface{})
}

// helper to creates a task result
func CreateTaskResult(err error) TaskResult {
	return TaskResult{
		Error: err,
	}
}

// helper to creates a task result
func CreateTaskResultContent(err error, content interface{}) TaskResult {
	return TaskResult{
		Error:   err,
		Content: content,
	}
}

func (t *TaskWatched) Init(name string) {
	t.taskName = name
	// if nothing set, we set to 30 minutes.
	// because this is ment to be used in needs (task they are runs for preperations)
	// it can be expected, some of these pre-tasks are time consuming.
	if t.TimeOutTiming == 0 {
		t.TimeOutTiming = 30 * time.Minute
	}

	t.TaskArgs = make(map[string]string)
	if t.GetRunId == nil {
		t.task.RunId = name
	} else {
		t.task.RunId = t.GetRunId()
	}
}

func (t *TaskWatched) Log(msg ...interface{}) {
	if t.LoggerFnc != nil {
		t.LoggerFnc(msg...)
	}
}

func (t *TaskWatched) trackStart() bool {
	if _, exists := procTracker.Load(t.task.RunId); exists {
		// if a decion method defined we ask them.
		// if not we disagree
		if t.CanRunAgain != nil {
			return t.CanRunAgain(t)
		}
		t.Log("allready runs ", t.task.RunId)
		return false
	}
	t.task.Started = true
	t.task.Done = false
	t.Log("save runtime tracking for task ", t.task.RunId)
	procTracker.Store(t.task.RunId, t.task)
	return true

}

func (t *TaskWatched) IsRunning() bool {
	if t.task.Done {
		return false
	}
	if task, exists := procTracker.Load(t.task.RunId); exists {
		var taskSet TaskRuntimeState = task.(TaskRuntimeState)
		if taskSet.Done {
			return false
		}

	}
	return true

}

func (t *TaskWatched) ReportDone() {
	if task, exists := procTracker.Load(t.task.RunId); exists {
		var taskSet TaskRuntimeState = task.(TaskRuntimeState)
		if taskSet.Done {
			t.Log("Task ", t.taskName, " was already set to DONE")
			return
		}
		taskSet.Done = true
		t.Log("Update Task. Set ", t.taskName, " DONE")
		procTracker.Store(taskSet.RunId, taskSet)
	} else {
		t.task.Done = true
		t.Log("Update Task ", t.taskName, " NOT EXISTS ")
	}
}

func (t *TaskWatched) Run() TaskResult {
	t.Log(" --> run func \t", t.taskName, " id ", t.task.RunId)
	var taskRes TaskResult
	if t.Exec == nil {
		t.ReportDone()
		taskRes.Error = errors.New("body function exec is undefined")
		return taskRes
	}
	// starting the body function and track the execution
	// first we check if can start the task
	if !t.trackStart() {
		if t.NoErrorIfBlocked {
			return taskRes
		}
		taskRes.Error = errors.New("task is already running")
		return taskRes
	}

	time.AfterFunc(t.TimeOutTiming, func() {

		t.Log("timeout reached on task ", t.taskName, " was set to ", t.TimeOutTiming)
		// there is no decsion alowed. timed out task
		// would not be executed.
		// a defined timeOut callback is just
		// ment for cleanup
		if t.TimeOutHandler != nil {
			t.TimeOutHandler()
		}
		// update task info
		if task, exists := procTracker.Load(t.task.RunId); exists {
			var taskDef TaskRuntimeState = task.(TaskRuntimeState)
			taskDef.TimedOut = true
			procTracker.Store(t.task.RunId, taskDef)
		}

	})
	defer t.ReportDone()
	res := t.Exec(t)
	if t.ResultFnc != nil {
		t.ResultFnc(res)
	}
	return res
}

func (t *TaskWatched) GetName() string {
	return t.taskName
}

func CreateMultipleTask(tasks []string, modifyTask func(*TaskWatched)) TaskGroup {
	var taskGrp TaskGroup
	for _, task := range tasks {
		newTask := TaskWatched{
			taskName: task,
		}
		newTask.Init(task)
		modifyTask(&newTask)
		taskGrp.tasks = append(taskGrp.tasks, newTask)
	}
	return taskGrp
}

func (Tg *TaskGroup) GetTask(name string) (bool, TaskWatched) {
	for _, tw := range Tg.tasks {
		if strings.EqualFold(tw.taskName, name) {
			return true, tw
		}
	}
	return false, TaskWatched{}
}

func (Tg *TaskGroup) AddTask(name string, wg TaskWatched) *TaskGroup {
	wg.Init(name)
	Tg.tasks = append(Tg.tasks, wg)
	return Tg
}

// counts the amount of task they are async tasks
func (Tg *TaskGroup) getAsyncCount() int {
	cnt := 0
	for _, tsk := range Tg.tasks {
		if tsk.Async && tsk.Exec != nil {
			cnt++
		}
	}
	return cnt
}

func (Tg *TaskGroup) Exec() *TaskGroup {
	var waitGroup sync.WaitGroup
	taskReturns := make(chan TaskResult, Tg.getAsyncCount())
	for _, tsk := range Tg.tasks {
		if tsk.Async {
			tsk.Log(" -> exec async \t", tsk.taskName, " id ", tsk.task.RunId)
			waitGroup.Add(1)
			go func(tsk TaskWatched) {
				defer waitGroup.Done()
				taskReturns <- tsk.Run()
			}(tsk)

		} else {
			tsk.Log(" => exec regular \t", tsk.taskName, " id ", tsk.task.RunId)
			tsk.Run()
		}
	}
	Tg.Log("waiting all tasks beeing done")
	go func() {
		waitGroup.Wait()
		close(taskReturns)
	}()
	Tg.Log("all tasks are done")
	return Tg
}

func (Tg *TaskGroup) Log(msg ...interface{}) {
	if Tg.LoggerFnc != nil {
		Tg.LoggerFnc(msg...)
	}
}

// Wait until all task are done, indepenet from any channel and waitgroup blocks
func (Tg *TaskGroup) Wait(wait time.Duration, timeOut time.Duration) {
	var timeOutHit bool = false
	time.AfterFunc(timeOut, func() {
		timeOutHit = true
	})
	for {
		canExists := true
		all := len(Tg.tasks)
		for indx, tsk := range Tg.tasks {
			indOut := indx + 1
			time.Sleep(wait)
			if tsk.IsRunning() {
				canExists = false
				tsk.Log(" x task ", tsk.taskName, " still running ", indOut, "/", all)
			} else {
				tsk.Log(" ✓ task ", tsk.taskName, " DONE ", indOut, "/", all)
			}
		}
		if timeOutHit {
			Tg.Log("Timeout reached. Exit wait")
			return
		}
		if canExists {
			Tg.Log("regular wait exit. all task done")
			return
		}
	}
}
