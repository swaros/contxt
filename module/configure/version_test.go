package configure_test

import (
	"runtime"
	"strings"
	"testing"

	"github.com/swaros/contxt/module/configure"
)

func TestGetOs(t *testing.T) {
	versionStr := configure.GetOs()
	if versionStr == "" {
		t.Error("versionstring should not being empty")
	}
}

func TestVersion(t *testing.T) {

	if configure.CheckVersion("0.3", "0.2.1") == true {
		t.Error("did we reach version 0.3 already? ", configure.GetVersion())
	}
}

func TestDefaultValues(t *testing.T) {
	if configure.GetBuild() != "" {
		t.Error("unexpected default value")
	}

	if configure.GetVersion() != "" {
		t.Error("unexpected default value")
	}

	// This is the most useless test i wrote in my life.
	// but i could not resist just because of the red color
	// in the coverage
	if configure.GetOs() != strings.ToLower(runtime.GOOS) {
		t.Error("what the .... how could this fail?")
	}
}
