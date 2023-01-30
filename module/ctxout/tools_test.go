package ctxout_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestClearString(t *testing.T) {
	str := ctxout.StringCleanEscapeCodes("this is a test")
	if str != "this is a test" {
		t.Errorf("Expected 'this is a test' but got '%s'", str)
	}

	str = ctxout.StringCleanEscapeCodes("this is a \x1b[31mtest\x1b[0m")
	if str != "this is a test" {
		t.Errorf("Expected 'this is a test' but got '%s'", str)
	}
	colCode := ctxout.ToString(ctxout.NewMOWrap(), "this is a ", ctxout.BackBlack, "test\n", ctxout.CleanTag)
	str = ctxout.StringPure(colCode)
	if str != "this is a test" {
		t.Errorf("Expected 'this is a test' but got '%s'", str)
	}

	str = ctxout.StringCleanEscapeCodes(ctxout.ToString(ctxout.NewMOWrap(), "this is a ", ctxout.BackBlack, "test", ctxout.CleanTag))
	if str != "this is a test" {
		t.Errorf("Expected 'this is a test' but got '%s'", str)
	}
}
