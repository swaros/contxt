package ctxout_test

import (
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

func TestInterfaces(t *testing.T) {

	tst := &TestAsInterface{}
	message := ctxout.ToString(tst, "hello", "test")
	if message != "hello TEST" {
		t.Error("message should be 'hello TEST'")
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
