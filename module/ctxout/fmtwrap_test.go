package ctxout_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestFmtConcurrentWithInject(t *testing.T) {
	prnt := ctxout.NewFmtWrap()
	assertConcurrentPrinter(t, prnt, 100, func(jobIndex int) []interface{} {
		for i := 0; i < 100; i++ {
			ctxout.Print(prnt, ".......")
			ctxout.PrintLn(prnt, "inline print ", jobIndex, " ", i)
		}
		return []interface{}{"testtask ", jobIndex}
	})

}

func TestFmtConcurrent(t *testing.T) {
	assertConcurrentPrinter(t, ctxout.NewFmtWrap(), 100, func(jobIndex int) []interface{} {
		for i := 0; i < 100; i++ {
			ctxout.Print(".......")
			ctxout.PrintLn("inline print ", jobIndex, " ", i)
		}
		return []interface{}{"testtask ", jobIndex}
	})

}
