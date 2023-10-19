package ctxout_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/swaros/contxt/module/ctxout"
)

var (
	testMu sync.Mutex
)

func assertConcurrentPrinter(t *testing.T, printer ctxout.StreamInterface, count int, f func(jobIndex int) []interface{}) []string {
	t.Helper()
	addedStrings := []string{}
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(jobIndex int) {
			defer wg.Done()
			msg := f(jobIndex)
			var reformMap []interface{}
			reformMap = append(reformMap, printer)
			reformMap = append(reformMap, msg...)
			addedStrings = append(addedStrings, messageToStrSlice(msg...))
			printer.StreamLn(reformMap...)
			// wait 10 ms to allow other goroutines to print
			time.Sleep(2 * time.Millisecond)

		}(i)
	}
	wg.Wait()
	return addedStrings
}

func assertTableOutput(t *testing.T, result, expected []string) {
	t.Helper()
	lenExpected := len(expected)
	if len(result) < lenExpected {
		t.Errorf("expected %d messages, got %d", lenExpected, len(result))
	}
	for i, exp := range expected {
		if i >= lenExpected {
			t.Error("expected more messages than we got")
			break
		}
		if result[i] != exp {
			t.Errorf("expected \n[%s]\ngot\n[%s]", exp, result[i])
		}
	}
}

func messageToStrSlice(msg ...interface{}) string {
	testMu.Lock()
	defer testMu.Unlock()
	strmap := []string{}
	for _, m := range msg {
		switch m.(type) {
		case string, int, int64, float64, bool, uint, uint64, float32, uint32, int32, int16, uint16, int8, uint8, uintptr, complex64, complex128, error:
			strmap = append(strmap, fmt.Sprintf("%v", m))
		}

	}
	addStr := strings.Join(strmap, "")
	return addStr
}

type TestStreamA struct {
	messages sync.Map
}

func (tst *TestStreamA) Stream(msg ...interface{}) {
	// just an uuid to make sure we don't have any duplicates
	id := uuid.New().String()
	key := fmt.Sprintf("%v-noln-%s", time.Now().UnixNano(), id)
	tst.messages.Store(key, messageToStrSlice(msg...))
}

func (tst *TestStreamA) StreamLn(msg ...interface{}) {
	id := uuid.New().String()
	key := fmt.Sprintf("%v-haveln-%s", time.Now().UnixNano(), id)
	tst.messages.Store(key, messageToStrSlice(msg...))
}

func (tst *TestStreamA) GetMessages() []string {
	res := []string{}
	tst.messages.Range(func(key, value interface{}) bool {
		res = append(res, value.(string))
		return true
	})
	return res
}

func (tst *TestStreamA) Reset() {
	tst.messages = sync.Map{}
}

func (tst *TestStreamA) GetSize() int {
	return len(tst.GetMessages())
}

func NewTestStreamA(expectedSize int) *TestStreamA {
	return &TestStreamA{
		messages: sync.Map{},
	}
}

func assertTestStreamAContents(t *testing.T, tst *TestStreamA, expected []string) {
	t.Helper()
	if tst.GetSize() < len(expected) {
		t.Errorf("expected at least %d messages, got %d", len(expected), tst.GetSize())
	}
	// check if all expected messages are in the stream
	for _, exp := range expected {
		found := false
		for _, msg := range tst.GetMessages() {
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
	added := assertConcurrentPrinter(t, prnt, counter, func(jobIndex int) []interface{} {
		for i := 0; i < counter; i++ {
			ctxout.Print(prnt, "nonln print ", jobIndex, " ", i)
			ctxout.PrintLn(prnt, "inline print ", jobIndex, " ", i)
		}
		return []interface{}{"testtask ", jobIndex}
	})
	assertTestStreamAContents(t, prnt, added)

}

type testCase struct {
	label   string
	content string
	info    string
}

func helperDrawTable(t *testing.T, prnt ctxout.PrintInterface, testCases []testCase) []string {
	t.Helper()
	result := []string{}
	for _, tc := range testCases {

		leftLabel := ctxout.ToString(ctxout.NewMOWrap(), ctxout.ForeYellow, "<sign debug> ")
		rightLabel := ctxout.ToString(ctxout.NewMOWrap(), ctxout.ForeYellow, "<sign error> ")
		msg := ctxout.ToString(
			prnt,
			ctxout.Row(

				ctxout.TD(
					tc.label,
					ctxout.Prop(ctxout.AttrSize, 10),
					ctxout.Prop(ctxout.AttrOrigin, 2),
					ctxout.Prop(ctxout.AttrPrefix, leftLabel),
					ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
				),

				ctxout.TD(
					tc.content,
					ctxout.Prop(ctxout.AttrSize, 85),
					ctxout.Prop(ctxout.AttrPrefix, rightLabel),
					ctxout.Prop(ctxout.AttrOverflow, "wordwrap"),
					ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
				),
				ctxout.TD(
					tc.info,
					ctxout.Prop(ctxout.AttrSize, 5),
					ctxout.Prop(ctxout.AttrOrigin, 2),
					ctxout.Prop(ctxout.AttrPrefix, ctxout.ForeDarkGrey),
					ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
				),
			),
		)
		result = append(result, msg)

	}
	return result
}

func TestFilterMixings(t *testing.T) {
	prnt := ctxout.NewMOWrap()

	taboutFilter := ctxout.NewTabOut()
	signsFilter := ctxout.NewSignFilter(ctxout.NewBaseSignSet())
	ctxout.AddPostFilter(signsFilter)
	ctxout.AddPostFilter(taboutFilter)

	info := ctxout.PostFilterInfo{
		Width:      100,   // give us a big width so we can render the whole line
		IsTerminal: true,  //no terminal
		Colored:    false, // no colors
		Height:     500,   // give us a big height so we can render the whole line
		Disabled:   false,
	}
	taboutFilter.Update(info)
	signsFilter.Update(info)

	testCases := []testCase{
		{
			label:   "label1",
			content: "content1",
			info:    "info1",
		},
		{
			label:   "label2",
			content: "content2",
			info:    "info2",
		},
	}
	result := helperDrawTable(t, prnt, testCases)
	expected := []string{
		"[•…] ...l1[×] content1                                                                         info1",
		"[•…] ...l2[×] content2                                                                         info2",
	}
	assertTableOutput(t, result, expected)

	signsFilter.ForceEmpty(true)

	expected = []string{
		"    label1 content1                                                                            info1",
		"    label2 content2                                                                            info2",
	}
	result = helperDrawTable(t, prnt, testCases)
	assertTableOutput(t, result, expected)

}
