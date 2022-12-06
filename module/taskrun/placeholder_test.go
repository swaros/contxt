package taskrun_test

import (
	"testing"

	"github.com/swaros/contxt/module/taskrun"
)

func TestSetAndGet(t *testing.T) {
	taskrun.SetPH("lost1", "data")
	value := taskrun.GetPH("lost1")
	if value != "data" {
		t.Error("unexpected result:'", value, "' expected was 'data'")
	}

	taskrun.ClearAll()
	valueAfterClean := taskrun.GetPH("lost1")
	if valueAfterClean != "" {
		t.Error("unexpected result:'", valueAfterClean, "' should be empty after ClearAll")
	}
}

func TestOverwrite(t *testing.T) {
	taskrun.SetPH("check", "flower")
	taskrun.SetPH("check", "main")
	value := taskrun.GetPH("check")
	if value != "main" {
		t.Error("unexpected result:'", value, "' expected was 'main'")
	}
}

func TestNotExists(t *testing.T) {
	nonExists := taskrun.GetPH("whatever")
	if nonExists != "" {
		t.Error("unexpected result: [", nonExists, "] this sould be a empty string")
	}

}
func TestBasicReplace(t *testing.T) {

	taskrun.SetPH("test1", "here i am")
	taskrun.SetPH("test2", "XXX")

	testLine := "a: ${test1}"
	testLine2 := "b: ${test2} and again ${test2}"

	result := taskrun.HandlePlaceHolder(testLine)
	if result != "a: here i am" {
		t.Error("noting was replaced:'", testLine, "' => ", result)
	}

	result2 := taskrun.HandlePlaceHolder(testLine2)
	if result2 != "b: XXX and again XXX" {
		t.Error("noting was replaced:'", testLine2, "' => ", result2)
	}

}

/* that seems an relic. but have to think about*/
/*
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
		taskrun.SetPH(key, value)
		time.Sleep(wait)
	}
}

func ReadFromMap(waitGroup *sync.WaitGroup, key string, loops int, wait time.Duration) {
	defer waitGroup.Done()
	for i := 0; i < loops; i++ {
		taskrun.GetPH(key)
		time.Sleep(wait)
	}
}
*/
