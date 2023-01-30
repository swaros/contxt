package ctxout_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestManOutWrap(t *testing.T) {
	mo := ctxout.NewMOWrap()
	//ctxout.AddPostFilter(mo)

	ctxout.PrintLn(mo, "Hello <f:red>World</>")
}
