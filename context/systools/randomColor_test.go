package systools_test

import (
	"testing"

	"github.com/swaros/contxt/context/systools"
)

func TestPrintColoredChanges(t *testing.T) {
	for i := 0; i < 40; i++ {
		var last string
		colorCode, _ := systools.CreateColorCode()

		if i > 1 {
			if last == colorCode {
				t.Error("colorcode is the same again", colorCode, "prevoius", last)
			}
		}
	}

}
