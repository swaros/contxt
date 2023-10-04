package ctxout_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestManOutWrap(t *testing.T) {
	mo := ctxout.NewMOWrap()

	msg := ctxout.ToString(mo, "Hello <f:red>World</>")
	if !mo.GetInfo().NoColored {
		t.Error("Color should be disabled")
	}
	if msg != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%v'", msg)
	}
	// forced color
	ctxout.SetBehavior(ctxout.CtxOutBehavior{NoColored: false})
	msg = ctxout.ToString(mo, "Hello <f:red>World</>")
	if msg != "Hello \033[31mWorld\033[0m" {
		t.Errorf("Expected 'Hello World', got '%v'", msg)
	}
}
