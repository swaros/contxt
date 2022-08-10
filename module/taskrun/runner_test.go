package taskrun_test

import (
	"testing"

	"github.com/swaros/contxt/taskrun"
)

func TestTargetAsMapUnique(t *testing.T) {
	strMap := []string{"one", "two", "three"}

	if taskrun.ExistInStrMap("four", strMap) {
		t.Error("four is not exists in map. but got true on check")
	}

	if !taskrun.ExistInStrMap("two", strMap) {
		t.Error("two is exists in map. but got false on check")
	}

	if !taskrun.ExistInStrMap("two ", strMap) {
		t.Error("two is exists in map. but got false on check. we checked 'two ' but the space should be Trim")
	}
}
