package systools_test

import (
	"strings"
	"testing"

	"github.com/swaros/contxt/context/systools"
)

func TestRandomColor(t *testing.T) {
	colorCode := systools.CreateColor()
	if len(colorCode) != 9 {
		t.Error("colorcode have to be 2 chars", len(colorCode), colorCode)
	}
	if !strings.Contains(colorCode, "\033[1;") {
		t.Error("colorcode needs escape chars", colorCode)
	}
}

func TestPrintColored(t *testing.T) {

	colored := systools.PrintColored("40", "something else")
	if colored != "\033[1;40msomething else" {
		t.Error("unexpected output format ", colored)
	}
}

func TestPrintColoredChanges(t *testing.T) {
	for i := 0; i < 40; i++ {
		var last string
		colorCode := systools.CreateColorCode()

		if i > 1 {
			if last == colorCode {
				t.Error("colorcode is the same again", colorCode, "prevoius", last)
			}
		}
		last = colorCode
	}
	
}
