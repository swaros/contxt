package ctxout_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestTermEnvWrap(t *testing.T) {
	te := ctxout.NewTermEnvWrap()

	ctxout.Print(te, "Hello <color>World")
}
