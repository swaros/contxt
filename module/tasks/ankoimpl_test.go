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
	} else {
		expected := "Hello World :)"
		if buff[0] != expected {
			t.Errorf("expected %s but got %s", expected, buff[0])
		}
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

func TestBuffers_01(t *testing.T) {
	ar := tasks.NewAnkoRunner()
	hookMessage := ""
	ar.SetBufferHook(func(msg string) {
		hookMessage += msg
	})
	ar.RunAnko(`println("Hello World :)")`)
	buff := ar.GetBuffer()
	if len(buff) != 1 {
		t.Error("expected 1 but got", len(buff))
	} else {
		expected := "Hello World :)"
		if buff[0] != expected {
			t.Errorf("expected %s but got %s", expected, buff[0])
		}
		expectedHookMessage := "Hello World :)"
		if hookMessage != expectedHookMessage {
			t.Errorf("expected [%s] but got [%s]", expectedHookMessage, hookMessage)
		}
	}
}

func TestBuffers_02(t *testing.T) {
	ar := tasks.NewAnkoRunner()
	hookMessage := ""
	ar.SetBufferHook(func(msg string) {
		hookMessage += msg
	})
	ar.RunAnko(`print("Hello World :)")`)
	buff := ar.GetBuffer()
	if len(buff) != 1 {
		t.Error("expected 1 but got", len(buff))
	} else {
		expected := "Hello World :)"
		if buff[0] != expected {
			t.Errorf("expected %s but got %s", expected, buff[0])
		}
	}
	expectedHookMessage := "Hello World :)"
	if hookMessage != expectedHookMessage {
		t.Errorf("expected [%s] but got [%s]", expectedHookMessage, hookMessage)
	}
}

func TestBuffers_03(t *testing.T) {
	ar := tasks.NewAnkoRunner()
	hookMessage := ""
	ar.SetBufferHook(func(msg string) {
		hookMessage += msg
	})
	ar.RunAnko(`print("Hello ")
	print("World :)")`)
	buff := ar.GetBuffer()
	if len(buff) != 1 {
		t.Error("expected 1 but got", len(buff))
	} else {
		expected := "Hello World :)"
		if buff[0] != expected {
			t.Errorf("expected %s but got %s", expected, buff[0])
		}
	}
	expectedHookMessage := "Hello World :)"
	if hookMessage != expectedHookMessage {
		t.Errorf("expected [%s] but got [%s]", expectedHookMessage, hookMessage)
	}
}

func TestErrors(t *testing.T) {
	ar := tasks.NewAnkoRunner()
	hookMessage := ""
	ar.SetBufferHook(func(msg string) {
		hookMessage += msg
	})
	_, err := ar.RunAnko(`print("Hello ")
	puffpaff("wtf")`)

	if err == nil {
		t.Error("expected error but got nil")
	} else {
		expectedError := "undefined symbol 'puffpaff'"
		if err.Error() != expectedError {
			t.Errorf("expected [%s] but got [%s]", expectedError, err.Error())
		}

	}

	buff := ar.GetBuffer()
	if len(buff) != 1 {
		t.Error("expected 1 but got", len(buff))
	} else {
		expected := "Hello "
		if buff[0] != expected {
			t.Errorf("expected %s but got %s", expected, buff[0])
		}
	}
	expectedHookMessage := "Hello "
	if hookMessage != expectedHookMessage {
		t.Errorf("expected [%s] but got [%s]", expectedHookMessage, hookMessage)
	}
}
