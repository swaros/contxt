// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

 package trigger

import (
	"errors"
	"sync"
)

type Event struct {
	name         string
	args         []interface{}
	listenerLst  []Listener
	errorInChain error
}

// NewEvent creates a new Event and register it
func NewEvent(name string) (*Event, error) {
	evt := &Event{name: name}    // new basic event with the name only
	if ok := addEvent(evt); ok { // we do not overwrite existing events. this is handled by addEvent
		return evt, nil
	}
	return evt, errors.New("event with the same name is already registered")
}

// NewEvents just wraps the NewEvent for multipe events
// and executes an callback for each of the events.
// this way we can stick to a more generic solution for a couple of events
func NewEvents(names []string, cb func(*Event)) error {
	for _, name := range names {
		if evnt, err := NewEvent(name); err == nil {
			cb(evnt)
		} else {
			return err
		}
	}
	return nil
}

// GetName just returns the name of the event
func (event *Event) GetName() string {
	return event.name
}

// AddListener adds listener to the event
func (event *Event) AddListener(lst ...Listener) error {
	if _, found := eventMapStore.Load(event.name); found {
		event.listenerLst = append(event.listenerLst, lst...)
		eventMapStore.Store(event.name, event)
		return nil
	}
	return errors.New("this event is not registered")
}

func (event *Event) ClearListener() error {
	if _, found := eventMapStore.Load(event.name); found {
		event.listenerLst = []Listener{}
		eventMapStore.Store(event.name, event)
		return nil
	}
	return errors.New("this event is not registered")
}

// SetArguments adds or changes the arguments for the event
func (event *Event) SetArguments(args ...interface{}) error {
	if _, found := eventMapStore.Load(event.name); found {
		event.args = args
		eventMapStore.Store(event.name, event)
		return nil
	}
	return errors.New("this event is not registered")
}

// Send calls any assigned callback
func (event *Event) Send() error {
	if event.errorInChain != nil {
		return event.errorInChain
	}
	if event.listenerLst == nil || len(event.listenerLst) == 0 {
		return errors.New("no listener for this event")
	}
	for _, listen := range event.listenerLst {
		listen.Trigger(event.args)
	}
	return nil
}

// eventMapStore contains any registered event
var eventMapStore sync.Map

// addEvent adds a new event. it will not override
// exiting events. if an event already exists
// it returns false. otherwise true is returned
func addEvent(event *Event) bool {
	if _, found := eventMapStore.Load(event.name); found {
		return false
	}
	eventMapStore.Store(event.name, event)
	return true
}

func GetEvent(name string) (*Event, error) {
	if evt, found := eventMapStore.Load(name); found {
		return evt.(*Event), nil
	}
	return nil, errors.New("Event not exists " + name)
}

func updateEvent(eventName string, updateCallBack func(*Event) error) error {
	if evt, found := eventMapStore.Load(eventName); found {
		event := evt.(*Event)
		updateCallBack(event)
		eventMapStore.Store(event.name, event)
	} else {
		return errors.New("event " + eventName + " not exists")
	}
	return nil
}

func ResetAllEvents() {
	eventMapStore.Range(func(key, value any) bool {
		eventMapStore.Delete(key)
		return true
	})
	ResetAllListener()
}
