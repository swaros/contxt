package cmdhandle_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/swaros/contxt/context/cmdhandle"
)

func TestBasicRun(t *testing.T) {
	var runner cmdhandle.TaskWatched
	runner.Init("testcase")
	counter := 10
	runner.Exec = func(state *cmdhandle.TaskWatched) error {
		fmt.Println("this is working")
		// we just increase the counter, so we have something we can check, if the
		// function was executed
		counter += 10
		return nil
	}

	// now starts
	if err := runner.Run(); err != nil {
		t.Error(err)
	}

	if counter == 10 {
		t.Error("seems body func was not executed. counter should be 60 but is still 10")
	}

	// we try to run a second time, now we should get an error
	if err := runner.Run(); err == nil {
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
	if err := runner.Run(); err != nil {
		t.Error(err)
	}
	// but also on the third try, the method should just ignored and not increase the counter
	if counter != 20 {
		t.Error("the function should not beeing executed and counter should be still 20. but is ", counter)
	}

}

func TestDelayedTargets(t *testing.T) {
	tmoutHdnlTriggered := false

	var task cmdhandle.TaskWatched = cmdhandle.TaskWatched{
		IsGlobalScope: true,
		TimeOutTiming: 1 * time.Millisecond,
		CanRun: func(task *cmdhandle.TaskWatched) bool {
			return true
		},
		LoggerFnc: func(msg ...interface{}) {
			t.Log(msg...)
		},
		TimeOutHandler: func() {
			t.Log("timeout handler is triggered")
			tmoutHdnlTriggered = true
		},
		Exec: func(state *cmdhandle.TaskWatched) error {
			// we simulate a very pure timeout handling
			// just by going out, if the timeoutHandler is
			// set the tmoutHdnlTriggered var to true
			for i := 0; i < 100; i++ {
				time.Sleep(5 * time.Millisecond)
				if tmoutHdnlTriggered {
					return nil
				}
			}
			// if we are still in the game, something went wrong
			return errors.New("i should not reach this line")
		},
	}
	task.Init("delayedTest")

	if err := task.Run(); err != nil {
		t.Error(err)
	}

	if tmoutHdnlTriggered == false {
		t.Error("timerouthanler seems not beeing executed")
	}
}

func TestTaskCreation(t *testing.T) {
	var messages []string
	var taskList []string = []string{"first", "second", "last"}
	taskHndl := cmdhandle.CreateMultipleTask(taskList, func(tw *cmdhandle.TaskWatched) {
		tw.Async = true
		tw.Exec = func(state *cmdhandle.TaskWatched) error {
			messages = append(messages, state.GetName())
			fmt.Println(" append ", state.GetName(), " ", len(messages))
			state.ReportDone()
			return nil
		}
		tw.LoggerFnc = func(msg ...interface{}) {
			fmt.Println(msg...)
		}

	})
	t.Log("run async start")
	taskHndl.Exec()
	taskHndl.Wait()
	t.Log("exec is done")
	if len(messages) != 3 {
		t.Error("unexpected amount of messages ", len(messages))
	}
}

func TestTaskCreationMixed(t *testing.T) {
	var messages []string
	var taskList []string = []string{"async-first", "async-second", "regular-one", "async-third", "regular-next", "regular-last"}
	taskHndl := cmdhandle.CreateMultipleTask(taskList, func(tw *cmdhandle.TaskWatched) {
		tw.Async = strings.Contains(tw.GetName(), "async")
		tw.Exec = func(state *cmdhandle.TaskWatched) error {
			messages = append(messages, state.GetName())
			fmt.Println(" append ", state.GetName(), " ", len(messages))
			state.ReportDone()
			return nil
		}
		tw.LoggerFnc = func(msg ...interface{}) {
			fmt.Println(msg...)
		}

	})
	t.Log("run async start")
	taskHndl.Exec()
	taskHndl.Wait()
	t.Log("exec is done")
	if len(messages) != len(taskList) {
		t.Error("unexpected amount of messages ", len(messages))
	}
}
