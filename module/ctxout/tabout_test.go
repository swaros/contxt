package ctxout_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestBasicTabout(t *testing.T) {
	to := ctxout.NewTabOut()
	output := to.Command("<row><tab size='23'>this is a test</tab><tab size='25' origin='2'>and this is another test</tab></row>")

	expect := "this is a test          and this is another test"
	if output != expect {
		t.Errorf("Expected '%s' but got '%s'", expect, output)
	}
}

func TestGettingEscape(t *testing.T) {
	escapInStr := "Hello \033[33m" + "World" + "\033[0m"

	lastEscape := ctxout.GetLastEscapeSequence(escapInStr)

	if lastEscape != "\033[0m" {
		t.Errorf("Expected '%s' but got '%s'", "\033[0m", lastEscape)
	}
}
