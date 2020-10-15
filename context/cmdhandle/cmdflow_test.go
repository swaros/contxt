package cmdhandle_test

import (
	"os"
	"testing"

	"github.com/swaros/contxt/context/dirhandle"

	"github.com/swaros/contxt/context/cmdhandle"
)

func caseRunner(id string, t *testing.T, testFunc func(t *testing.T)) {
	cmdhandle.ClearAll()
	old, derr := dirhandle.Current()
	if derr != nil {
		t.Error(derr)
	}
	os.Chdir("./../../docs/test/case" + id)
	testFunc(t)
	os.Chdir(old)

}

func TestCase0(t *testing.T) {
	caseRunner("0", t, func(t *testing.T) {
		cmdhandle.RunTargets("test1,test2")
		test1Result := cmdhandle.GetPH("RUN.test1.LOG.LAST")
		if test1Result == "" {
			t.Error("result 1 should not be empty.", test1Result)
		}

		test2Result := cmdhandle.GetPH("RUN.test2.LOG.LAST")
		if test2Result == "" {
			t.Error("result 2 should not be empty.", test2Result)
		}

		if test2Result != "runs" {
			t.Error("result 2 should be 'runs' instead we got.", test2Result)
		}
	})
}

func TestRunTargetCase1(t *testing.T) {

	caseRunner("1", t, func(t *testing.T) {
		cmdhandle.RunTargets("case1_1,case1_2")
		test1Result := cmdhandle.GetPH("RUN.case1_1.LOG.LAST")
		if test1Result == "" {
			t.Error("result 1 should not be empty.", test1Result)
		}

		test2Result := cmdhandle.GetPH("RUN.case1_2.LOG.LAST")
		if test2Result == "" {
			t.Error("result 2 should not be empty.", test2Result)
		}

		if test2Result != "runs" {
			t.Error("result 2 should be 'runs' instead we got.", test2Result)
		}

		scriptLast := cmdhandle.GetPH("RUN.SCRIPT_LINE")
		if scriptLast != "echo 'runs'" {
			t.Error("unexpected result [", scriptLast, "]")
		}
	})

}

func TestRunTargetCase2(t *testing.T) {

	caseRunner("2", t, func(t *testing.T) {
		cmdhandle.RunTargets("base")
		test1Result := cmdhandle.GetPH("RUN.base.LOG.HIT")
		if test1Result != "start-task-2" {
			t.Error("unexpected result ", test1Result)
		}

		test2Result := cmdhandle.GetPH("RUN.task-2.LOG.LAST")
		if test2Result != "im-task-2" {
			t.Error("unexpected result [", test2Result, "]")
		}
	})
}

func TestRunTargetCase3(t *testing.T) {
	// testing PID of my own and the parent process
	caseRunner("3", t, func(t *testing.T) {
		cmdhandle.RunTargets("base")
		test1Result := cmdhandle.GetPH("RUN.base.LOG.HIT")
		if test1Result != "launch" {
			t.Error("unexpected result ", test1Result)
		}
		pid_1 := cmdhandle.GetPH("RUN.base.LOG.LAST")
		pid_2 := cmdhandle.GetPH("RUN.task-2.LOG.LAST")
		if pid_2 != pid_1 {
			t.Error("PID should be the same [", pid_1, " != ", pid_2, "]")
		}
	})

}

func TestCase4(t *testing.T) {
	caseRunner("4", t, func(t *testing.T) {
		//stopped because log entrie to big
		cmdhandle.RunTargets("base")
		log := cmdhandle.GetPH("RUN.base.LOG.LAST")
		if log != "sub 4-6" {
			t.Error("last log entrie should not be:", log)
		}

		cmdhandle.RunTargets("contains")
		log = cmdhandle.GetPH("RUN.contains.LOG.LAST")
		if log != "come and die" {
			t.Error("last log entrie should not be", log)
		}
	})
}
