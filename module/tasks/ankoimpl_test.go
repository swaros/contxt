package tasks_test

import (
	"testing"

	"github.com/swaros/contxt/module/tasks"
)

func TestRunAnko(t *testing.T) {
	ar := tasks.NewAnkoRunner()
	res, rErr := ar.RunAnko(`println("Hello World :)")`)
	if rErr != nil {
		t.Error(rErr)
	}

	if res != nil {
		t.Log(res)
	}
	buff := ar.GetBuffer()
	if len(buff) != 1 {
		t.Error("expected 1 but got", len(buff))
	}
	expected := "Hello World :)"
	if buff[0] != expected {
		t.Errorf("expected %s but got %s", expected, buff[0])
	}
}

func TestRunAnkoWithDefine(t *testing.T) {
	var verifyValue int64 = 0
	var expectedValue int64 = 60
	ar := tasks.NewAnkoRunner()
	ar.Define("sum", func(a ...interface{}) (int64, error) {
		var res int64 = 0
		for _, v := range a {
			res += v.(int64)
		}
		verifyValue = res
		return res, nil
	})

	res, rErr := ar.RunAnko(`sum(10, 20, 30)`)
	if rErr != nil {
		t.Error(rErr)
	}

	if verifyValue != expectedValue {
		t.Errorf("expected %d but got %d", expectedValue, verifyValue)
	}

	if res != nil {
		t.Log(res)
	}
}

func TestRunAnkoWithDefineAndDefault(t *testing.T) {
	var verifyValue int64 = 0
	var expectedValue int64 = 60
	ar := tasks.NewAnkoRunner()
	ar.Define("sum", func(a ...interface{}) (int64, error) {
		var res int64 = 0
		for _, v := range a {
			res += v.(int64)
		}
		verifyValue = res
		return res, nil
	})

	res, rErr := ar.RunAnko(`sum(10, 20, 30)`)
	if rErr != nil {
		t.Error(rErr)
	}

	derr := ar.AddDefaultDefine("defaultSum", func(a ...interface{}) (int64, error) {
		var res int64 = 0
		return res, nil
	})

	if derr == nil {
		t.Error("expected error but got nil. default define should not be added after initialization (lazyInit on RunAnko)")
	}

	if verifyValue != expectedValue {
		t.Errorf("expected %d but got %d", expectedValue, verifyValue)
	}

	if res != nil {
		t.Log(res)
	}
}
