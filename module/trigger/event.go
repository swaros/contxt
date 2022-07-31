package trigger

import (
	"errors"
	"sync"
)

type Event struct {
	name     string
	args     []interface{}
	listener []Listener
}

// NewEvent creates a new Event and register it
func NewEvent(name string) (*Event, error) {
	evt := &Event{name: name}    // new basic event with the name only
	if ok := addEvent(evt); ok { // we do not overwrite existing events. this is handled by addEvent
		return evt, nil
	}
	return evt, errors.New("event with the same name is already registered")
}

// AddListener adds listener to the event
func (event *Event) AddListener(lst ...Listener) error {
	if _, found := eventMap.Load(event.name); found {
		event.listener = append(event.listener, lst...)
		eventMap.Store(event.name, event)
		return nil
	}
	return errors.New("this event is not registered")
}

func (event *Event) ClearListener() error {
	if _, found := eventMap.Load(event.name); found {
		event.listener = []Listener{}
		eventMap.Store(event.name, event)
		return nil
	}
	return errors.New("this event is not registered")
}

// SetArguments adds or changes the arguments for the event
func (event *Event) SetArguments(args ...interface{}) error {
	if _, found := eventMap.Load(event.name); found {
		event.args = args
		eventMap.Store(event.name, event)
		return nil
	}
	return errors.New("this event is not registered")
}

func (event *Event) Send() {
	for _, listen := range event.listener {
		listen.Trigger(event.args)
	}
}

// eventMap contains any registered event
var eventMap sync.Map

// addEvent adds a new event. it will not override
// exiting events. if an event already exists
// it returns false. otherwise true is returned
func addEvent(event *Event) bool {
	if _, found := eventMap.Load(event.name); found {
		return false
	}
	eventMap.Store(event.name, event)
	return true
}

func updateEvent(eventName string, updateCallBack func(*Event)) {
	if evt, found := eventMap.Load(eventName); found {
		event := evt.(*Event)
		updateCallBack(event)
		eventMap.Store(event.name, event)
	}
}
