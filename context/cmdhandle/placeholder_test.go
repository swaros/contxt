package cmdhandle_test

import (
	"testing"

	"github.com/swaros/contxt/context/cmdhandle"
)

func TestBasicReplace(t *testing.T) {
	placeHolder := cmdhandle.NewPlaceHolderMap()
	placeHolder["test1"] = "here i am"
	placeHolder["test2"] = "XXX"

	testLine := "a: ${test1}"
	testLine2 := "b: ${test2} and again ${test2}"

	result := cmdhandle.HandlePlaceHolder(placeHolder, testLine)
	if result != "a: here i am" {
		t.Error("noting was replaced:'", testLine, "' => ", result)
	}

	result2 := cmdhandle.HandlePlaceHolder(placeHolder, testLine2)
	if result2 != "b: XXX and again XXX" {
		t.Error("noting was replaced:'", testLine2, "' => ", result2)
	}
}
