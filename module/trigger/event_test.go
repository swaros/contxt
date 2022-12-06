package trigger_test

import (
	"testing"

	"github.com/swaros/contxt/module/trigger"
)

func TestEvent(t *testing.T) {
	trigger.ResetAllEvents()
	// creata a listener for an event named test
	listen := trigger.NewListener("test", func(any ...interface{}) {
		// parsing the argument 0 and expecting a string
		if string(any[0].(string)) != "Hallo" {
			t.Error("first argument should be Hallo")
		}
		// parsing argument 1 and expecting string too
		if string(any[1].(string)) != "Welt" {
			t.Error("first argument should be Welt")
		}
	})

	// register the listener to the event listener
	listen.RegisterToEvent("test_1")

	// trigger the event
	if test1, err := trigger.NewEvent("test_1"); err == nil {

		trigger.UpdateEvents()

		test1.SetArguments("Hallo", "Welt")
		test1.Send()

	} else {
		t.Error(err)
	}
}

func TestMultiple(t *testing.T) {
	trigger.ResetAllEvents()
	// # LISTENER
	// here we define the listener first
	baseListenerA := false                                               // this is just our flag to find out if the method was ever called
	baseA := trigger.NewListener("trigger_a", func(any ...interface{}) { // creating a new listener named trigger_a
		baseListenerA = true // if the callback is executed, we set the flag to true
	})
	baseA.RegisterToEvent("testa") // now we register this trigger to an event, that is currently not exists

	// # EVENTS
	var events []string = []string{"testa", "testb"}               // this will be the list of events we will create
	aEvent := false                                                // flag for callback a
	bEvent := false                                                // ...and b
	eventErr := trigger.NewEvents(events, func(e *trigger.Event) { // executes the setup callbacks for the events
		if e.GetName() == "testa" {
			aEvent = true
		}
		if e.GetName() == "testb" {
			bEvent = true
		}
	})

	if eventErr != nil { // just test no errors so far
		t.Error(eventErr)
	}

	if !aEvent {
		t.Error("a event not triggered")
	}
	if !bEvent {
		t.Error("b event not triggered")
	}

	trigger.UpdateEvents()                                  // update the evenst so triggers can be assigned to the events
	if evnt, err := trigger.GetEvent("testa"); err == nil { // get the event named 'testa'
		evnt.Send() // and trigger it. now the listener method should be executed
	} else {
		t.Error(err)
	}

	if !baseListenerA { // check if the listener method is executed
		t.Error("listener A not executed")
	}

}

func TestErrorCases(t *testing.T) {
	trigger.ResetAllEvents()
	if _, err := trigger.NewEvent("ax"); err != nil {
		t.Error(err)
	}

	if _, err := trigger.NewEvent("ax"); err == nil {
		t.Error("creating the same event twice should create an error")
	}

	var events []string = []string{"ba", "ba"}
	eventErr := trigger.NewEvents(events, func(e *trigger.Event) {})
	if eventErr == nil {
		t.Error("we have defined a event twice, this should end up in an error")
	}

	tmp := trigger.NewListener("check", func(any ...interface{}) {})
	tmp.RegisterToEvent("yolo")
	if err := trigger.UpdateEvents(); err == nil {
		t.Error("update events should find the binding to an non existing event")
	}

	if _, err := trigger.GetEvent("wtf"); err == nil {
		t.Error("this event not exists. have to create an error")
	}
}
