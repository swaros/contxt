package cmdhandle_test

import (
	"sync"
	"testing"
	"time"

	"github.com/swaros/contxt/context/cmdhandle"
)

func TestSetAndGet(t *testing.T) {
	cmdhandle.SetPH("lost1", "data")
	value := cmdhandle.GetPH("lost1")
	if value != "data" {
		t.Error("unexpected result:'", value, "' expected was 'data'")
	}

	cmdhandle.ClearAll()
	valueAfterClean := cmdhandle.GetPH("lost1")
	if valueAfterClean != "" {
		t.Error("unexpected result:'", valueAfterClean, "' should be empty after ClearAll")
	}
}

func TestOverwrite(t *testing.T) {
	cmdhandle.SetPH("check", "flower")
	cmdhandle.SetPH("check", "main")
	value := cmdhandle.GetPH("check")
	if value != "main" {
		t.Error("unexpected result:'", value, "' expected was 'main'")
	}
}

func TestNotExists(t *testing.T) {
	nonExists := cmdhandle.GetPH("whatever")
	if nonExists != "" {
		t.Error("unexpected result: [", nonExists, "] this sould be a empty string")
	}

}
func TestBasicReplace(t *testing.T) {

	cmdhandle.SetPH("test1", "here i am")
	cmdhandle.SetPH("test2", "XXX")

	testLine := "a: ${test1}"
	testLine2 := "b: ${test2} and again ${test2}"

	result := cmdhandle.HandlePlaceHolder(testLine)
	if result != "a: here i am" {
		t.Error("noting was replaced:'", testLine, "' => ", result)
	}

	result2 := cmdhandle.HandlePlaceHolder(testLine2)
	if result2 != "b: XXX and again XXX" {
		t.Error("noting was replaced:'", testLine2, "' => ", result2)
	}

}

func TestAsync(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go WriteToMap(&wg, "async", "check2", 500, 100*time.Microsecond)
	wg.Add(1)
	go ReadFromMap(&wg, "async", 500, 80*time.Microsecond)
	wg.Add(1)
	go WriteToMap(&wg, "async", "check3", 300, 200*time.Microsecond)
	wg.Add(1)
	go ReadFromMap(&wg, "async", 280, 150*time.Microsecond)
	wg.Wait()
}

func WriteToMap(waitGroup *sync.WaitGroup, key, value string, loops int, wait time.Duration) {
	defer waitGroup.Done()
	for i := 0; i < loops; i++ {
		cmdhandle.SetPH(key, value)
		time.Sleep(wait)
	}
}

func ReadFromMap(waitGroup *sync.WaitGroup, key string, loops int, wait time.Duration) {
	defer waitGroup.Done()
	for i := 0; i < loops; i++ {
		cmdhandle.GetPH(key)
		time.Sleep(wait)
	}
}
