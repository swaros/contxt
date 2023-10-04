package tasks_test

import (
	"testing"

	"github.com/swaros/contxt/module/tasks"
)

func TestBase(t *testing.T) {
	ph := tasks.NewDefaultPhHandler()
	ph.SetPH("test", "res-test")
	ph.SetPH("test2", "res-test2")
	ph.SetPH("test3", "res-test3")
	if ph.GetPH("test") != "res-test" {
		t.Error("test failed")
	}
	if ph.GetPH("test2") != "res-test2" {
		t.Error("test failed")
	}
	if ph.GetPH("test3") != "res-test3" {
		t.Error("test failed")
	}
	if ph.GetPH("test4") != "" {
		t.Error("test failed")
	}
}

func TestHandlePlaceHolder(t *testing.T) {
	ph := tasks.NewDefaultPhHandler()
	ph.SetPH("test", "res-test")

	assertStrEqual(t, "hello res-test", ph.HandlePlaceHolder("hello test"))
	scopemap := make(map[string]string)
	scopemap["test"] = "scoped-test"
	assertStrEqual(t, "hello scoped-test", ph.HandlePlaceHolderWithScope("hello test", scopemap))
}
