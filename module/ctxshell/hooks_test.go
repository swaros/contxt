package ctxshell_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxshell"
)

func TestSimpleHookMatch(t *testing.T) {
	hook := ctxshell.NewHook("foo", nil, nil)
	if !hook.Match("foo") {
		t.Error("expected hook to match 'foo'")
	}
}

func TestSimpleHookNoMatch(t *testing.T) {
	hook := ctxshell.NewHook("foo", nil, nil)
	if hook.Match("bar") {
		t.Error("expected hook to not match 'bar'")
	}
}

func TestWildcardHookMatch(t *testing.T) {
	hook := ctxshell.NewHook("foo*", nil, nil)
	if !hook.Match("foobar") {
		t.Error("expected hook to match 'foobar'")
	}
}

func TestWildcardHookNoMatch(t *testing.T) {
	hook := ctxshell.NewHook("foo*", nil, nil)
	if hook.Match("bar") {
		t.Error("expected hook to not match 'bar'")
	}
}
