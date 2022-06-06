package cmdhandle_test

import (
	"testing"

	"github.com/swaros/contxt/context/cmdhandle"
)

func TestTargetAsMapUnique(t *testing.T) {
	strMap := []string{"one", "two", "three"}

	if cmdhandle.ExistInStrMap("four", strMap) {
		t.Error("four is not exists in map. but got true on check")
	}

	if !cmdhandle.ExistInStrMap("two", strMap) {
		t.Error("two is exists in map. but got false on check")
	}

	if !cmdhandle.ExistInStrMap("two ", strMap) {
		t.Error("two is exists in map. but got false on check. we checked 'two ' but the space should be Trim")
	}
}
