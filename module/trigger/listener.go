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
