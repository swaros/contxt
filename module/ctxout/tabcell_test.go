package ctxout_test

import (
	"strings"
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

func GetWithSpecialChars(s string) string {
	s = strings.ReplaceAll(s, "\n", "[\\n]")
	s = strings.ReplaceAll(s, "\t", "[\\t]")
	s = strings.ReplaceAll(s, "\r", "[\\r]")
	return s
}

func TestCutStringWithNewLine(t *testing.T) {
	tbout := ctxout.NewTabOut()
	tHandle := ctxout.NewTableHandle(tbout)
	row := ctxout.NewTabRow(tHandle)
	cell := ctxout.NewTabCell(row)

	cell.SetText("1234567890abcdef\nghijkl\nmnopqrstuvwxyz")
	cell.SetOverflowMode("wordwrap")

	outcome := cell.CutString(10)
	overlow := cell.GetOverflowContent()

	if outcome != "1234567890" {
		t.Errorf("TestCutStringWithNewLine: expected '1234567890' got '%s'", outcome)
	}

	expexcted := "abcdef\nghijkl\nmnopqrstuvwxyz"
	if overlow != "abcdef\nghijkl\nmnopqrstuvwxyz" {
		t.Errorf("TestCutStringWithNewLine: expected\\got\n'%s'\n'%s'", GetWithSpecialChars(expexcted), GetWithSpecialChars(overlow))
	}

}

func assertTextCut(t *testing.T, content string, size int, warp string, expected string) {
	t.Helper()
	tbout := ctxout.NewTabOut()
	tHandle := ctxout.NewTableHandle(tbout)
	row := ctxout.NewTabRow(tHandle)
	cell := ctxout.NewTabCell(row)

	cell.SetText(content)
	cell.SetOverflowMode(warp)

	outcome := cell.CutString(size)

	outcomeForPrint := strings.ReplaceAll(outcome, "\n", "\\n")

	if expected != "" && outcome != expected {
		t.Errorf("fail: expected\n\t[%s]\n but got\n\t[%s]", expected, outcomeForPrint)
	}

	if len(outcome) != size {
		t.Error("fail: expected size", size, "but got", len(outcome))
	}
}

func TestCutStringWithNewLine2(t *testing.T) {
	assertTextCut(t, "1234567890abcdef\nghijkl\nmnopqrstuvwxyz", 10, "wordwrap", "1234567890")
	assertTextCut(t, "hello\nWorld on \nFire", 20, "wordwrap", "hello               ")
}

func assertWordWrap(t *testing.T, content string, size int, wrapMode string) {
	t.Helper()
	tbout := ctxout.NewTabOut()
	tHandle := ctxout.NewTableHandle(tbout)
	row := ctxout.NewTabRow(tHandle)
	cell := ctxout.NewTabCell(row)

	cell.SetOverflowMode(wrapMode)
	cell.SetText(content)
	newCont, _ := cell.WrapText(size)

	if len(newCont) != size {
		printable := strings.ReplaceAll(newCont, "\n", "\\n")
		t.Error("fail: expected size", size, "but got", len(content), "\n --> (", printable, ")")
	}
}

func TestWordWrapImpl(t *testing.T) {
	assertWordWrap(t, "small", 20, ctxout.OfWordWrap)
	assertWordWrap(t, "1234567890abcdefghijklmnopqrstuvwxyz", 10, ctxout.OfIgnore)
	assertWordWrap(t, "small", 20, ctxout.OfIgnore)
	assertWordWrap(t, "1234567\n890abcdefghijklmn\nopqrstuvwxyz", 10, ctxout.OfIgnore)
	assertWordWrap(t, "1234567890abcdefghijklmnopqrstuvwxyz", 10, ctxout.OfWordWrap)
	assertWordWrap(t, "1234567\n890abcdefghijklmn\nopqrstuvwxyz", 10, ctxout.OfWordWrap)
}

func TestOverflowIgnore(t *testing.T) {
	tbout := ctxout.NewTabOut()
	tHandle := ctxout.NewTableHandle(tbout)
	row := ctxout.NewTabRow(tHandle)
	cell := ctxout.NewTabCell(row)

	cell.SetText("this text will be to long for the cell")
	cell.SetOverflowMode(ctxout.OfIgnore)

	outcome := cell.CutString(10)

	expected := "this t ..."
	if outcome != expected {
		t.Errorf("TestCutStringWithNewLine: expected '%s' got '%s'", expected, outcome)
	}

}

func TestOverflowIgnoreToRight(t *testing.T) {
	tbout := ctxout.NewTabOut()
	tHandle := ctxout.NewTableHandle(tbout)
	row := ctxout.NewTabRow(tHandle)
	cell := ctxout.NewTabCell(row)

	cell.SetText("this text will be to long for the cell")
	cell.SetOverflowMode(ctxout.OfIgnore)
	cell.SetOrigin(ctxout.OriginRight)

	outcome := cell.CutString(10)

	expected := " ...e cell"
	if outcome != expected {
		t.Errorf("TestCutStringWithNewLine: expected '%s' got '%s'", expected, outcome)
	}

}
