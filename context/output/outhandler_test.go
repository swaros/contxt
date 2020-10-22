package output_test

import (
	"fmt"
	"testing"

	"github.com/swaros/contxt/context/output"
)

func TestOuthandler(t *testing.T) {
	output.ColorEnabled = true
	outstr := output.Message("hello world", " you sucks")
	fmt.Println(outstr)
	if len(outstr) != 21 {
		t.Error("wrong result size", len(outstr))
	}

	if outstr != "hello world you sucks" {
		t.Error("unexpected result: ", outstr)
	}
}

func TestColorReplaced(t *testing.T) {
	output.ColorEnabled = true
	outstr := output.Message("<f:yellow>hello world in yellow</> this is back to regular")
	fmt.Println(outstr)
	if len(outstr) != 54 {
		t.Error("wrong result size", len(outstr))
	}
}

func TestColorReplacedDisabled(t *testing.T) {
	output.ColorEnabled = false
	outstr := output.Message("<f:yellow>hello world in yellow</>")
	fmt.Println(outstr)
	if len(outstr) != 21 {
		t.Error("wrong result size", len(outstr))
	}
}
