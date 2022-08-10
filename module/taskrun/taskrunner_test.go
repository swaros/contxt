package taskrun_test

/*
import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/swaros/contxt/context/cmdhandle"
)

func TestBasicRun(t *testing.T) {
	// do not test experimental features
	if !cmdhandle.Experimental {
		return
	}
	var runner cmdhandle.TaskWatched
	runner.Init("testcase")
	counter := 10
	runner.Exec = func(state *cmdhandle.TaskWatched) cmdhandle.TaskResult {
		fmt.Println("this is working")
		// we just increase the counter, so we have something we can check, if the
		// function was executed
		counter += 10
		return cmdhandle.CreateTaskResult(nil)
	}

	// now starts
	if res := runner.Run(); res.Error != nil {
		t.Error(res.Error)
	}

	if counter == 10 {
		t.Error("seems body func was not executed. counter should be 60 but is still 10")
	}

	// we try to run a second time, now we should get an error
	if res := runner.Run(); res.Error == nil {
		t.Error("error expected. the second run should be not allowed")
	}
	// if the method would be executed twice, the counter should have increased
	if counter != 20 {
		t.Error("the function should not beeing executed and counter should be still 20. but is ", counter)
	}

	// retry the same without reporting an error, for cases the logic should accept multiple
	// trys to run the same target, but just ignore them

	runner.NoErrorIfBlocked = true
	// we try to run a third time, but now no error should be triggered
	if res := runner.Run(); res.Error != nil {
		t.Error(res.Error)
	}
	// but also on the third try, the method should just ignored and not increase the counter
	if counter != 20 {
		t.Error("the function should not beeing executed and counter should be still 20. but is ", counter)
	}

}

func TestDelayedTargets(t *testing.T) {
	// do not test experimental features
	if !cmdhandle.Experimental {
		return
	}
	tmoutHdnlTriggered := false

	var task cmdhandle.TaskWatched = cmdhandle.TaskWatched{
		TimeOutTiming: 1 * time.Millisecond,
		CanRunAgain: func(task *cmdhandle.TaskWatched) bool {
			return true
		},
		LoggerFnc: func(msg ...interface{}) {
			t.Log(msg...)
		},
		TimeOutHandler: func() {
			t.Log("timeout handler is triggered")
			tmoutHdnlTriggered = true
		},
		Exec: func(state *cmdhandle.TaskWatched) cmdhandle.TaskResult {
			// we simulate a very pure timeout handling
			// just by going out, if the timeoutHandler is
			// set the tmoutHdnlTriggered var to true
			for i := 0; i < 100; i++ {
				time.Sleep(5 * time.Millisecond)
				if tmoutHdnlTriggered {
					return cmdhandle.CreateTaskResult(nil)
				}
			}
			// if we are still in the game, something went wrong
			return cmdhandle.CreateTaskResult(errors.New("i should not reach this line"))
		},
	}
	task.Init("delayedTest")

	if tres := task.Run(); tres.Error != nil {
		t.Error(tres.Error)
	}

	if tmoutHdnlTriggered == false {
		t.Error("timerouthanler seems not beeing executed")
	}
}

// helper function to verify result of executions
func verifyTaskSlices(t *testing.T, messages []string, taskList []string) {
	// do not test experimental features
	if !cmdhandle.Experimental {
		return
	}
	if len(messages) != len(taskList) {
		t.Error("unexpected amount of messages ", len(messages), " expected ", len(taskList))
		t.Log(messages)

	}

	for _, looking := range taskList {
		hitNr := 0
		for _, msg := range messages {
			if strings.EqualFold(looking, msg) {
				hitNr++
			}
		}
		if hitNr == 0 {
			t.Error("missing task: ", looking)
		}
		if hitNr > 1 {
			t.Error(looking, " executes more then once. it runs ", hitNr, " times")
		}
	}
}

func TestTaskCreation(t *testing.T) {
	// do not test experimental features
	if !cmdhandle.Experimental {
		return
	}
	var messages []string
	var taskList []string = []string{"first", "second", "last"}
	taskHndl := cmdhandle.CreateMultipleTask(taskList, func(tw *cmdhandle.TaskWatched) {
		tw.Async = true
		tw.Exec = func(state *cmdhandle.TaskWatched) cmdhandle.TaskResult {
			messages = append(messages, state.GetName())
			fmt.Println(" --> [EXEC] append ", state.GetName(), " ", len(messages))
			state.ReportDone()
			return cmdhandle.CreateTaskResult(nil)
		}
		tw.LoggerFnc = func(msg ...interface{}) {
			fmt.Println(msg...)
		}

	})
	t.Log("run async start")
	taskHndl.Exec()
	taskHndl.Wait(10*time.Millisecond, 10*time.Second)
	t.Log("exec is done")
	verifyTaskSlices(t, messages, taskList)
}

func TestTaskCreationMixed(t *testing.T) {
	// do not test experimental features
	if !cmdhandle.Experimental {
		return
	}
	var messages []string
	var taskList []string = []string{"async-first-1", "async-second-2", "regular-one-3", "async-third-4", "regular-next-5", "regular-last-6"}
	taskHndl := cmdhandle.CreateMultipleTask(taskList, func(tw *cmdhandle.TaskWatched) {
		tw.Async = strings.Contains(tw.GetName(), "async")
		tw.Exec = func(state *cmdhandle.TaskWatched) cmdhandle.TaskResult {
			messages = append(messages, state.GetName())
			fmt.Println(" --> [EXEC]  append ", state.GetName(), " ", len(messages))
			return cmdhandle.CreateTaskResultContent(nil, "hello world")
		}
		tw.LoggerFnc = func(msg ...interface{}) {
			fmt.Println(msg...)
		}

	})
	t.Log("run async start")
	taskHndl.Exec()
	taskHndl.Wait(10*time.Millisecond, 10*time.Second)
	t.Log("exec is done")
	verifyTaskSlices(t, messages, taskList)
}

func TestTaskCreationAndTimeout(t *testing.T) {
	// do not test experimental features
	if !cmdhandle.Experimental {
		return
	}
	var taskGrp cmdhandle.TaskGroup = cmdhandle.TaskGroup{}
	hitTimout := false
	task2executed := false
	taskGrp.AddTask("test1", cmdhandle.TaskWatched{
		Async: true,
		Exec: func(tw *cmdhandle.TaskWatched) cmdhandle.TaskResult {
			for {
				time.Sleep(5 * time.Millisecond)
				if hitTimout {
					return cmdhandle.CreateTaskResult(nil)
				}
			}
		},
		NoErrorIfBlocked: true,
		TimeOutTiming:    10 * time.Millisecond,
		LoggerFnc: func(i ...interface{}) {
			t.Log(i...)
		},
		TimeOutHandler: func() {
			hitTimout = true
		},
	}).AddTask("Test2", cmdhandle.TaskWatched{
		Async: true,
		Exec: func(tw *cmdhandle.TaskWatched) cmdhandle.TaskResult {
			task2executed = true
			return cmdhandle.CreateTaskResultContent(nil, "dada")
		},
		ResultFnc: func(rs cmdhandle.TaskResult) {
			if rs.Content == nil {
				t.Error("result content is nil.", rs)
			} else {
				if !strings.EqualFold("dada", rs.Content.(string)) {
					t.Error("invalid result from second task.", rs)
				}
			}
		},
	}).Exec().Wait(1*time.Millisecond, 500*time.Second)

	if !hitTimout {
		t.Error("no timeout hit")
	}

	if !task2executed {
		t.Error("second task is not executed")
	}

}

func TestRuntimeCancelation(t *testing.T) {
	// do not test experimental features
	if !cmdhandle.Experimental {
		return
	}
	runA := false
	runB := false

	var taskGrp cmdhandle.TaskGroup = cmdhandle.TaskGroup{
		LoggerFnc: t.Log,
	}
	taskGrp.AddTask("first-run", cmdhandle.TaskWatched{
		Exec: func(tw *cmdhandle.TaskWatched) cmdhandle.TaskResult {
			runA = true
			tw.Log(" --- taskA here")
			time.Sleep(5 * time.Millisecond)
			return cmdhandle.CreateTaskResultContent(nil, "this works")
		},
		LoggerFnc: t.Log,
		Async:     true,
	}).AddTask("second-run", cmdhandle.TaskWatched{
		Exec: func(tw *cmdhandle.TaskWatched) cmdhandle.TaskResult {
			runB = true
			tw.Log(" --- taskB here")
			time.Sleep(15 * time.Millisecond)
			return cmdhandle.CreateTaskResult(nil)
		},
		LoggerFnc: t.Log,
		Async:     true,
	}).Exec().Wait(1*time.Microsecond, 500*time.Millisecond)

	if !runA {
		t.Error("first task is not executed")
	}

	if !runB {
		t.Error("second task is not executed")
	}
}

func TestSecondCallIsFine(t *testing.T) {
	// do not test experimental features
	if !cmdhandle.Experimental {
		return
	}
	counter := 0
	second := 0
	var taskGrp cmdhandle.TaskGroup = cmdhandle.TaskGroup{
		LoggerFnc: t.Log,
	}
	taskGrp.AddTask("BBBB-1", cmdhandle.TaskWatched{
		Exec: func(tw *cmdhandle.TaskWatched) cmdhandle.TaskResult {
			counter++
			return cmdhandle.CreateTaskResult(nil)
		},
	}).Exec().Wait(1*time.Microsecond, 500*time.Millisecond)

	var taskGrp2 cmdhandle.TaskGroup = cmdhandle.TaskGroup{
		LoggerFnc: t.Log,
	}
	taskGrp2.AddTask("BBBB-1", cmdhandle.TaskWatched{
		Exec: func(tw *cmdhandle.TaskWatched) cmdhandle.TaskResult {
			counter++
			second++
			return cmdhandle.CreateTaskResult(nil)
		},
	}).Exec().Wait(1*time.Microsecond, 500*time.Millisecond)

	if counter != 1 {
		t.Error("more (or less?) then one run:", counter)
	}

	if second != 0 {
		t.Error("second should not be > 0 because this method should never be executed:", second)
	}
}

func TestLayeredNeeds(t *testing.T) {
	// yes ... do not test experimental features....but this one is valid for booth cases
	// it is also the test that do not works in experimental state

	folderRunner("./../../docs/test/02needlayer", t, func(t *testing.T) {
		cmdhandle.RunTargets("main", true)
		test1Result := cmdhandle.GetPH("RUN.main.LOG.LAST")
		expectd := "main X123"
		orexpectd := "main X213"
		if test1Result != expectd && test1Result != orexpectd {
			t.Error("result should be [", expectd, "] but is [", test1Result, "]")
		}
	})
}
*/
