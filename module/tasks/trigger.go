package tasks

import (
	"github.com/swaros/contxt/module/trigger"
)

type TaskMessage struct {
	// TaskName is the name of the task
	TaskName string
	// TaskID is the id of the task
	TaskID string
	// TaskTarget is the target of the task
	TaskTarget string
	// TaskMessage is the message of the task
	TaskMessage string
	// TaskError is the error of the task
	TaskError error
	// TaskDone is the done state of the task
	TaskDone bool
}

const (
	TaskStatusChange = "task_status_change"
)

type TaskEvent struct {
	updateEvent *trigger.Event
}

func NewTaskEvent() *TaskEvent {
	if updEvt, err := trigger.NewEvent(TaskStatusChange); err != nil {
		panic(err)
	} else {
		return &TaskEvent{
			updateEvent: updEvt,
		}
	}
}

func (t *TaskEvent) GetTaskMessage() TaskMessage {
	return TaskMessage{}
}

func (t *TaskEvent) Init() {

}

func (t *TaskEvent) Update(args ...any) error {
	event, err := trigger.GetEvent(TaskStatusChange)
	if err != nil {
		return err
	} else {
		event.SetArguments(args...)
		return event.Send()
	}

}

func (t *TaskEvent) RegisterListener(name string, callback func(any ...interface{})) error {
	listener := trigger.NewListener(name, callback)
	listener.RegisterToEvent(TaskStatusChange)
	return trigger.UpdateEvents()
}
