package tasks_test

import (
	"strings"
	"testing"
	"time"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
)

func TestRunnerBasic(t *testing.T) {

	var testTask configure.Task = configure.Task{
		ID: "testsleep", // we need an ID for the task
	}
	task := tasks.New(testTask.ID, nil)

	messages := []string{}
	runner, err := task.GetRunnerForTask(testTask, func(s string, err error) bool {
		if err != nil {
			t.Error(err)
		}
		messages = append(messages, s)
		return true
	})
	if err != nil {
		t.Error(err)
	}
	if runner == nil {
		t.Error("runner is nil")
	} else {
		if err := runner.GetRunner().Command("ls -ga"); err != nil {
			t.Error(err)
		} else {
			// wait to the task get running have to be very fast
			// because some tasks are very fast started and done.
			// so this test can fail if the system is faster then the system was using while writing the test
			// and 10 nanoseconds is then to long to wait.
			// so we will ignore the false result for testing
			task.WaitTilTaskRunnerIsRunning(testTask, 1*time.Millisecond, 100) // so first make sure the task is running. at least wait 100 millisecond
			task.WaitTilTaskRunnerIsDone(testTask, 10*time.Millisecond)        // and then wait til it is done
			if task.RunnersActive() {                                          // this shuld be obvious but we check it anyway
				t.Error("runner should not be active anymore")
			}
			// messages should contains the names of the files in the current directory
			// so we check if the first message contains the string "runner_test.go"

			if len(messages) == 0 {
				t.Error("no messages received")
			} else {
				expected := []string{"runners_test.go", "..", ".", "watchman.go", "watchmanTask.go"}
				for _, e := range expected {
					if !systools.SliceContainsSub(messages, e) {
						t.Error("missing", e, "in messages:", strings.Join(messages, "\n"))
					}
				}
			}
			task.StopAllTaskRunner() // at least stop the task runner

		}
	}

}

func TestRepeatedBasic(t *testing.T) {
	countOfTest := 15
	for i := 0; i < countOfTest; i++ {
		TestRunnerBasic(t)
	}
}

func TestRunnerCmd(t *testing.T) {

	var testTask configure.Task = configure.Task{
		ID: "testsleep", // we need an ID for the task
	}
	task := tasks.New(testTask.ID, nil)
	messages := []string{}
	runner, err := task.GetRunnerForTask(testTask, func(s string, err error) bool {
		if err != nil {
			t.Error(err)
		}
		messages = append(messages, s)
		return true
	})
	if err != nil {
		t.Error(err)
	}
	if runner == nil {
		t.Error("runner is nil")
	} else {
		if err := runner.Cmd("ls -ga"); err != nil {
			t.Error(err)
		} else {
			if len(messages) == 0 {
				t.Error("no messages received")
			} else {
				expected := []string{"runners_test.go", "..", ".", "watchman.go", "watchmanTask.go"}
				for _, e := range expected {
					if !systools.SliceContainsSub(messages, e) {
						t.Error("missing", e, "in messages:", strings.Join(messages, "\n"))
					}
				}
			}
			// now we run a different command
			messages = []string{}
			if err := runner.Cmd("ls .."); err != nil {
				t.Error(err)
			} else {
				if len(messages) == 0 {
					t.Error("no messages received")
				} else {
					expected := []string{"configure", "ctemplate", "mimiclog", "yamc"}
					for _, e := range expected {
						if !systools.SliceContainsSub(messages, e) {
							t.Error("should have", e, "in messages:", strings.Join(messages, "\n"))
						}
					}
				}
			}
		}
	}

}
