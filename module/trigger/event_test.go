package trigger_test

import (
	"testing"

	"github.com/swaros/contxt/trigger"
)

func TestEvent(t *testing.T) {

	listen := trigger.NewListener("test", func(any ...interface{}) {
		if string(any[0].(string)) != "Hallo" {
			t.Error("first argument should be Hallo")
		}

		if string(any[1].(string)) != "Welt" {
			t.Error("first argument should be Welt")
		}
	})

	listen.RegisterToEvent("test_1")

	if test1, err := trigger.NewEvent("test_1"); err == nil {

		trigger.UpdateEvents()

		test1.SetArguments("Hallo", "Welt")
		test1.Send()

	} else {
		t.Error(err)
	}
}
