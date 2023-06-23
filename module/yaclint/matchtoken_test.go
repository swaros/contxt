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
