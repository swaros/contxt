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
		{"this is a test", 14},                // rnd 0
		{"this is a \x1b[31mtest\x1b[0m", 14}, // rnd 1
		{"你好世界", 8},                           // rnd 2
		{"🖵", 1},                              // rnd 3
		{"🖵\x1b[31m", 1},                      // rnd 4
		{"🌎\x1b[31m🌎🖵", 5},                    // rnd 5
		{"\u2588", 1},                         // rnd 6
	}

	for rnd, test := range tests {
		strLen := ctxout.VisibleLen(test.in)
		if strLen != test.out {
			t.Errorf("[rnd %d] Expected %d but got %d [%s]", rnd, test.out, strLen, test.in)
		} else {
			t.Logf(" OK [rnd %d]", rnd)
		}

	}
}

func TestStringCut(t *testing.T) {

	testStr := "1234567890abcdefghijklmnopqrstuvwxyz"
	expexted := "1234567890"

	cutStr, rest := ctxout.StringCut(testStr, 10)
	if cutStr != expexted {
		t.Errorf("Expected '%s' but got '%s'", expexted, cutStr)
	}
	if rest != "abcdefghijklmnopqrstuvwxyz" {
		t.Errorf("Expected '%s' but got '%s'", "abcdefghijklmnopqrstuvwxyz", rest)
	}

	testStr = "123456"
	expexted = "123456"

	cutStr, rest = ctxout.StringCut(testStr, 10)
	if cutStr != expexted {
		t.Errorf("Expected '%s' but got '%s'", expexted, cutStr)
	}
	if rest != "" {
		t.Errorf("Expected '%s' but got '%s'", "", rest)
	}

}

func TestStringCutRight(t *testing.T) {

	testStr := "1234567890abcdefghijklmnopqrstuvwxyz"
	expexted := "qrstuvwxyz"

	cutStr, rest := ctxout.StringCutFromRight(testStr, 10)
	if cutStr != expexted {
		t.Errorf("Expected '%s' but got '%s'", expexted, cutStr)
	}
	if rest != "1234567890abcdefghijklmnop" {
		t.Errorf("Expected '%s' but got '%s'", "1234567890abcdefghijklmnop", rest)
	}

	testStr = "123456"
	expexted = "123456"

	cutStr, rest = ctxout.StringCutFromRight(testStr, 10)
	if cutStr != expexted {
		t.Errorf("Expected '%s' but got '%s'", expexted, cutStr)
	}
	if rest != "" {
		t.Errorf("Expected '%s' but got '%s'", "", rest)
	}

}
