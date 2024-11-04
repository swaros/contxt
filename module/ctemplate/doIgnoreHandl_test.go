package ctemplate_test

import (
	"strings"
	"testing"

	"github.com/swaros/contxt/module/ctemplate"
)

func TestIgnoreHndlBase(t *testing.T) {
	origin := `Hello World
we replacing the word World.
execpt for any masked word, that is defined before to exclude from being replaced.
so we ignore this [World] and this (World) but any other World should be replaced.
	`
	expected := `Hello Mars
we replacing the word Mars.
execpt for any masked word, that is defined before to exclude from being replaced.
so we ignore this [World] and this (World) but any other World should be replaced.
	`

	ignoreHndl := ctemplate.NewIgnorreHndl(origin)
	ignoreHndl.AddIgnores("[World]", "(World)", "other World")
	maskedStr := ignoreHndl.CreateMaskedString()
	maskedStr = strings.ReplaceAll(maskedStr, "World", "Mars")

	restored := ignoreHndl.RestoreOriginalString(maskedStr)
	if restored != expected {
		t.Errorf("Expected\n%s\ngot\n%s", expected, restored)
	}

}
