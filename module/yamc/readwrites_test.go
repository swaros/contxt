package yamc

import (
	"sync"
	"testing"
)

func TestPlainReadWrite(t *testing.T) {

	ymc := New()
	ymc.Store("test", "test")
	if val, found := ymc.Get("test"); !found {
		t.Error("error by getting data from yamc")
	} else {
		if val != "test" {
			t.Error("error by getting data from yamc")
		}
	}

}

func TestAsyncReadWrite(t *testing.T) {

	ymc := New()
	ymc.Store("a", 0)
	ymc.Store("b", 0)
	runCount := 1000
	var wg sync.WaitGroup
	doInc := func(name string, n int) {
		for i := 0; i < n; i++ {
			ymc.Update(name, func(val interface{}) interface{} {
				return val.(int) + 1
			})
		}
		wg.Done()
	}

	wg.Add(3)
	go doInc("a", runCount)
	go doInc("a", runCount)
	go doInc("b", runCount)
	wg.Wait()

	if valA, foundA := ymc.Get("a"); !foundA {
		t.Error("error by getting data from yamc")
	} else {
		if valA != runCount*2 {
			t.Error("error by getting data from yamc (a). unexpected value: ", valA)
		}
	}

	if valB, foundB := ymc.Get("b"); !foundB {
		t.Error("error by getting data from yamc")
	} else {
		if valB != runCount {
			t.Error("error by getting data from yamc (b). unexpected value: ", valB)
		}
	}

}
