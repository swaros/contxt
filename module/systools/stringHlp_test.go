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

}
