package trigger

type Listener struct {
	callback func(any ...interface{})
}

type ListenerContainer struct {
	name     string
	toEvents []string
	callback func(any ...interface{})
}

var trigger map[string]*ListenerContainer = make(map[string]*ListenerContainer)

func ResetAllListener() {
	trigger = make(map[string]*ListenerContainer)
}

func NewListener(name string, calbck func(any ...interface{})) *ListenerContainer {
	trigger[name] = &ListenerContainer{
		name:     name,
		toEvents: []string{},
		callback: calbck,
	}
	return trigger[name]
}

func UpdateEvents() error {
	for _, listener := range trigger {
		for _, eventName := range listener.toEvents {
			return updateEvent(eventName, func(e *Event) error {
				e.ClearListener()
				return e.AddListener(Listener{callback: listener.callback})
			})
		}
	}
	return nil
}

func (trig *ListenerContainer) RegisterToEvent(eventName string) {
	trig.toEvents = append(trig.toEvents, eventName)
}

func (li Listener) Trigger(any []interface{}) {
	li.callback(any...)
}
