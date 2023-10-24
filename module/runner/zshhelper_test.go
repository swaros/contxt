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
			t.Error(ferr)
		} else {
			if fp == "" {
				t.Error("FirstFPath is empty")
			}
		}
	}
}
