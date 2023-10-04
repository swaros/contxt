package taskrun_test

import (
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/taskrun"
)

func TestFindPwrShellModule(t *testing.T) {
	if configure.GetOs() != "windows" {
		t.Skip("skipping test in non windows os")
	}
	found, _ := taskrun.FindPwrShellProfile()
	if !found {
		t.Errorf("powershell module not found")
	}
}
