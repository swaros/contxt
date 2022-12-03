package systools_test

import (
	"testing"

	"github.com/swaros/contxt/systools"
)

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
