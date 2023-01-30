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

type TPostFilter struct {
	config ctxout.PostFilterInfo
}

func (tst *TPostFilter) CanHandleThis(text string) bool {

	return strings.Contains(text, "<UPPER>") && strings.Contains(text, "</UPPER>")
}

// simple example of a post filter
func (tst *TPostFilter) Command(cmd string) string {
	leftparts := strings.Split(cmd, "<UPPER>")
	rightparts := strings.Split(leftparts[1], "</UPPER>")
	if tst.config.Disabled {
		return leftparts[0] + rightparts[0] + rightparts[1]
	}
	return leftparts[0] + strings.ToUpper(rightparts[0]) + rightparts[1]
}

func (tst *TPostFilter) Update(info ctxout.PostFilterInfo) {
	tst.config = info
}

func TestPostFilter(t *testing.T) {
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
