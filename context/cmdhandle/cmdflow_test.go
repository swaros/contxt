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

func TestCase5(t *testing.T) {
	caseRunner("5", t, func(t *testing.T) {
		//contains a mutliline shell script
		cmdhandle.RunTargets("base")
		log := cmdhandle.GetPH("RUN.base.LOG.LAST")
		if log != "line4" {
			t.Error("last log entrie should not be", log)
		}
	})
}

// testing the thread run. do we wait for the subjobs also if they run longer then then main Task?
func TestCase6(t *testing.T) {
	caseRunner("6", t, func(t *testing.T) {
		cmdhandle.RunTargets("base")
		log := cmdhandle.GetPH("RUN.sub.LOG.LAST")
		if log != "sub-end" {
			t.Error("failed wait for ending subrun. last log entrie should be 'sub-end' got [", log, "] instead")
		}

	})
}

// testing error handling by script fails
func TestCase7(t *testing.T) {
	caseRunner("7", t, func(t *testing.T) {
		cmdhandle.RunTargets("base")
		logMain := cmdhandle.GetPH("RUN.base.LOG.LAST")
		if logMain != "done-main" {
			t.Error("last runstep should be excuted. but stopped on:", logMain)
		}

		log := cmdhandle.GetPH("RUN.sub.LOG.LAST")
		if log == "sub-end" {
			t.Error("the script runs without erros, but hey have an error. script have to stop")
		}

	})
}

// test variables. replace set at config variables to hallo-welt
func TestCase8(t *testing.T) {
	caseRunner("8", t, func(t *testing.T) {
		cmdhandle.RunTargets("base")
		logMain := cmdhandle.GetPH("RUN.base.LOG.LAST")
		if logMain != "hallo-welt" {
			t.Error("variable should be replaced. but got:", logMain)
		}
	})
}

// test variables. replace set at config variables to hallo-welt but then overwrittn in task to hello-world
func TestCase9(t *testing.T) {
	caseRunner("9", t, func(t *testing.T) {
		cmdhandle.RunTargets("base,test2")
		logMain := cmdhandle.GetPH("RUN.base.LOG.LAST")
		if logMain != "hello-world" {
			t.Error("variable should be replaced. but got:", logMain)
		}
		test2 := cmdhandle.GetPH("RUN.test2.LOG.LAST")
		if test2 != "lets go" {
			t.Error("placeholder was not used in task variables. got:", test2)
		}
	})
}
