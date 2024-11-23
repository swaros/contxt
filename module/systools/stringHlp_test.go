package systools_test

import (
	"testing"

	"github.com/swaros/contxt/module/systools"
)

func TestCheckForCleanString(t *testing.T) {
	if str, err := systools.CheckForCleanString("0.4.6"); err != nil {
		t.Error("should be able to handle this string")
	} else {
		if str != "0-4-6" {
			t.Error("unexpectd string ", str)
		}
	}
}

func TestCheckForCleanString2(t *testing.T) {
	type test struct {
		in            string
		out           string
		errorExpected bool
	}

	tests := []test{
		{"0.4.6", "0-4-6", false},
		{"0.4.6-rc1", "0-4-6-rc1", false},
		{"0.4.6-rc1+build1", "", true},
		{"yamama", "yamama", false},
		{"\\m/", "_m_", false},
		{"\033[1;32mCHECK\033[0m", "", true},
		{"??.'", "", true},
	}

	for i, v := range tests {
		str, err := systools.CheckForCleanString(v.in)
		if err != nil && !v.errorExpected {
			t.Error("unexpected error: [", err, "] for string ", v.in, " test ", i)
		} else if err == nil && v.errorExpected {
			t.Error("expected error, got none. for string ", v.in, " test ", i)
		}
		if str != v.out {
			t.Error("unexpected string: [", str, "] for string ", v.in, " test ", i)
		}
	}
}

func TestPrintableString(t *testing.T) {
	type test struct {
		in  string
		out string
	}

	tests := []test{
		{"0.4.6", "0.4.6"},
		{"0.4.6-rc1", "0.4.6-rc1"},
		{"0.4.6-rc1+build1", "0.4.6-rc1+build1"},
		{"yamama", "yamama"},
		{"\\m/", "\\m/"},
		{"\033[1;32mCHECK\033[0m", "[1;32mCHECK[0m"},
		{"??.'", "??.'"},
		{"\033[1;32mCHECK\033[0m\033[1;31mCHECK\033[0m\033[1;33mCHECK\033[0m", "[1;32mCHECK[0m[1;31mCHECK[0m[1;33mCHECK[0m"},
	}

	for i, v := range tests {
		str := systools.PrintableChars(v.in)
		if str != v.out {
			t.Error("unexpected string: [", str, "] for string ", v.in, " test ", i)
		}
	}
}

func TestNoEscapeString(t *testing.T) {

	type test struct {
		in  string
		out string
	}

	tests := []test{
		{"0.4.6", "0.4.6"},
		{"0.4.6-rc1", "0.4.6-rc1"},
		{"0.4.6-rc1+build1", "0.4.6-rc1+build1"},
		{"yamama", "yamama"},
		{"\\m/", "\\m/"},
		{"\033[1;32mCHECK\033[0m", "CHECK"},
		{"??.'", "??.'"},
		{"\033[1;32mCHECK\033[0m\033[1;31mCHECK\033[0m\033[1;33mCHECK\033[0m", "CHECKCHECKCHECK"},
	}

	for i, v := range tests {
		str := systools.NoEscapeSequences(v.in)
		if str != v.out {
			t.Error("unexpected string: [", str, "] for string ", v.in, " test ", i)
		}
	}
}

func TestShortLabel(t *testing.T) {
	testChars := []string{
		"hello-my-friend",
		"hello_my_friend",
		"hello my friend",
		"hello.my.friend",
	}
	for _, v := range testChars {
		str := systools.ShortLabel(v, 3)
		if str != "hmf" {
			t.Error("unexpected string: [", str, "] for string ", v)
		}
	}
	if str := systools.ShortLabel("hello-my-friend", 20); str != "hmf" {
		t.Error("unexpected string: [", str, "] for string hello-my-friend")
	}
	if str := systools.ShortLabel("this-is-the first\t-day...of_the\tcentury", 50); str != "titfdotc" {
		t.Error("unexpected string: [", str, "] for string hello-my-friend")
	}

	if str := systools.ShortLabel("this-is-the first\t-day...of_the\tcentury", 3); str != "tit" {
		t.Error("unexpected string: [", str, "] for string hello-my-friend")
	}
	//\033[1;32mCHECK\033[0m
	if str := systools.ShortLabel("\033[1;32mC.H.E.C.K\033[0m", 20); str != "CHECK" {
		t.Error("unexpected string: [", str, "] for string \033[1;32mCHECK\033[0m")
	}
}

func TestFindStartChars(t *testing.T) {
	// Basic test case
	if chars := systools.FindStartChars("hello world"); chars != "hw" {
		t.Error("unexpected chars: [", chars, "] for string hello world")
	}

	// UTF-8 characters
	if chars := systools.FindStartChars("你好，世界"); chars != "" {
		t.Error("unexpected chars: [", chars, "] for string 你好，世界")
	}

	// New lines and tabs
	if chars := systools.FindStartChars("hello\nworld\t!"); chars != "hw" {
		t.Error("unexpected chars: [", chars, "] for string hello\nworld\t!")
	}

	// Escape sequences
	if chars := systools.FindStartChars("hello\\nworld\\t!"); chars != "hnt" {
		t.Error("unexpected chars: [", chars, "] for string hello\\nworld\\t!")
	}

	// Special characters
	if chars := systools.FindStartChars("this-is-the first\t-day...of_the\tcentury"); chars != "titfdotc" {
		t.Error("unexpected chars: [", chars, "] for string this-is-the first\t-day...of_the\tcentury")
	}

	// ANSI escape sequences
	if chars := systools.FindStartChars("\033[1;32mCHECK\033[0m"); chars != "130" {
		t.Error("unexpected chars: [", chars, "] for string \033[1;32mCHECK\033[0m")
	}

	// Long string with special characters
	if chars := systools.FindStartChars("bhd sdkh lskshl .lshlfh  lsjjh lkskhg the-wild-wild-world the-wild-wild-world"); chars != "bslllltwwwtwww" {
		t.Error("unexpected chars: [", chars, "] for string bhd sdkh lskshl .lshlfh  lsjjh lkskhg the-wild-wild-world the-wild-wild-world")
	}
}

func TestStringSplitArgs(t *testing.T) {
	strRes, mapRes := systools.StringSplitArgs("command check01 data last ", "arg")
	if strRes != "command" {
		t.Error("unexpected string: [", strRes, "]")
	}
	expectedMap := map[string]string{
		"arg0": "command",
		"arg1": "check01",
		"arg2": "data",
		"arg3": "last",
	}
	if len(mapRes) != len(expectedMap) {
		t.Error("unexpected map length: [", len(mapRes), "]")
	} else {
		for k, v := range mapRes {
			if expectedMap[k] != v {
				t.Error("unexpected map value: [", v, "] for key ", k)
			}
		}
	}

}

func TestStrLen(t *testing.T) {
	len := systools.StrLen("hello world")

	if len != 11 {
		t.Error("invalid length", len)
	}

	len = systools.StrLen("")
	if len != 0 {
		t.Error("invalid length", len)
	}
}

func TestPathStringSub(t *testing.T) {
	str := systools.StringSubLeft("this is the string", 20)
	if str != "this is the string" {
		t.Error("the string should being changed, because it is shorter then max")
	}

	str = systools.StringSubLeft("the wild wild world", 10)
	if systools.StrLen(str) != 10 {
		t.Error("expected lenght is 10, got ", systools.StrLen(str))
	}
	if str != "the wild w" {
		t.Error("string should be reduced: ", str)
	}

	str = systools.StringSubRight("the-wild-wild-world", 10)
	if systools.StrLen(str) != 10 {
		t.Error("expected lenght is 10, got ", systools.StrLen(str))
	}
	if str != "wild-world" {
		t.Error("string should be reduced but starts from right: [", str, "]", systools.StrLen(str))
	}

	str = systools.StringSubRight("bhd sdkh lskshl .lshlfh  lsjjh lkskhg the-wild-wild-world the-wild-wild-world", 20)
	if systools.StrLen(str) != 20 {
		t.Error("expected lenght is 10, got ", systools.StrLen(str))
	}
	if str != " the-wild-wild-world" {
		t.Error("string should be reduced but starts from right: [", str, "]", systools.StrLen(str))
	}

	str = systools.StringSubRight("test", 20)
	if str != "test" {
		t.Error("the string should being changed, because it is shorter then max")
	}

}

func TestStringTrimAllSpaces(t *testing.T) {
	str := systools.TrimAllSpaces("hello   my friend")
	if str != "hello my friend" {
		t.Error("string should be reduced: ", str)
	}
}

func TestFillString(t *testing.T) {
	str := systools.FillString("hello", 10)
	if str != "hello     " {
		t.Error("string should be filled: ", str)
	}

	str = systools.FillString("hello", 5)
	if str != "hello" {
		t.Error("string should be filled: ", str)
	}

	str = systools.FillString("hello", 0)
	if str != "hello" {
		t.Error("string should not be changed, because it is already longer then 0: ", str)
	}
}

func TestStrContains(t *testing.T) {
	tests := []struct {
		in  string
		sub string
		out bool
	}{
		{"hello world", "world", true},
		{"hello world", "worlds", false},
		{"hello world", "hello", true},
		{"hello world", "hello world", true},
		{"hello world", "hello world ", false},
		{"{{- if}}testdata{{- end }}", "{{- if}}testdata{{- end }}", true},
		{"check outer {{- if}}testdata{{- end }}sample", "{{- if}}testdata{{- end }}", true},
		{"check outer {{- if}}testdata{{- end }}sample", "{{- if}}testdata{{- end }}sample", true},
		{"check outer {{- if}}testdata{{- end }}sample", "testdata ", false},
		{"'\"'masken\"'\"'", "\"'masken\"'\"", true},
		// escape sequences
		{"\033[1;32mCHECK\033[0m", "CHECK", true},
		{"\033[1;32mCHECK\033[0m", "CHECK ", false},
		// utf8
		{" 你好世界", "你好", true},
		{" 你好世界", "你好 ", false},
		{" 你好世界", "你好世界", true},
		// special chars
		{"\\m/", "\\m/", true},
		{"\\m/", "\\m", true},
		{"\\m/", "m", true},
		{"\\m/", "m/", true},
		// empty strings
		{"", "", true},
		{"", " ", false},
		{" ", "", false},
		{" ", " ", true},
		// spaces
		{"hello world", " ", true},
		{"hello world", "  ", false},
		{"hello world", "   ", false},
		// tabs
		{"hello world", "\t", false},
		{"hello world", "\t\t", false},
		{"hello world", "\t\t\t", false},
		{"hello\tworld", "\t", true},
		{"hello\tworld", "\t\t", false},
		{"hello\tworld", "\t\t\t", false},
		// newlines
		{"hello world", "\n", false},
		{"hello world", "\n\n", false},
		{"hello\nworld", "\n", true},
		{"hello\nworld", "\nworld", true},
	}

	for i, v := range tests {
		if systools.StrContains(v.in, v.sub) != v.out {
			t.Error("unexpected result for test ", i, " [", v.in, "] [", v.sub, "]")
		}
	}
}

func TestSplitQuoted(t *testing.T) {
	cmdStr := "command 'check01 data last' 'new world data' test"
	cmds := systools.SplitQuoted(cmdStr, " ")
	if len(cmds) != 4 {
		t.Error("unexpected length ", len(cmds))
	}
	if cmds[0] != "command" {
		t.Error("unexpected string ", cmds[0])
	}
	if cmds[1] != "check01 data last" {
		t.Error("unexpected string ", cmds[1])
	}
	if cmds[2] != "new world data" {
		t.Error("unexpected string ", cmds[2])
	}
	if cmds[3] != "test" {
		t.Error("unexpected string ", cmds[3])
	}
	cmdStr = "command check"
	cmds = systools.SplitQuoted(cmdStr, " ")
	if len(cmds) != 2 {
		t.Error("unexpected length ", len(cmds))
	}
	if cmds[0] != "command" {
		t.Error("unexpected string ", cmds[0])
	}
	if cmds[1] != "check" {
		t.Error("unexpected string ", cmds[1])
	}

}

func TestAnyToStrNoTabs(t *testing.T) {
	str := systools.AnyToStrNoTabs("hello\tworld")
	if str != "hello world" {
		t.Error("unexpected string ", str)
	}
}
