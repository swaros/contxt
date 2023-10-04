package linehack_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaros/contxt/module/linehack"
)

func helperApplyTokens(parser *linehack.Parser) {
	var checks []linehack.TokenSelfProvider
	var neighbors []linehack.TokenSelfProvider
	// add all token types to checks
	checks = append(checks,
		&linehack.If{},
		&linehack.TBracketOpen{},
		&linehack.TBracketClose{},
		&linehack.TCurlyOpen{},
		&linehack.TCurlyClose{},
		&linehack.TSemiColon{},
		&linehack.Then{},
		&linehack.Else{},
		&linehack.TPrint{},
		&linehack.TSet{},
		&linehack.TVariable{},
		&linehack.TString{},
		&linehack.TDivide{},
		&linehack.TMultiply{},
		&linehack.TPlus{},
		&linehack.TMinus{},
		&linehack.TEqual{},
		&linehack.TNotEqual{},
		&linehack.TLess{},
		&linehack.TLessOrEqual{},
		&linehack.TGreater{},
		&linehack.TGreaterOrEqual{},
		&linehack.TAnd{},
		&linehack.TOr{},
		&linehack.TModulo{},
		&linehack.TInt{},
		&linehack.TFloat{},
		&linehack.TBool{},
		&linehack.TVar{},
		&linehack.TNot{},
		&linehack.TAssign{},
		&linehack.TAssignPlus{},
		&linehack.TOrPrecedence{},
		&linehack.TXor{},
		&linehack.TPrefixedVariable{},
	)

	parser.SetUseTokens(checks)

	// add all token types to neighbors
	neighbors = append(neighbors, &linehack.TGreaterOrEqual{},
		&linehack.TLessOrEqual{},
		&linehack.TNotEqual{},
		&linehack.TAssignPlus{},
		&linehack.TEqual{},
		&linehack.TOr{},
	)

	parser.SetNeighborTokens(neighbors)
}

func TestParse(t *testing.T) {
	parser := linehack.NewParser()
	helperApplyTokens(parser)
	parser.PrintByFmt = true
	parser.Parse(`if ($testvar == "test" || $testvar < "check") then {set output = "test"; print "hello"} else { print "world"}`)
}

func TestExecute(t *testing.T) {
	parser := linehack.NewParser()
	helperApplyTokens(parser)
	parser.SetVariableRequester(func(name string) (interface{}, error) {
		return "test", nil
	})

	parser.PrintByFmt = true
	err, _ := parser.Execute(`if ($testvar == "test" || $testvar < "check") then {set output = "test"; print "hello"} else { print "world"}`)
	if err != nil {
		t.Error(err)
	}
}

// tesing out put
type TestPrint struct {
	Pos      int
	RunCount int
	Output   string
}

func (tp *TestPrint) ItsMe(tok linehack.ScanToken) bool {
	return tok.Value == "TestPrint"
}
func (tp *TestPrint) SetValue(tok linehack.ScanToken) {
	tp.Pos = tok.Pos
}

func (tp *TestPrint) Copy() linehack.TokenSelfProvider {
	return tp // we do not copy oure self. we like being a pointer
}

func (tp *TestPrint) Run(args ...interface{}) (interface{}, error) {
	tp.RunCount++
	if len(args) > 0 {
		for _, arg := range args {
			switch tkn := arg.(type) {
			case string:
				tp.Output += tkn
			}
		}
	}
	return true, nil
}

func (tp *TestPrint) GetPos() int {
	return tp.Pos
}

func TestBasicCondition(t *testing.T) {
	parser := linehack.NewParser()
	helperApplyTokens(parser)
	var testVar string
	testPrinter := &TestPrint{}
	parser.AddUseTokenFirst(testPrinter)
	parser.SetVariableRequester(func(name string) (interface{}, error) {
		if name == "testVar" {
			testVar = "HIT"

			return "test", nil
		}
		return "", nil
	})

	parser.PrintByFmt = true
	//err := parser.Execute(`if "test" == "test" THEN { print "hello" }`)
	_, err := parser.Execute(`if "test" == $testVar then TestPrint("hello")`)
	if err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "HIT", testVar)
		assert.Equal(t, "hello", testPrinter.Output)
	}

	_, err = parser.Execute(`if "test" == $testVar then TestPrint("_again" "_next")`)
	if err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "hello_again_next", testPrinter.Output)
	}

	_, err = parser.Execute(`if "nomatch" == $testVar then TestPrint("NOOO")`)
	if err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "hello_again_next", testPrinter.Output)
	}

	_, err = parser.Execute(`if "nomatch" == $testVar || "test" == $testVar then TestPrint("_chained")`)
	if err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, "hello_again_next_chained", testPrinter.Output)
	}
}

type MultiRunTest struct {
	Pos      int
	RunCount int
	Result   string
}

func (mr *MultiRunTest) ItsMe(tok linehack.ScanToken) bool {
	return tok.Value == "Check"
}
func (mr *MultiRunTest) SetValue(tok linehack.ScanToken) {
	mr.Pos = tok.Pos
}

func (mr *MultiRunTest) Copy() linehack.TokenSelfProvider {
	return mr // we do not copy oure self. we like being a pointer
}

func (mr *MultiRunTest) Run(args ...interface{}) (interface{}, error) {
	mr.RunCount++
	if len(args) > 0 {
		for _, arg := range args {
			switch tkn := arg.(type) {
			case string:
				mr.Result += tkn
			}
		}
	}
	return true, nil
}

func (mr *MultiRunTest) GetPos() int {
	return mr.Pos
}

func TestMultipleCondition(t *testing.T) {

	type testConditions struct {
		testCondition string
		expected      string
		varName       string
		varValue      string
	}

	var mutliRunTest []testConditions
	mutliRunTest = append(mutliRunTest,
		testConditions{testCondition: `if "test" == "test"`, expected: "hello"},
		testConditions{testCondition: `if "test" != "XXX"`, expected: "round2"},
		testConditions{testCondition: `if $testVar != "test"`, expected: "dada", varName: "testVar", varValue: "dada"},
	)

	parser := linehack.NewParser()
	helperApplyTokens(parser)
	var varName = "testVar"
	var varValue = "test"
	parser.SetVariableRequester(func(name string) (interface{}, error) {
		if name == varName {
			return varValue, nil
		}
		return "", nil
	})

	parser.PrintByFmt = true
	cmd := &MultiRunTest{}
	parser.AddUseTokenFirst(cmd)

	for loopNr, test := range mutliRunTest {
		cmd.Result = ""
		if test.varName != "" {
			varName = test.varName
		}

		if test.varValue != "" {
			varValue = test.varValue
		}

		_, err := parser.Execute(test.testCondition + ` then Check("` + test.expected + `")`)
		if err != nil {
			t.Error(err)
		} else {
			assert.Equal(t, test.expected, cmd.Result, "failed to verify test case: "+test.testCondition+" @loop:"+strconv.Itoa(loopNr+1))
		}
	}

}

func TestVarUnknow(t *testing.T) {
	parser := linehack.NewParser()
	helperApplyTokens(parser)
	parser.PrintByFmt = true
	_, err := parser.Execute(` # $testvar = "test"`)
	if err == nil {
		t.Error("expected error")
	}
}

// This is a test token that will be used to test the parser
type CheckTest struct {
	Pos       int
	RunCount  int
	Argurment string
}

func (ct *CheckTest) ItsMe(tok linehack.ScanToken) bool {
	return tok.Value == "Check"
}
func (ct *CheckTest) SetValue(tok linehack.ScanToken) {
	ct.Pos = tok.Pos
}

func (ct *CheckTest) Copy() linehack.TokenSelfProvider {
	return ct // we do not copy oure self. we like being a pointer
}

func (ct *CheckTest) Run(args ...interface{}) (interface{}, error) {
	ct.RunCount++
	if len(args) > 0 {
		for _, arg := range args {
			switch tkn := arg.(type) {
			case string:
				ct.Argurment = tkn
			}
		}
	}
	return true, nil
}

func (ct *CheckTest) GetPos() int {
	return ct.Pos
}

// Testing a simple command callback
func TestVarSet(t *testing.T) {
	parser := linehack.NewParser()
	helperApplyTokens(parser)
	parser.PrintByFmt = true

	chck := &CheckTest{}
	parser.AddUseTokenFirst(chck)

	result, err := parser.Execute(`Check()`)
	if err != nil {
		t.Error(err)
	}

	assert.True(t, result.(bool))
	assert.Greater(t, chck.RunCount, 0)

	result, err = parser.Execute(`Check("hello")`)
	if err != nil {
		t.Error(err)
	}

	assert.Greater(t, chck.RunCount, 1)
	assert.Equal(t, "hello", chck.Argurment)
	assert.True(t, result.(bool))
}
