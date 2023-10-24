package runner_test

import (
	"testing"

	"github.com/swaros/contxt/module/runner"
)

func TestGetFPath(t *testing.T) {
	zsh := runner.NewZshHelper()
	zPath, err := zsh.GetBinPath()
	if err != nil || zPath == "" {
		t.Log("skipped zsh Testing, because it seems zsh not being installed.")
		t.SkipNow()
	} else {
		if fp, ferr := zsh.GetFirstFPath(); ferr != nil {
			// test also skipped if no first path is found.
			// it is too much depending on the system to throw an error on this case.
			t.SkipNow()
		} else {
			// this is close the only test we can do.
			// we can not test if the fpath is correct, because it is too much depending on the system.
			if fp == "" {
				t.Error("FirstFPath is empty")
			}
		}
	}
}
