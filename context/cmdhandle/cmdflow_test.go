package cmdhandle_test

import (
	"os"
	"testing"

	"github.com/swaros/contxt/context/dirhandle"

	"github.com/swaros/contxt/context/cmdhandle"
)

// caseRunner helps to switch a testrunn in testcase directory to this
// this folder. and go back after the test is done
// it also resets all variables
// the id is just the number of the test/case folder (postfix)
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

func folderRunner(folder string, t *testing.T, testFunc func(t *testing.T)) {
	cmdhandle.ClearAll()
	old, derr := dirhandle.Current()
	if derr != nil {
		t.Error(derr)
	}
	os.Chdir(folder)
	testFunc(t)
	os.Chdir(old)

}

func TestVariableReset(t *testing.T) {
	folderRunner("./../../docs/test/valueRedefine", t, func(t *testing.T) {
		cmdhandle.RunTargets("case1", true)
		test1Result := cmdhandle.GetPH("RUN.case1.LOG.LAST")
		if test1Result != "initial" {
			t.Error("result 1 should be 'initial'.", test1Result)
		}

		cmdhandle.RunTargets("case2", true)
		test2Result := cmdhandle.GetPH("RUN.case2.LOG.LAST")
		if test2Result != "in-case-2" {
			t.Error("result 2 should be 'in-case-2'.", test1Result)
		}

		cmdhandle.RunTargets("case1,case2", true)
		test3Result := cmdhandle.GetPH("RUN.case2.LOG.LAST")
		if test3Result != "in-case-2" {
			t.Error("result 2 should be 'in-case-2'.", test1Result)
		}

		// testing main variables do not reset already changes variables
		cmdhandle.RunTargets("case2,case1", true)
		test4Result := cmdhandle.GetPH("RUN.case2.LOG.LAST")
		if test4Result != "in-case-2" {
			t.Error("result 2 should be 'in-case-2'.", test1Result)
		}

		test5Result := cmdhandle.GetPH("RUN.case1.LOG.LAST")
		if test5Result != "in-case-2" {
			t.Error("result 2 should be 'in-case-2'.", test1Result)
		}
	})
}

func TestCase0(t *testing.T) {
	caseRunner("0", t, func(t *testing.T) {
		cmdhandle.RunTargets("test1,test2", true)
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
		cmdhandle.RunTargets("case1_1,case1_2", true)
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
		cmdhandle.RunTargets("base", true)
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
		cmdhandle.RunTargets("base", true)
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
		cmdhandle.RunTargets("base", true)
		log := cmdhandle.GetPH("RUN.base.LOG.LAST")
		if log != "sub 4-6" {
			t.Error("last log entrie should not be:", log)
		}

		cmdhandle.RunTargets("contains", true)
		log = cmdhandle.GetPH("RUN.contains.LOG.LAST")
		if log != "come and die" {
			t.Error("last log entrie should not be", log)
		}
	})
}

func TestCase5(t *testing.T) {
	caseRunner("5", t, func(t *testing.T) {
		//contains a mutliline shell script
		cmdhandle.RunTargets("base", true)
		log := cmdhandle.GetPH("RUN.base.LOG.LAST")
		if log != "line4" {
			t.Error("last log entrie should not be", log)
		}
	})
}

// testing the thread run. do we wait for the subjobs also if they run longer then then main Task?
func TestCase6(t *testing.T) {
	caseRunner("6", t, func(t *testing.T) {
		cmdhandle.RunTargets("base", true)
		log := cmdhandle.GetPH("RUN.sub.LOG.LAST")
		if log != "sub-end" {
			t.Error("failed wait for ending subrun. last log entrie should be 'sub-end' got [", log, "] instead")
		}

	})
}

// testing error handling by script fails
func TestCase7(t *testing.T) {
	caseRunner("7", t, func(t *testing.T) {
		cmdhandle.RunTargets("base", true)
		logMain := cmdhandle.GetPH("RUN.base.LOG.LAST")
		if logMain != "done-main" {
			t.Error("last runstep should be excuted. but stopped on:", logMain)
		}

		log := cmdhandle.GetPH("RUN.sub.LOG.LAST")
		if log == "sub-end" {
			t.Error("the script runs without erros, but hey have an error. script have to stop. log=", log)
		}

	})
}

// test variables. replace set at config variables to hallo-welt
func TestCase8(t *testing.T) {
	caseRunner("8", t, func(t *testing.T) {
		cmdhandle.RunTargets("base", true)
		logMain := cmdhandle.GetPH("RUN.base.LOG.LAST")
		if logMain != "hallo-welt" {
			t.Error("variable should be replaced. but got:", logMain)
		}
	})
}

// test variables. replace set at config variables to hallo-welt but then overwrittn in task to hello-world
func TestCase9(t *testing.T) {
	caseRunner("9", t, func(t *testing.T) {
		cmdhandle.RunTargets("base,test2", true)
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

func TestCase12Requires(t *testing.T) {
	caseRunner("12", t, func(t *testing.T) {
		os.Setenv("TESTCASE_12_VAL", "HELLO")
		cmdhandle.RunTargets("test1,test2,test3,test4,test5,test6", true)
		logMain := cmdhandle.GetPH("RUN.test1.LOG.LAST")
		if logMain != "run_a" {
			t.Error("got unexpected result:", logMain)
		}

		test2 := cmdhandle.GetPH("RUN.test2.LOG.LAST")
		if test2 != "" {
			t.Error("got unexpected result for test2. got:", test2, "test should not run because of checking file")
		}

		test3 := cmdhandle.GetPH("RUN.test3.LOG.LAST")
		if test3 != "" {
			t.Error("got unexpected result for test3. got:", test3, "test should not run because env-var check")
		}

		test4 := cmdhandle.GetPH("RUN.test4.LOG.LAST")
		if test4 != "run_d" {
			t.Error("got unexpected result for test4. got:", test4, "test should run because env-var check")
		}

		test5 := cmdhandle.GetPH("RUN.test5.LOG.LAST")
		if test5 != "" {
			t.Error("got unexpected result for test5. got:", test5, "test should not run because variable check")
		}

		test6 := cmdhandle.GetPH("RUN.test6.LOG.LAST")
		if test6 != "run_f" {
			t.Error("got unexpected result for test6. got:", test6, "test should run because variable check")
		}
	})
}

func TestCase13Next(t *testing.T) {
	caseRunner("13", t, func(t *testing.T) {

		cmdhandle.RunTargets("start", true)
		logMain := cmdhandle.GetPH("RUN.start.LOG.LAST")
		if logMain != "start" {
			t.Error("got unexpected result:(", logMain, ")")
		}

		test2 := cmdhandle.GetPH("RUN.next_a.LOG.LAST")
		if test2 != "run-a" {
			t.Error("got unexpected result for test2. got:(", test2, ")")
		}

		test := cmdhandle.GetPH("RUN.next_b.LOG.LAST")
		if test != "run-b" {
			t.Error("got unexpected result for test3. got:(", test, ")")
		}

		test = cmdhandle.GetPH("RUN.next_c.LOG.LAST")
		if test != "run-c" {
			t.Error("got unexpected result for test4. got:(", test, ")")
		}

	})
}

func TestCase14Imports(t *testing.T) {
	caseRunner("14", t, func(t *testing.T) {

		cmdhandle.RunTargets("start,usertest", true)
		check := cmdhandle.GetJSONPathValueString("import.yaml", "testdata.check")
		if check != "hello" {
			t.Error("expect hello but got", check)
		}

		outcome := cmdhandle.GetPH("RUN.start.LOG.LAST")
		if outcome != "hello world" {
			t.Error("import looks not working. got ", outcome)
		}

		usertest := cmdhandle.GetPH("RUN.usertest.LOG.LAST")
		if usertest != "hello john miller" {
			t.Error("user data import looks not working. got ", usertest)
		}
	})
}

func TestCase14Needs(t *testing.T) {
	caseRunner("15", t, func(t *testing.T) {

		cmdhandle.RunTargets("start", true)

		usertest := cmdhandle.GetPH("RUN.start.LOG.LAST")
		needOneRuns := cmdhandle.GetPH("RUN.need_one.LOG.LAST")
		needTwoRuns := cmdhandle.GetPH("RUN.need_two.LOG.LAST")
		if needOneRuns != "done need_one" {
			t.Error("NEEDS needOne should be executed [done need_one] we got : ", needOneRuns)
		}
		if needTwoRuns != "done need_two" {
			t.Error("NEEDS needTwo should be executed [done need_two] we got : ", needTwoRuns)
		}

		if usertest != "the-main-task" {
			t.Error("did not get expected result instead [the-main-task] we got : ", usertest)
		}
	})
}

func TestStringMatcher(t *testing.T) {
	// positive expectations
	if !cmdhandle.StringMatchTest("=test", "test") {
		t.Error("expect TRUE, =test should match with test")
	}

	if !cmdhandle.StringMatchTest("test", "test") {
		t.Error("expect TRUE, test should not match with test")
	}

	if !cmdhandle.StringMatchTest(">50", "51") {
		t.Error("expect TRUE, 51 should accepted as greater then 50")
	}

	if !cmdhandle.StringMatchTest("<50", "49") {
		t.Error("expect TRUE, 49 should accepted as lower then 50")
	}
	if !cmdhandle.StringMatchTest("?", "something") {
		t.Error("expect TRUE, ? should match with anything")
	}

	// negative tests result expected
	if cmdhandle.StringMatchTest("=test", "test2") {
		t.Error("expect FALSE, =test should not match with test2")
	}

	if cmdhandle.StringMatchTest("!test", "test") {
		t.Error("expect FALSE, !test should (not) match with test by condition")
	}

	if cmdhandle.StringMatchTest(">test20", "test15") {
		t.Error("expect FALSE, test20 is greater then test15")
	}

	if cmdhandle.StringMatchTest("<test20", "test35") {
		t.Error("expect FALSE, test20 is lower then test35")
	}

	if cmdhandle.StringMatchTest("", "something") {
		t.Error("expect FALSE, empty should not match")
	}

	if !cmdhandle.StringMatchTest("", "") {
		t.Error("expect TRUE, empty should match to empty")
	}

	if cmdhandle.StringMatchTest("*", "") {
		t.Error("expect FALSE, not-empty (*) placeholder should not match with empty")
	}

	if !cmdhandle.StringMatchTest("*", "something") {
		t.Error("expect TRUE, not-empty placeholder should match with something")
	}
}
