package cmdhandle

import (
	"testing"
	"time"
)

func TestTimout(t *testing.T) {
	ResetAllTaskInfos()
	incTaskCount("testTask")
	incTaskCount("testTask")

	waitHits := 0
	timeOutHit := false

	var tasks []string
	tasks = append(tasks, "testTask")
	WaitForTasksDone(tasks, 30*time.Millisecond, time.Millisecond, func() bool {
		waitHits++
		return true
	}, func() {}, func() {
		// timeout
		timeOutHit = true
	}, func(targetFull string, target string, args map[string]string) bool {
		return true
	})

	if waitHits == 0 {
		t.Error("never waits for complete the tasks..", waitHits)
	}

	if timeOutHit == false {
		t.Error("timeout never called..", waitHits)
	}
}

/*
func TestRegular(t *testing.T) {
	ResetAllTaskInfos()
	incTaskCount("testTask")
	incTaskCount("testTask")
	incTaskCount("testTask")
	incTaskCount("testTask")

	waitHits := 0
	timeOutHit := false
	doneCalled := false

	var tasks []string
	tasks = append(tasks, "testTask")
	WaitForTasksDone(tasks, 30*time.Millisecond, time.Millisecond, func() bool {
		waitHits++
		incTaskDoneCount("testTask")
		return true
	}, func() {
		doneCalled = true
	}, func() {
		// timeout
		timeOutHit = true
	}, func(target string) bool {
		return true
	})

	if waitHits == 0 {
		t.Error("never waits for complete the tasks..", waitHits)
	}

	if timeOutHit == true {
		t.Error("timeout should not be called..", waitHits)
	}

	if doneCalled == false {
		t.Error("done should be called..", waitHits)
	}
}
*/
func TestNeverStarts(t *testing.T) {
	ResetAllTaskInfos()

	waitHits := 0
	timeOutHit := false
	doneCalled := false
	notStartedCalled := false

	var tasks []string
	tasks = append(tasks, "testTask")
	WaitForTasksDone(tasks, 30*time.Millisecond, time.Millisecond, func() bool {
		waitHits++
		//incTaskDoneCount("testTask")
		return true
	}, func() {
		doneCalled = true
	}, func() {
		// timeout
		timeOutHit = true
	}, func(targetFull string, target string, args map[string]string) bool {
		notStartedCalled = true
		return false
	})

	if waitHits != 0 {
		t.Error("there is nothing to wait for .. bit got wait hits: ", waitHits)
	}

	if timeOutHit == true {
		t.Error("timeout should not be called..", waitHits)
	}

	if doneCalled == false {
		t.Error("done should be called..", waitHits)
	}
	if notStartedCalled == false {
		t.Error("not started should be triggered..", waitHits)
	}
}
