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
