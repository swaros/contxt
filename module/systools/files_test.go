package systools_test

import (
	"os"
	"testing"

	"github.com/swaros/contxt/module/systools"
)

func TestIsWriteable(t *testing.T) {
	// get user home
	home, err := os.UserHomeDir()
	if err != nil {
		t.Error(err)
	}
	// just check if the home dir is writeable
	if !systools.IsDirWriteable(home) {
		t.Error("home dir is not writeable")
	}

	// this should not be writeable
	if systools.IsDirWriteable("/root") {
		t.Error("root dir is writeable")
	}
}
