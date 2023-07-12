package systools_test

import (
	"testing"

	"github.com/swaros/contxt/module/systools"
)

func TestContains(t *testing.T) {
	slice := []string{"hello", "world"}

	if systools.SliceContains(slice, "yolo") {
		t.Error("yolo is not on the slice")
	}

	if !systools.SliceContains(slice, "world") {
		t.Error("world should be found")
	}
}

func TestStringInSlice(t *testing.T) {
	slice := []string{"hello", "world"}

	if systools.StringInSlice("yolo", slice) {
		t.Error("yolo is not on the slice")
	}

	if !systools.StringInSlice("world", slice) {
		t.Error("world should be found")
	}
}

func TestMap(t *testing.T) {
	testDataMap := map[string]any{
		"hello": "world",
		"yolo":  "swag",
		"foo":   "bar",
	}

	outTest := []string{}
	systools.MapRangeSortedFn(testDataMap, func(key string, value any) {
		outTest = append(outTest, key+":"+value.(string))
	})

	expectedSlice := []string{"foo:bar", "hello:world", "yolo:swag"}

	if len(outTest) != len(expectedSlice) {
		t.Errorf("expected %d elements, got %d", len(expectedSlice), len(outTest))
	}

	for i, v := range outTest {
		if v != expectedSlice[i] {
			t.Errorf("expected %s, got %s", expectedSlice[i], v)
		}
	}

}
