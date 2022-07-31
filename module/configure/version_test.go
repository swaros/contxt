package configure_test

import (
	"testing"

	"github.com/swaros/contxt/configure"
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
