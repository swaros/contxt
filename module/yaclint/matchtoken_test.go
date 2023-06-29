package yaclint_test

import (
	"testing"

	"github.com/swaros/contxt/module/yaclint"
)

func TestDetectValue(t *testing.T) {
	str := "\"test\""

	detected := yaclint.DetectedValueFromString(str)
	if detected != str {
		t.Errorf("we should get the same value back, but got %s", detected)
	}

	str = " true"
	detectedV := yaclint.DetectedValueFromString(str)
	if detectedV != true {
		t.Errorf("we should get an boolean (true) but got:%v", detectedV)
	}

}

func TestMultiple(t *testing.T) {
	type testValueDtct struct {
		in  string
		out interface{}
	}

	testValues := []testValueDtct{
		{"\" false\"", "\" false\""},
		{"\"true\"", "\"true\""},
		{" true", true},
		{" false", false},
		{"\"1.44\"", "\"1.44\""},
		{" 1.44", 1.44},
		{"\"1\"", "\"1\""},
		{" 1", 1},
	}

	for _, test := range testValues {
		detectedV := yaclint.DetectedValueFromString(test.in)
		if detectedV != test.out {
			t.Errorf("we expected (%v) but got :(%v)", test.out, detectedV)
		}
	}

}

func TestMatchToString(t *testing.T) {
	match := yaclint.MatchToken{
		UuId:       "test1234567",
		KeyWord:    "test",
		Value:      "test",
		SequenceNr: 1,
		Type:       "string",
	}

	str := match.ToString()
	if str != "[-] test (): [0] val[test] indx[0] seq[1] (string)" {
		t.Errorf("unexpected string: %s", str)
	}

	pairToken := yaclint.MatchToken{
		UuId:       "test1234567_2",
		KeyWord:    "test2",
		Value:      "test2",
		SequenceNr: 1,
		Type:       "string",
	}

	match.PairToken = &pairToken
	str = match.ToString()
	if str != "[-] test (): [0] val[test] (test2 ())pval[test2] indx[0] seq[1] (string)" {
		t.Errorf("unexpected string: '%s'", str)
	}

	match.Value = 1
	match.VerifyValue()
	str = match.ToString()
	if str != "[-] test (): [0] val[1] (test2 ())pval[test2] indx[0] seq[1] (string)" {
		t.Errorf("unexpected string: %s", str)
	}

	match.Value = 1.44
	match.VerifyValue()
	str = match.ToString()
	if str != "[-] test (): [0] val[1.44] (test2 ())pval[test2] indx[0] seq[1] (string)" {
		t.Errorf("unexpected string: %s", str)
	}

	match.Value = true
	match.VerifyValue()
	str = match.ToString()
	if str != "[-] test (): [0] val[true] (test2 ())pval[test2] indx[0] seq[1] (string)" {
		t.Errorf("unexpected string: %s", str)
	}

	match.Value = false
	match.VerifyValue()
	str = match.ToString()
	if str != "[-] test (): [0] val[false] (test2 ())pval[test2] indx[0] seq[1] (string)" {
		t.Errorf("unexpected string: %s", str)
	}

	match.Value = "1.44"
	match.VerifyValue()
	str = match.ToString()
	if str != "[-] test (): [0] val[1.44] (test2 ())pval[test2] indx[0] seq[1] (string)" {
		t.Errorf("unexpected string: %s", str)
	}
}
