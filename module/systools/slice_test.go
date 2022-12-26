package systools_test

import (
	"testing"

	"github.com/swaros/contxt/module/systools"
)

func TestContains(t *testing.T) {
	slice := []string{"hello", "world"}

	if systools.SliceContains(slice, "yolo") {
		t.Error("yolo is not on the slice")
	}

	if !systools.SliceContains(slice, "world") {
		t.Error("world should be found")
	}
}
