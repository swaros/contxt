package ctxout_test

import (
	"strings"
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestPrinter(t *testing.T) {
	if ctxout.PreHook != nil {
		t.Error("PreHook should be nil")
	}
	ctxout.PreHook = func(msg ...interface{}) bool {
		return true
	}
	if ctxout.PreHook == nil {
		t.Error("PreHook should not be nil")
	}
	ctxout.Print("Hello")
	ctxout.PreHook = nil
	if ctxout.PreHook != nil {
		t.Error("PreHook should be nil")
	}
	ctxout.Print("Hello")
}

// setup a test interface
type TestAsInterface struct{}

// we need to implement the Filter method
// this method will be called by the ctxout package
func (tst *TestAsInterface) Filter(msg interface{}) interface{} {
	if msg == "test" {
		return "TEST"
	}
	return msg
}

func (tst *TestAsInterface) Update(info ctxout.CtxOutBehavior) {
}

func (tst *TestAsInterface) CanHandleThis(text string) bool {
	return true
}

func TestInterfaces(t *testing.T) {

	tst := &TestAsInterface{}
	if !ctxout.IsPrinterInterface(tst) {
		t.Error("TestAsInterface should be a printer interface")
	}

	message := ctxout.ToString(tst, "hello", "test")
	if message != "helloTEST" {
		t.Error("message should be 'helloTEST'")
	}

}

type TestStream struct {
	message string
}

func (tst *TestStream) Stream(msg ...interface{}) {
	tst.message += msg[0].(string)
}

func (tst *TestStream) StreamLn(msg ...interface{}) {
	tst.message += msg[0].(string) + ";"
}

func TestStreamFn(t *testing.T) {
	tst := &TestStream{}
	ctxout.PrintLn(tst, "hello")
	if tst.message != "hello;" {
		t.Error("message should be 'hello;'")
	}
	ctxout.Print(tst, "hello")
	if tst.message != "hello;hello" {
		t.Error("message should be 'hello;hello'")
	}
}

// create a test post filter
// that looks for <UPPER> and </UPPER> and converts the text between to upper case
type TPostFilter struct {
	config ctxout.PostFilterInfo
}

// looking for <UPPER> and </UPPER>
func (tst *TPostFilter) CanHandleThis(text string) bool {
	return strings.Contains(text, "<UPPER>") && strings.Contains(text, "</UPPER>")
}

// simple example of a post filter that handles
// <UPPER>test</UPPER> and converts the content to upper case
func (tst *TPostFilter) Command(cmd string) string {
	leftparts := strings.Split(cmd, "<UPPER>")
	rightparts := strings.Split(leftparts[1], "</UPPER>")
	if tst.config.Disabled {
		return leftparts[0] + rightparts[0] + rightparts[1]
	}
	return leftparts[0] + strings.ToUpper(rightparts[0]) + rightparts[1]
}

// update the config
func (tst *TPostFilter) Update(info ctxout.PostFilterInfo) {
	tst.config = info
}

func TestPostFilter(t *testing.T) {
	ctxout.ClearPostFilters() // clear all post filters
	filter := &TPostFilter{}
	ctxout.AddPostFilter(filter)
	message := ctxout.ToString("hello <UPPER>test</UPPER> world")
	if message != "hello TEST world" {
		t.Error("message should be 'hello TEST world' got '" + message + "'")
	}
	filter.config.Disabled = true
	message = ctxout.ToString("hello <UPPER>test</UPPER> world")
	if message != "hello test world" {
		t.Error("message should be 'hello test world' got '" + message + "'")
	}
}

// we create a second Test Filter they reverts the text
type TPostSecondFilter struct {
	config ctxout.PostFilterInfo
}

// looking for <REVERT> and </REVERT>
func (tst *TPostSecondFilter) CanHandleThis(text string) bool {
	return strings.Contains(text, "<REVERT>") && strings.Contains(text, "</REVERT>")
}

// simple example of a post filter that handles
// <REVERT>test</REVERT> and converts the content in revert order
// so <REVERT>test</REVERT> will be tset
func (tst *TPostSecondFilter) Command(cmd string) string {
	leftparts := strings.Split(cmd, "<REVERT>")
	rightparts := strings.Split(leftparts[1], "</REVERT>")
	if tst.config.Disabled {
		return leftparts[0] + rightparts[0] + rightparts[1]
	}
	return leftparts[0] + revert(rightparts[0]) + rightparts[1]
}

// update the config
func (tst *TPostSecondFilter) Update(info ctxout.PostFilterInfo) {
	tst.config = info
}

// revert the text. used by the TPostFilter2
func revert(text string) string {
	result := ""
	for i := len(text) - 1; i >= 0; i-- {
		result += string(text[i])
	}
	return result
}

func TestMutliplePostFilter(t *testing.T) {
	ctxout.ClearPostFilters() // clear all post filters
	filter := &TPostFilter{}
	ctxout.AddPostFilter(filter)

	testMessage := "<UPPER>now <REVERT>lliw ew</REVERT> see</UPPER> if this works"
	message := ctxout.ToString(testMessage)
	// first just the upper filter
	expected := "NOW <REVERT>LLIW EW</REVERT> SEE if this works"
	if message != expected {
		t.Error("message should be '" + expected + "' got '" + message + "'")
	}
	// now we add the second filter
	filter2 := &TPostSecondFilter{}
	ctxout.AddPostFilter(filter2)
	message = ctxout.ToString(testMessage)
	// now we have both filters
	expected = "NOW WE WILL SEE if this works"
	if message != expected {
		t.Error("message should be '" + expected + "' got '" + message + "'")
	}

	// here we go with the TabOut Filter
	ctxout.AddPostFilter(ctxout.NewTabOut())
	testMessage = "<table><row><tab><UPPER>now <REVERT>lliw ew</REVERT></tab><tab> see</UPPER> if this works</row></table>"
	message = ctxout.ToString(testMessage)
	// now we have 3 filters
	expected = "NOW WE WILL SEE if this works"
	if message != expected {
		t.Error("message should be '" + expected + "' got '" + message + "'")
	}

	// same with more complex stuff
	testMessage = "<table><row size='40' fill='.' draw='fixed' origin='2'><tab><UPPER>now <REVERT>lliw ew</REVERT></tab><tab size='60'> see</UPPER> if this works</row></table>"
	expected = "NOW WE WILL SEE if this works"
	message = ctxout.ToString(testMessage)
	if message != expected {
		t.Error("message should be '" + expected + "' got '" + message + "'")
	}
	// recheck reported filters executes. even if the message expectations are fulfilled.
	runInfos := ctxout.GetRunInfosF("(YES) post filter")
	if len(runInfos) != 3 {
		t.Error("should be 3 run infos. got ", len(runInfos), runInfos)
	}

	// check any filter is executed
	runInfos = ctxout.GetRunInfosF("(YES) post filter: *ctxout_test.TPostFilter")
	if len(runInfos) != 1 {
		t.Error("should be 1 run infos. got ", len(runInfos), runInfos)
	}

	runInfos = ctxout.GetRunInfosF("(YES) post filter: *ctxout_test.TPostSecondFilter")
	if len(runInfos) != 1 {
		t.Error("should be 1 run infos. got ", len(runInfos), runInfos)
	}

	runInfos = ctxout.GetRunInfosF("(YES) post filter: *ctxout.TabOut")
	if len(runInfos) != 1 {
		t.Error("should be 1 run infos. got ", len(runInfos), runInfos)
	}
}
