package tasks_test

import (
	"strings"
	"testing"

	"github.com/swaros/contxt/module/tasks"
)

func TestAnkoVerify(t *testing.T) {
	script := `println('Hello World')
println('Have a nice day')`
	av := tasks.NewAnkVerifier()
	verified, err := av.VerifyLines(strings.Split(script, "\n"))
	if err != nil {
		t.Error(err)
	}
	if len(verified) != 2 {
		t.Error("expected 2 but got", len(verified))
	} else {
		for _, v := range verified {
			if v.Err != nil {
				t.Error(v.Err)
				t.Log(v.Line)
			}
		}
	}
}

func TestAnkoVerifyWithError(t *testing.T) {
	script := `println('Hello World')
come to the lalaland
println('Have a nice day')`
	av := tasks.NewAnkVerifier()
	verified, err := av.VerifyLines(strings.Split(script, "\n"))
	if err != nil {
		t.Error(err)
	}
	lineCount := len(strings.Split(script, "\n"))
	if len(verified) != lineCount {
		t.Error("expected ", lineCount, " but got ", len(verified))
	} else {
		for i, v := range verified {
			if v.Err != nil && i != 1 {
				t.Error(v.Err)
				t.Log(v.Line)
			}
		}
	}
}
