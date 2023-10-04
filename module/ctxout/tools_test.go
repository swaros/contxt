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

func TestStringLengthPrintable(t *testing.T) {

	type lenTest struct {
		in  string
		out int
	}

	tests := []lenTest{
		{"this is a test", 14},
		{"this is a \x1b[31mtest\x1b[0m", 14},
		{"this is a \x1b[31mtest\x1b[0m\n", 14},
		{"你好世界", 8},
		{"你好世界\n", 8},
		{"你好世界\x1b[31m\n", 8},
		{"🖵", 1},
		{"🖵\n", 1},
		{"🖵\x1b[31m\t", 1},
		{"🌎\x1b[31m\n", 2},
		{"🌎\x1b[31m\t🌎🖵", 5},
		{"\u2588", 1},
		{"\u2588\n", 1},
	}

	for rnd, test := range tests {
		strLen := ctxout.LenPrintable(test.in)
		if strLen != test.out {
			t.Errorf("[rnd %d] Expected %d but got %d [%s]", rnd, test.out, strLen, test.in)
		}
	}
}
