package ctxout_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestCutString(t *testing.T) {
	tbout := ctxout.NewTabOut()
	tHandle := ctxout.NewTableHandle(tbout)
	row := ctxout.NewTabRow(tHandle)
	cell := ctxout.NewTabCell(row)

	cell.SetText("1234567890abcdefghijklmnopqrstuvwxyz")

	// whatever we throw in, cut text should always return excactly the
	// same length as we requested by the parameter
	for i := 1; i < 160; i++ {
		res := cell.CutString(i)
		if len(res) != i {
			t.Errorf("TestCutString: len(['%s']) {%d} != %d) ", res, len(res), i)
		}
	}

}
