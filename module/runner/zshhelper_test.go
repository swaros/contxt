package runner_test

import (
	"testing"

	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/runner"
	"github.com/swaros/contxt/module/systools"
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
			t.Log("skipped zsh Testing, because we did not get any usefull path.")
			t.SkipNow()
		} else {
			// this is close the only test we can do.
			// we can not test if the fpath is correct, because it is too much depending on the system.
			if fp == "" {
				t.Error("FirstFPath is empty")
			}

			if !systools.IsDirWriteable(fp) {
				t.Error("FirstFPath is not writeable")
			}

		}
	}

}

func TestExistingPath(t *testing.T) {
	zsh := runner.NewZshHelper()
	zPath, err := zsh.GetBinPath()
	if err != nil || zPath == "" {
		t.Log("skipped zsh Testing, because it seems zsh not being installed.")
		t.SkipNow()
	} else {
		firstExisting, err := zsh.GetFirstExistingPath()
		if err != nil {
			t.SkipNow()
		}
		if firstExisting == "" {
			t.Error("FirstExistingPath is empty")

		} else {
			if exists, err := dirhandle.Exists(firstExisting); err != nil || !exists {
				t.Log("FirstExistingPath is not existing")
			}
		}
	}
}
