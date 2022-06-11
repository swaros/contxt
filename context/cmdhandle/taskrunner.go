package cmdhandle

import (
	"errors"
	"sync"
	"time"
)

type timeOutReachedFunc func()
type timerTickFunction func()
type canRunFunc func(*TaskWatched) bool
type getRundIdFunc func() string
type bodyFunc func(*TaskWatched) error
type loggerFunc func(...interface{})

var procTracker sync.Map

type TaskRuntimeState struct {
	AsyncedRunning bool
	Started        bool
	TimedOut       bool
	RunningCount   int
	Done           bool
	RunId          string
}

type TaskWatched struct {
	task             TaskRuntimeState
	taskName         string
	IsGlobalScope    bool
	TaskArgs         map[string]string
	NoErrorIfBlocked bool
	TimeOutTiming    time.Duration
	TimeOutHandler   timeOutReachedFunc
	CanRun           canRunFunc
	TimerTick        timerTickFunction
	GetRunId         getRundIdFunc
	Exec             bodyFunc
	LoggerFnc        loggerFunc
	Async            bool
}

type TaskGroup struct {
	tasks []TaskWatched
	Async bool
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
		if t.CanRun != nil {
			return t.CanRun(t)
		}
		return false
	}
	t.task.Started = true
	t.task.RunningCount = 1
	t.task.Done = false
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

func (t *TaskWatched) Run() error {
	if t.Exec == nil {
		return errors.New("body function exec is undefined")
	}
	// starting the body function and track the execution
	// first we check if can start the task
	if !t.trackStart() {
		if t.NoErrorIfBlocked {
			return nil
		}
		return errors.New("task is already running")
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

	if err := t.Exec(t); err != nil {
		return err
	}

	return nil
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

func (Tg *TaskGroup) Exec() {
	var waitGroup sync.WaitGroup
	errors := make(chan error, len(Tg.tasks))
	for _, tsk := range Tg.tasks {
		if tsk.Async {
			tsk.Log(" -> exec async \t", tsk.taskName, " id ", tsk.task.RunId)
			for _, task := range Tg.tasks {

				waitGroup.Add(1)
				go func(tsk TaskWatched) {
					defer waitGroup.Done()
					err := tsk.Run()
					if err != nil {
						errors <- err
						return
					}

				}(task)
			}

		} else {
			tsk.Log(" => exec regular \t", tsk.taskName, " id ", tsk.task.RunId)
			tsk.Run()
		}
	}
	GetLogger().Debug("waiting task beeing done")
	go func() {
		waitGroup.Wait()
		close(errors)
	}()
	GetLogger().Debug("task done")

}

func (Tg *TaskGroup) Wait() {
	for {
		canExists := true
		all := len(Tg.tasks)
		for indx, tsk := range Tg.tasks {
			if tsk.IsRunning() {
				canExists = false
				tsk.Log(" <-> task ", tsk.taskName, " still running ", indx, "/", all)
			} else {
				tsk.Log(" ✓ task ", tsk.taskName, " DONE ", indx, "/", all)
			}
		}
		if canExists {

			return
		}
	}
}
