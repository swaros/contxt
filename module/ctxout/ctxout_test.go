package ctxout_test

import (
	"sync"
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func assertConcurrentPrinter(t *testing.T, printer ctxout.StreamInterface, count int, f func(jobIndex int) []interface{}) {
	t.Helper()
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(jobIndex int) {
			defer wg.Done()
			msg := f(jobIndex)
			printer.StreamLn(msg...)

		}(i)
	}
	wg.Wait()

}
