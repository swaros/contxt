package ctxout_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/swaros/contxt/module/ctxout"
)

var (
	testMu sync.Mutex
)

func assertConcurrentPrinter(t *testing.T, printer ctxout.StreamInterface, count int, f func(jobIndex int) []interface{}) {
	t.Helper()
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(jobIndex int) {
			defer wg.Done()
			msg := f(jobIndex)
			var reformMap []interface{}
			reformMap = append(reformMap, printer)
			reformMap = append(reformMap, msg...)
			printer.StreamLn(reformMap...)
			// wait 10 ms to allow other goroutines to print
			time.Sleep(2 * time.Millisecond)

		}(i)
	}
	wg.Wait()

}

func messageToStrSlice(result []string, msg ...interface{}) []string {
	testMu.Lock()
	defer testMu.Unlock()
	strmap := []string{}
	for _, m := range msg {
		strmap = append(strmap, fmt.Sprintf("%v", m))
	}
	return append(result, strings.Join(strmap, ""))
}

type TestStreamA struct {
	messages []string
}

func (tst *TestStreamA) Stream(msg ...interface{}) {
	tst.messages = messageToStrSlice(tst.messages, msg...)
}

func (tst *TestStreamA) StreamLn(msg ...interface{}) {
	tst.messages = messageToStrSlice(tst.messages, msg...)
}

func (tst *TestStreamA) GetMessages() []string {
	return tst.messages
}

func (tst *TestStreamA) Reset() {
	tst.messages = []string{}
}

func NewTestStreamA(expectedSize int) *TestStreamA {
	return &TestStreamA{
		messages: make([]string, 0, expectedSize),
	}
}

func assertTestStreamAContents(t *testing.T, tst *TestStreamA, expected []string) {
	t.Helper()
	if len(tst.messages) < len(expected) {
		t.Errorf("expected at least %d messages, got %d", len(expected), len(tst.messages))
	}
	// check if all expected messages are in the stream
	for _, exp := range expected {
		found := false
		for _, msg := range tst.messages {
			if msg == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected message '%s' not found", exp)
			return
		}
	}
}

func TestInjectedStreams(t *testing.T) {
	counter := 15
	prnt := NewTestStreamA(counter * counter * 3)
	assertConcurrentPrinter(t, prnt, counter, func(jobIndex int) []interface{} {
		for i := 0; i < counter; i++ {
			ctxout.Print(prnt, "nonln print ", jobIndex, " ", i)
			ctxout.PrintLn(prnt, "inline print ", jobIndex, " ", i)
		}
		return []interface{}{"testtask ", jobIndex}
	})
}
