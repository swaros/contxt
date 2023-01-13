package linehack_test

import (
	"testing"

	"github.com/swaros/contxt/module/linehack"
)

func TestParse(t *testing.T) {
	parser := linehack.NewParser()
	parser.PrintByFmt = true
	parser.Parse(`if ($testvar == "test" || $testvar < "check") then {set output = "test"; print "hello"} else { print "world"}`)
}

func TestExecute(t *testing.T) {
	parser := linehack.NewParser()
	parser.PrintByFmt = true
	err := parser.Execute(`if ($testvar == "test" || $testvar < "check") then {set output = "test"; print "hello"} else { print "world"}`)
	if err != nil {
		t.Error(err)
	}
}

func TestVarUnknow(t *testing.T) {
	parser := linehack.NewParser()
	parser.PrintByFmt = true
	err := parser.Execute(` # $testvar = "test"`)
	if err == nil {
		t.Error("expected error")
	}
}

func TestVarSet(t *testing.T) {
	parser := linehack.NewParser()
	parser.PrintByFmt = true
	err := parser.Execute(`set $testvar = "test"`)
	if err != nil {
		t.Error(err)
	}
}
