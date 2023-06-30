package yaclint_test

import (
	"fmt"
	"testing"

	"github.com/swaros/contxt/module/yaclint"
	"github.com/swaros/contxt/module/yamc"
)

func TestLogger(t *testing.T) {

	dirtylog := yaclint.NewDirtyLogger().SetAddTime(false)
	dirtylog.Trace("test")
	dirtylog.Trace("test2")
	dirtylog.Trace("test3")
	dirtylog.Trace("another topic", "test4")
	dirtylog.Trace("another topic", 4569)
	dirtylog.Trace("another topic", false, "what ever")

	all := dirtylog.GetTrace()
	if len(all) != 7 { // 6 entries + 1 for the "CreateSimpleTracer: there was no traceFn set. so we create a simple one."
		t.Errorf("expected 6 entries, got %d [%v]", len(all), all)
	}

	topic := dirtylog.GetTrace("another topic")
	if len(topic) != 3 {
		t.Errorf("expected 3 entries, got %d [%v]", len(topic), topic)
	}

	outStr := dirtylog.Print()
	expectStr := `CreateSimpleTracer: there was no traceFn set. so we create a simple one.
test
test2
test3
another topictest4
another topic4569
another topicfalsewhat ever`
	if outStr != expectStr {
		t.Errorf("expected \n%s\n, got \n%s", expectStr, outStr)
	}

}

func TestLoggerSetTraceFn(t *testing.T) {

	var outStr string
	dirtylog := yaclint.NewDirtyLogger().SetAddTime(false)
	dirtylog.SetTraceFn(func(args ...interface{}) {
		for _, a := range args {
			outStr += fmt.Sprintf("[%v]", a)
		}
	})
	dirtylog.Trace("test")
	dirtylog.Trace("test2")
	dirtylog.Trace("test3")
	dirtylog.Trace("another topic", "test4")
	dirtylog.Trace("another topic", 4569)
	dirtylog.Trace("another topic", false, "what ever")

	expectStr := `[test][test2][test3][another topic][test4][another topic][4569][another topic][false][what ever]`
	if outStr != expectStr {
		t.Errorf("expected \n%s\n, got \n%s", expectStr, outStr)
	}

}

func TestLoggerWithMatchToken(t *testing.T) {
	dirtylog := yaclint.NewDirtyLogger().SetAddTime(false)

	matchToken := yaclint.NewMatchToken(
		yamc.StructDef{}, func(args ...interface{}) {
			dirtylog.Trace(args...)
		},
		&yaclint.LintMap{},
		"test: \"([a-z]+)\"",
		1,
		2,
		true,
	)

	dirtylog.Trace("test: \"test1\"", matchToken)
	expectedOutStr := `CreateSimpleTracer: there was no traceFn set. so we create a simple one.
NewMatchToken:parse: test: "([a-z]+)"
MatchToken:[+ ()]  [test] !No Tag found!
NewMatchToken:[+] test (test): [-1] val[([a-z]+)] indx[1] seq[2] (string)
test: "test1"`
	outStr := dirtylog.Print()
	if outStr != expectedOutStr {
		t.Errorf("expected \n%s\n, got \n%s", expectedOutStr, outStr)
	}
}

func TestSomeValues(t *testing.T) {

	dirtylog := yaclint.NewDirtyLogger().SetAddTime(false)
	dirtylog.Trace("test", 1, 2, 3, 4, 5, 6, 7, 8, 9, 0)
	dirtylog.Trace("test a string slice", []string{"a", "b", "c"})
	dirtylog.Trace("test a string slice", []string{"a", "b", "c"}, "and some more", 1, 2, 3)

	outStr := dirtylog.Print()
	expectStr := `CreateSimpleTracer: there was no traceFn set. so we create a simple one.
test1234567890
test a string slice['a','b','c',]
test a string slice['a','b','c',]and some more123`

	if outStr != expectStr {
		t.Errorf("expected \n%s\n, got \n%s", expectStr, outStr)
	}

}

func TestLoggerHitMax(t *testing.T) {

	dirtylog := yaclint.NewDirtyLogger()
	for i := 0; i < 1000; i++ {
		dirtylog.Trace("test", i)
	}

	outBuffer := dirtylog.GetTrace()
	if len(outBuffer) != 499 {
		t.Errorf("expected 499 entries, got %d", len(outBuffer))
	}
}

func TestLoggerWithCreateCtxoutTracer(t *testing.T) {

	dirtylog := yaclint.NewDirtyLogger().SetAddTime(false).CreateCtxoutTracer()
	dirtylog.Trace("test", 1, 2, 3, 4, 5, 6, 7, 8, 9, 0)
	dirtylog.Trace("test a string slice", []string{"a", "b", "c"})
	dirtylog.Trace("test a string slice", []string{"a", "b", "c"}, "and some more", 1, 2, 3, true, false)

	matchToken := yaclint.NewMatchToken(
		yamc.StructDef{}, func(args ...interface{}) {
			dirtylog.Trace(args...)
		},
		&yaclint.LintMap{},
		"test: \"([a-z]+)\"",
		1,
		2,
		true,
	)

	dirtylog.Trace("test: \"test1\"", matchToken)
	expectedOutStr := `test1234567890
test a string slice['a','b','c']
test a string slice['a','b','c']and some more123truefalse
NewMatchToken:parse: test: "([a-z]+)"
MatchToken:[+ ()] "([a-z]+)" [test] !No Tag found!
NewMatchToken:[+] test (test): [-1] val[([a-z]+)] indx[1] seq[2] (string)
test: "test1"`

	outStr := dirtylog.Print()
	if outStr != expectedOutStr {
		t.Errorf("expected \n%s\n, got \n%s", expectedOutStr, outStr)
	}

}

func TestLoggerHitMaxWithCtxout(t *testing.T) {

	dirtylog := yaclint.NewDirtyLogger().SetAddTime(true).CreateCtxoutTracer()
	for i := 0; i < 1000; i++ {
		dirtylog.Trace("test", i)
	}

	outBuffer := dirtylog.GetTrace()
	if len(outBuffer) != 499 {
		t.Errorf("expected 499 entries, got %d", len(outBuffer))
	}
}
