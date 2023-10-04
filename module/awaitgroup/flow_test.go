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
	err := testFlow.Go(3, "hello", 3.14).
		Go(4, "world", 6.28).
		Go(1, "!", 9.42).
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

	if err != nil {
		t.Errorf("error running flow: %s", err)
	}

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

func theTestMethod(name string, age int, lastName string) (string, int) {
	return name + ", " + lastName, age
}

func TestFlowWithReflect(t *testing.T) {
	TestFlow := awaitgroup.NewFlow()
	if err := TestFlow.Use(theTestMethod); err != nil {
		t.Errorf("error using theTestMethod: %s", err)
	}

	keepResultNames := []string{}
	keepResultAges := []int{}

	TestFlow.Go("Michael", 42, "Miller").
		Go("Jane", 24, "Doe").
		Handler(func(args ...interface{}) {
			for _, arg := range args {
				switch m := arg.(type) {
				case string:
					keepResultNames = append(keepResultNames, m)
					t.Logf("string: %s", m)
				case int:
					t.Logf("int: %d", m)
					keepResultAges = append(keepResultAges, m)
				}
			}
		}).Run()

	expectNames := []string{"Michael, Miller", "Jane, Doe"}
	expectAges := []int{42, 24}

	if len(keepResultNames) != len(expectNames) {
		t.Errorf("names length is not equal to expected names length")
	} else {
		for i := range keepResultNames {
			if keepResultNames[i] != expectNames[i] {
				t.Errorf("names[%d] is not equal to expected names[%d]", i, i)
			}
		}
	}

	if len(keepResultAges) != len(expectAges) {
		t.Errorf("ages length is not equal to expected ages length")
	} else {
		for i := range keepResultAges {
			if keepResultAges[i] != expectAges[i] {
				t.Errorf("ages[%d] is not equal to expected ages[%d]", i, i)
			}
		}
	}

}

func TestFlowWithReflectAndError(t *testing.T) {
	TestFlow := awaitgroup.NewFlow()
	if err := TestFlow.Use(theTestMethod); err != nil {
		t.Errorf("error using theTestMethod: %s", err)
	}

	keepResultNames := []string{}
	keepResultAges := []int{}

	runError := TestFlow.
		Go(88, 42, "Miller").
		Go("Jane", 24, "Doe").
		Handler(func(args ...interface{}) {
			for _, arg := range args {
				switch m := arg.(type) {
				case string:
					keepResultNames = append(keepResultNames, m)
					t.Logf("string: %s", m)
				case int:
					t.Logf("int: %d", m)
					keepResultAges = append(keepResultAges, m)
				}
			}
		}).Run()

	if runError == nil {
		t.Errorf("runError should not be nil")
	} else {
		expectedError := "argument type does not match"
		if runError.Error() != expectedError {
			t.Errorf("runError.Error() is not equal to expectedError. Expected: %s, got: %s", expectedError, runError.Error())
		}
	}

}

func TestFlowWithReflectAndError2(t *testing.T) {
	TestFlow := awaitgroup.NewFlow()
	if err := TestFlow.Use(nil); err == nil {
		t.Error("error should not be nil")
	} else {
		expectedError := "argument must be a function. got nil"
		if err.Error() != expectedError {
			t.Errorf("err.Error() is not equal to expectedError. Expected: %s, got: %s", expectedError, err.Error())
		}
	}

}

func TestFlowWithReflectAndError4(t *testing.T) {
	TestFlow := awaitgroup.NewFlow()
	if err := TestFlow.Use(theTestMethod); err != nil {
		t.Errorf("error using theTestMethod: %s", err)
	}

	keepResultNames := []string{}
	keepResultAges := []int{}

	runError := TestFlow.
		Go(88).
		Handler(func(args ...interface{}) {
			for _, arg := range args {
				switch m := arg.(type) {
				case string:
					keepResultNames = append(keepResultNames, m)
					t.Logf("string: %s", m)
				case int:
					t.Logf("int: %d", m)
					keepResultAges = append(keepResultAges, m)
				}
			}
		}).Run()

	if runError == nil {
		t.Errorf("runError should not be nil")
	} else {
		expectedError := "number of arguments does not match"
		if runError.Error() != expectedError {
			t.Errorf("runError.Error() is not equal to expectedError. Expected: %s, got: %s", expectedError, runError.Error())
		}
	}

}

func TestFlowWithReflectAndError5(t *testing.T) {
	TestFlow := awaitgroup.NewFlow()
	flag := false

	if err := TestFlow.Use(flag); err == nil {
		t.Error("error should not be nil")
	} else {
		expectedError := "argument must be a function"
		if err.Error() != expectedError {
			t.Errorf("err.Error() is not equal to expectedError. Expected: %s, got: %s", expectedError, err.Error())
		}
	}

	keepResultNames := []string{}
	keepResultAges := []int{}

	runError := TestFlow.
		Go(88).
		Handler(func(args ...interface{}) {
			for _, arg := range args {
				switch m := arg.(type) {
				case string:
					keepResultNames = append(keepResultNames, m)
					t.Logf("string: %s", m)
				case int:
					t.Logf("int: %d", m)
					keepResultAges = append(keepResultAges, m)
				}
			}
		}).Run()

	if runError == nil {
		t.Errorf("runError should not be nil")
	} else {
		expectedError := "function is not defined"
		if runError.Error() != expectedError {
			t.Errorf("runError.Error() is not equal to expectedError. Expected: %s, got: %s", expectedError, runError.Error())
		}
	}

}
