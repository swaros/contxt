package awaitgroup_test

import (
	"testing"
	"time"

	"github.com/swaros/contxt/module/awaitgroup"
)

func TestFlow(t *testing.T) {
	testFlow := awaitgroup.NewFlow()
	testFlow.Func(func(args ...interface{}) []interface{} {
		// just return the arguments unchanged
		// right now we just now if we get the right order

		// get argument 0, cast them them to int and multiply them
		// with 100 and wait this as milliseconds
		wait := args[0].(int) * 100
		time.Sleep(time.Duration(wait) * time.Millisecond)
		return args
	})
	numbers := []int{}
	stringsSlice := []string{}
	floats := []float64{}
	testFlow.Each(3, "hello", 3.14).
		Each(4, "world", 6.28).
		Each(1, "!", 9.42).
		Handler(func(args ...interface{}) {
			for _, arg := range args {
				switch m := arg.(type) {
				case int:
					numbers = append(numbers, m)
				case string:
					stringsSlice = append(stringsSlice, m)
				case float64:
					floats = append(floats, m)
				}
			}
		}).
		Run()

	expectNumbers := []int{3, 4, 1}
	expectStrings := []string{"hello", "world", "!"}
	expectFloats := []float64{3.14, 6.28, 9.42}

	if len(numbers) != len(expectNumbers) {
		t.Errorf("numbers length is not equal to expected numbers length")
	} else {
		for i := range numbers {
			if numbers[i] != expectNumbers[i] {
				t.Errorf("numbers[%d] is not equal to expected numbers[%d]", i, i)
			}
		}
	}

	if len(stringsSlice) != len(expectStrings) {
		t.Errorf("strings length is not equal to expected strings length")
	} else {
		for i := range stringsSlice {
			if stringsSlice[i] != expectStrings[i] {
				t.Errorf("strings[%d] is not equal to expected strings[%d]", i, i)
			}
		}
	}

	if len(floats) != len(expectFloats) {
		t.Errorf("floats length is not equal to expected floats length")
	} else {
		for i := range floats {
			if floats[i] != expectFloats[i] {
				t.Errorf("floats[%d] is not equal to expected floats[%d]", i, i)
			}
		}
	}

}
