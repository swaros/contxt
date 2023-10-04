package taskrun_test

import (
	"os"
	"strings"
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/taskrun"
)

type RuntimeGroupExpected struct {
	Contains []string
}

type TestRuntimeGroup struct {
	tests []RuntimeGroupExpected
}

func listHaveString(lookFor string, list []string) bool {
	for _, s := range list {
		if s == lookFor {
			return true
		}
	}
	return false
}

func listContainsEach(lista, listb []string) (string, bool) {
	for _, chk := range lista {
		if !listHaveString(chk, listb) {
			return chk, false
		}
	}
	return "", true
}

func sliceList(offset int, list []string) ([]string, []string) {
	var newList []string
	var restList []string
	for i, cont := range list {
		if i < offset {
			newList = append(newList, cont)
		} else {
			restList = append(restList, cont)
		}
	}
	return newList, restList
}

func assertRuntimeGroup(t *testing.T, path string, target string, testGroup TestRuntimeGroup) bool {
	tresult := true
	folderRunner(path, t, func(t *testing.T) {
		taskrun.RunTargets(target, true)
		result := taskrun.GetPH("teststr")
		resultArr := strings.Split(result, ":")

		//offset := 0
		for tbIndex, runtimeCheck := range testGroup.tests {
			var checkList []string
			testLen := len(runtimeCheck.Contains) // how many entries we have to check
			if testLen <= len(resultArr) {        // do we have enough entries in the result array?
				checkList, resultArr = sliceList(len(runtimeCheck.Contains), resultArr)
				if missingStr, ok := listContainsEach(checkList, runtimeCheck.Contains); ok == false {
					t.Error("missing ", missingStr, " in test block ", tbIndex, " got ", checkList, " expected", runtimeCheck.Contains)
					tresult = false
				}
			}
		}

	})

	return tresult
}

func TestIssue58(t *testing.T) {
	folderRunner("./../../docs/test/issue58", t, func(t *testing.T) {

		taskrun.RunTargets("test1", true)
		res := taskrun.GetPH("RUN.test1.LOG.LAST")
		if res != "full check john miller" {
			t.Error("unexpected result [", res, "]")
		}

		taskrun.RunTargets("start", true)
		test1Result := taskrun.GetPH("RUN.start.LOG.LAST")
		if test1Result != "hello world" {
			t.Error("result 2 should be 'hello world'.[", test1Result, "]")
		}

		taskrun.RunTargets("replace", true)
		rep := taskrun.GetPH("RUN.replace.LOG.LAST")
		if rep != "[testcase /usr - mother]" {
			t.Error("the replacing of inline variables in defined variables is not working")
		}
	})
}

func TestMultipleTargets(t *testing.T) {
	folderRunner("./../../docs/test/01multi", t, func(t *testing.T) {
		taskrun.RunTargets("task", true)
		test1Result := taskrun.GetPH("RUN.task.LOG.LAST")
		if test1Result != "hello 2" {
			t.Error("result 2 should be 'hello 2'.", test1Result)
		}
	})
}

func TestMultipleTargetsOs(t *testing.T) {
	folderRunner("./../../docs/test/02multi", t, func(t *testing.T) {
		taskrun.RunTargets("task", true)
		test1Result := taskrun.GetPH("RUN.task.LOG.LAST")
		if configure.GetOs() == "linux" {
			if test1Result != "hello linux" {
				t.Error("result 2 should be 'hello 2'. got[", test1Result, "]")
			}
		}
	})
}

func TestRunIfEquals(t *testing.T) {
	folderRunner("./../../docs/test/ifequals", t, func(t *testing.T) {
		taskrun.RunTargets("check-eq", true)
		test1Result := taskrun.GetPH("RUN.check-eq.LOG.LAST")
		if test1Result != "inline" {
			t.Error("result 2 should be 'inline'.", test1Result)
		}

		taskrun.RunTargets("check-noeq", true)
		test2Result := taskrun.GetPH("RUN.check-noeq.LOG.LAST")
		if test2Result != "start2" {
			t.Error("result 2 should be 'start2'.", test2Result)
		}
	})
}

func TestVariableReset(t *testing.T) {
	folderRunner("./../../docs/test/valueRedefine", t, func(t *testing.T) {
		taskrun.RunTargets("case1", true)
		test1Result := taskrun.GetPH("RUN.case1.LOG.LAST")
		if test1Result != "initial" {
			t.Error("result 1 should be 'initial'.", test1Result)
		}

		taskrun.RunTargets("case2", true)
		test2Result := taskrun.GetPH("RUN.case2.LOG.LAST")
		if test2Result != "in-case-2" {
			t.Error("result 2 should be 'in-case-2'.", test1Result)
		}

		taskrun.RunTargets("case1,case2", true)
		test3Result := taskrun.GetPH("RUN.case2.LOG.LAST")
		if test3Result != "in-case-2" {
			t.Error("result 2 should be 'in-case-2'.", test1Result)
		}

		// testing main variables do not reset already changes variables
		taskrun.RunTargets("case2,case1", true)
		test4Result := taskrun.GetPH("RUN.case2.LOG.LAST")
		if test4Result != "in-case-2" {
			t.Error("result 2 should be 'in-case-2'.", test1Result)
		}

		test5Result := taskrun.GetPH("RUN.case1.LOG.LAST")
		if test5Result != "in-case-2" {
			t.Error("result 2 should be 'in-case-2'.", test1Result)
		}
	})
}

func TestCase0(t *testing.T) {
	caseRunner("0", t, func(t *testing.T) {
		taskrun.RunTargets("test1,test2", true)
		test1Result := taskrun.GetPH("RUN.test1.LOG.LAST")
		if test1Result == "" {
			t.Error("result 1 should not be empty.", test1Result)
		}

		test2Result := taskrun.GetPH("RUN.test2.LOG.LAST")
		if test2Result == "" {
			t.Error("result 2 should not be empty.", test2Result)
		}

		if test2Result != "runs" {
			t.Error("result 2 should be 'runs' instead we got.", "["+test2Result+"]")
		}
	})
}

func TestRunTargetCase1(t *testing.T) {

	caseRunner("1", t, func(t *testing.T) {
		taskrun.RunTargets("case1_1,case1_2", true)
		test1Result := taskrun.GetPH("RUN.case1_1.LOG.LAST")
		if test1Result == "" {
			t.Error("result 1 should not be empty.", test1Result)
		}

		test2Result := taskrun.GetPH("RUN.case1_2.LOG.LAST")
		if test2Result == "" {
			t.Error("result 2 should not be empty.", test2Result)
		}

		if test2Result != "runs" {
			t.Error("result 2 should be 'runs' instead we got.", test2Result)
		}

		scriptLast := taskrun.GetPH("RUN.SCRIPT_LINE")
		if scriptLast != "echo runs" {
			t.Error("unexpected result [", scriptLast, "]")
		}
	})

}

func TestRunTargetCase2(t *testing.T) {

	caseRunner("2", t, func(t *testing.T) {
		taskrun.RunTargets("base", true)
		test1Result := taskrun.GetPH("RUN.base.LOG.HIT")
		if test1Result != "start-task-2" {
			t.Error("unexpected result ", test1Result)
		}

		test2Result := taskrun.GetPH("RUN.task-2.LOG.LAST")
		if test2Result != "im-task-2" {
			t.Error("unexpected result [", test2Result, "]")
		}
	})
}

// TBH: not sure if these is not too specific
/*
func TestRunTargetCase3(t *testing.T) {
	// testing PID of my own and the parent process
	if configure.GetOs() == "windows" {
		return
	}
	caseRunner("3", t, func(t *testing.T) {
		taskrun.RunTargets("base", true)
		test1Result := taskrun.GetPH("RUN.base.LOG.HIT")
		if test1Result != "launch" {
			t.Error("unexpected result ", test1Result)
		}
		pid_1 := taskrun.GetPH("RUN.base.LOG.LAST")
		pid_2 := taskrun.GetPH("RUN.task-2.LOG.LAST")
		if pid_2 != pid_1 {
			t.Error("PID should be the same [", pid_1, " != ", pid_2, "]")
		}
	})

}*/

func TestCase4(t *testing.T) {
	caseRunner("4", t, func(t *testing.T) {
		//stopped because log entrie to big
		taskrun.RunTargets("base", true)
		log := taskrun.GetPH("RUN.base.LOG.LAST")
		if log != "sub 4-6" {
			t.Error("last log entrie should not be:", log)
		}

		taskrun.RunTargets("contains", true)
		log = taskrun.GetPH("RUN.contains.LOG.LAST")
		if log != "come and die" {
			t.Error("last log entrie should not be", log)
		}
	})
}

func TestCase5(t *testing.T) {
	caseRunner("5", t, func(t *testing.T) {
		//contains a mutliline shell script
		taskrun.RunTargets("base", true)
		log := taskrun.GetPH("RUN.base.LOG.LAST")
		if log == "" {
			t.Error("got empty result. that is not expected")
		}
		if log != "line4" {
			t.Error("expected 'line4'...but got ", log)
		}
	})
}

// testing the thread run. do we wait for the subjobs also if they run longer then then main Task?
func TestCase6(t *testing.T) {
	caseRunner("6", t, func(t *testing.T) {
		taskrun.RunTargets("base", true)
		log := taskrun.GetPH("RUN.sub.LOG.LAST")
		if log != "sub-end" {
			t.Error("failed wait for ending subrun. last log entrie should be 'sub-end' got [", log, "] instead")
		}

	})
}

// testing error handling by script fails
func TestCase7(t *testing.T) {
	caseRunner("7", t, func(t *testing.T) {
		taskrun.RunTargets("base", true)
		logMain := taskrun.GetPH("RUN.base.LOG.LAST")
		if logMain != "done-main" {
			t.Error("last runstep should be excuted. but stopped on:", logMain)
		}

		log := taskrun.GetPH("RUN.sub.LOG.LAST")
		if log == "sub-end" {
			t.Error("the script runs without erros, but hey have an error. script have to stop. log=", log)
		}

	})
}

// testing on error behavior for the stop reasons
func TestCase71(t *testing.T) {
	assertCaseLogLastEquals(t, "7", "trigger", "two")                   // testing outpout match
	assertCaseLogLastEquals(t, "7", "len", "123")                       // test onoutLess
	assertCaseLogLastEquals(t, "7", "lenmore", "b123456abcdefghijklmn") // test onoutMore
}

// test variables. replace set at config variables to hallo-welt
func TestCase8(t *testing.T) {
	caseRunner("8", t, func(t *testing.T) {
		taskrun.RunTargets("base", true)
		logMain := taskrun.GetPH("RUN.base.LOG.LAST")
		if logMain != "hallo-welt" {
			t.Error("variable should be replaced. but got:", logMain)
		}
	})
}

// test variables. replace set at config variables to hallo-welt but then overwrittn in task to hello-world
func TestCase9(t *testing.T) {
	caseRunner("9", t, func(t *testing.T) {
		taskrun.RunTargets("base,test2", true)
		logMain := taskrun.GetPH("RUN.base.LOG.LAST")
		if logMain != "hello-world" {
			t.Error("variable should be replaced. but got:", logMain)
		}
		test2 := taskrun.GetPH("RUN.test2.LOG.LAST")
		if test2 != "lets go" {
			t.Error("placeholder was not used in task variables. got:[", test2, "]")
		}
	})
}

func TestCase12Requires(t *testing.T) {
	caseRunner("12", t, func(t *testing.T) {
		os.Setenv("TESTCASE_12_VAL", "HELLO")
		taskrun.RunTargets("test1,test2,test3,test4,test5,test6", true)
		logMain := taskrun.GetPH("RUN.test1.LOG.LAST")
		if logMain != "run_a" {
			t.Error("got unexpected result:", logMain)
		}

		test2 := taskrun.GetPH("RUN.test2.LOG.LAST")
		if test2 != "" {
			t.Error("got unexpected result for test2. got:", test2, "test should not run because of checking file")
		}

		test3 := taskrun.GetPH("RUN.test3.LOG.LAST")
		if test3 != "" {
			t.Error("got unexpected result for test3. got:", test3, "test should not run because env-var check")
		}

		test4 := taskrun.GetPH("RUN.test4.LOG.LAST")
		if test4 != "run_d" {
			t.Error("got unexpected result for test4. got:", test4, "test should run because env-var check")
		}

		test5 := taskrun.GetPH("RUN.test5.LOG.LAST")
		if test5 != "" {
			t.Error("got unexpected result for test5. got:", test5, "test should not run because variable check")
		}

		varValue := taskrun.GetPH("test_var")
		if varValue == "HELLO_KLAUS" {
			t.Error("Expected value 'HELLO_KLAUS' not in placeholder 'test_var'. got instead: ", varValue)
		}
		test6 := taskrun.GetPH("RUN.test6.LOG.LAST")
		if test6 != "" {
			t.Error("got unexpected result for test6. got:[", test6, "]test should run because variable check")
		}
	})
}

func TestCase13Next(t *testing.T) {
	caseRunner("13", t, func(t *testing.T) {

		taskrun.RunTargets("start", true)
		logMain := taskrun.GetPH("RUN.start.LOG.LAST")
		if logMain != "start" {
			t.Error("got unexpected result:(", logMain, ")")
		}

		test2 := taskrun.GetPH("RUN.next_a.LOG.LAST")
		if test2 != "run-a" {
			t.Error("got unexpected result for test2. got:(", test2, ")")
		}

		test := taskrun.GetPH("RUN.next_b.LOG.LAST")
		if test != "run-b" {
			t.Error("got unexpected result for test3. got:(", test, ")")
		}

		test = taskrun.GetPH("RUN.next_c.LOG.LAST")
		if test != "run-c" {
			t.Error("got unexpected result for test4. got:(", test, ")")
		}

	})
}

func TestCase14Imports(t *testing.T) {
	caseRunner("14", t, func(t *testing.T) {

		taskrun.RunTargets("start,usertest", true)
		check := taskrun.GetJSONPathValueString("import.yaml", "testdata.check")
		if check != "hello" {
			t.Error("expect hello but got", check)
		}

		outcome := taskrun.GetPH("RUN.start.LOG.LAST")
		if outcome != "hello world" {
			t.Error("import looks not working. got ", outcome)
		}

		usertest := taskrun.GetPH("RUN.usertest.LOG.LAST")
		if usertest != "hello john miller" {
			t.Error("user data import looks not working. got ", usertest)
		}
	})
}

func TestCase15Needs(t *testing.T) {
	caseRunner("15", t, func(t *testing.T) {

		taskrun.RunTargets("start", true)

		usertest := taskrun.GetPH("RUN.start.LOG.LAST")
		needOneRuns := taskrun.GetPH("RUN.need_one.LOG.LAST")
		needTwoRuns := taskrun.GetPH("RUN.need_two.LOG.LAST")
		if needOneRuns != "<<< 1 >>> done need_one" {
			t.Error("NEEDS needOne should be executed [done need_one] we got : ", needOneRuns)
		}
		if needTwoRuns != "<<< 2 >>> done need_two" {
			t.Error("NEEDS needTwo should be executed [done need_two] we got : ", needTwoRuns)
		}

		if usertest != "the-main-task" {
			t.Error("did not get expected result instead [the-main-task] we got : ", usertest)
		}
	})
}

func TestCase16WorkingDir(t *testing.T) {
	caseRunner("16", t, func(t *testing.T) {
		old, derr := dirhandle.Current()
		if derr != nil {
			t.Error(derr)
		}
		taskrun.RunTargets("origin_dir", true)
		origin_dir := clearStrings(taskrun.GetPH("RUN.origin_dir.LOG.LAST"))
		oldChkStr := clearStrings(old)

		// strToUpper required on windows because of differents in drive letter
		if !strings.EqualFold(strings.ToUpper(origin_dir), strings.ToUpper(oldChkStr)) {
			t.Error("do not get the expected folder ", oldChkStr, origin_dir)
		}

		taskrun.RunTargets("sub_a", true)
		expected := "testcase_run_check_7725569sjghfghf"
		current := taskrun.GetPH("RUN.sub_a.LOG.LAST")
		if current != expected {
			t.Error("we expected file content from test.txt in subfolder sub_a which is ", expected, " but got ", current)
		}

		newDir, nerr := dirhandle.Current()
		if nerr != nil {
			t.Error(derr)
		}

		if newDir != old {
			t.Error("directory have to beeing set back to old dir after rinning task with a workingdir. but we are still at ", newDir)
		}

	})
}

func TestStringMatcher(t *testing.T) {
	// positive expectations
	if !taskrun.StringMatchTest("=test", "test") {
		t.Error("expect TRUE, =test should match with test")
	}

	if !taskrun.StringMatchTest("test", "test") {
		t.Error("expect TRUE, test should not match with test")
	}

	if !taskrun.StringMatchTest(">50", "51") {
		t.Error("expect TRUE, 51 should accepted as greater then 50")
	}

	if !taskrun.StringMatchTest("<50", "49") {
		t.Error("expect TRUE, 49 should accepted as lower then 50")
	}
	if !taskrun.StringMatchTest("?", "something") {
		t.Error("expect TRUE, ? should match with anything")
	}

	// negative tests result expected
	if taskrun.StringMatchTest("=test", "test2") {
		t.Error("expect FALSE, =test should not match with test2")
	}

	if taskrun.StringMatchTest("!test", "test") {
		t.Error("expect FALSE, !test should (not) match with test by condition")
	}

	if taskrun.StringMatchTest(">test20", "test15") {
		t.Error("expect FALSE, test20 is greater then test15")
	}

	if taskrun.StringMatchTest("<test20", "test35") {
		t.Error("expect FALSE, test20 is lower then test35")
	}

	if taskrun.StringMatchTest("", "something") {
		t.Error("expect FALSE, empty should not match")
	}

	if !taskrun.StringMatchTest("", "") {
		t.Error("expect TRUE, empty should match to empty")
	}

	if taskrun.StringMatchTest("*", "") {
		t.Error("expect FALSE, not-empty (*) placeholder should not match with empty")
	}

	if !taskrun.StringMatchTest("*", "something") {
		t.Error("expect TRUE, not-empty placeholder should match with something")
	}
}

func TestNeedWithArgs(t *testing.T) {
	folderRunner("./../../docs/test/needWArgs", t, func(t *testing.T) {
		taskrun.RunTargets("test-need", false)

	})
}

func TestJsonExec(t *testing.T) {
	folderRunner("./../../docs/test/execjson", t, func(t *testing.T) {
		taskrun.RunTargets("test-load", false)
		found, json := taskrun.GetData("JSON")
		if !found {
			t.Error("expected to find JSON data")
		} else {
			if _, ok := json["0"]; !ok {
				t.Error("expected to find key 0 in JSON data")
			}
		}

		lastLogLine := taskrun.GetPH("RUN.test-load.LOG.LAST")
		if lastLogLine != "buildkit.dockerfile.v0" {
			t.Error("expected to find buildkit.dockerfile.v0 in log")
		}
	})
}

func TestConcurrent(t *testing.T) {
	expected := ""
	folderRunner("./../../docs/test/01concurrent", t, func(t *testing.T) {
		taskrun.RunTargets("main_a", false)
		main_a := taskrun.GetPH("teststr")
		expected = "BASE:MA:"
		if main_a != expected {
			t.Error("expected:", expected, " instead:", main_a)
		}
	})
}

func TestConcurrentMainB(t *testing.T) {
	var test TestRuntimeGroup = TestRuntimeGroup{
		[]RuntimeGroupExpected{
			{
				Contains: []string{"BASE"},
			},
			{
				Contains: []string{"NB", "NA", "NC"},
			},
			{
				Contains: []string{"MB"},
			},
		},
	}
	assertRuntimeGroup(t, "./../../docs/test/01concurrent", "main_b", test)
}

func TestConcurrentMainC(t *testing.T) {
	var test TestRuntimeGroup = TestRuntimeGroup{
		[]RuntimeGroupExpected{
			{
				Contains: []string{"BASE"}, // base fiirst at all
			},
			{
				Contains: []string{"NB", "NA", "NC"}, // needs as second
			},
			{
				Contains: []string{"MC", "TC", "TA", "TB"}, //anything else at the end unordered MC:TC:TA:TB:
			},
		},
	}
	assertRuntimeGroup(t, "./../../docs/test/01concurrent", "main_c", test)
}

func TestConcurrentMainD(t *testing.T) {
	// BASE:NB:MC:TC:NA:TA:TB:NC:MD:
	/*
	 - id: main_d
	    needs:
	      - need_a
	      - main_c
	      - need_c
	    script:
	      - "#@add teststr MD:"
	      - echo ${teststr}
	*/
	var test TestRuntimeGroup = TestRuntimeGroup{
		[]RuntimeGroupExpected{
			{
				Contains: []string{"BASE"}, // base first at all
			},
			{
				Contains: []string{"NB"}, // need b first as it is a need of main_c
			},
			{
				Contains: []string{"MC", "NC", "NA", "TC", "TA", "TB"}, //anything else at the end unordered
			},
			{
				Contains: []string{"MD"}, // the last is main_d
			},
		},
	}

	// have to investigate why, but on windows
	// the order of execution seems different.
	// but the imortant order seems beeing intact
	if configure.GetOs() == "windows" {
		test = TestRuntimeGroup{
			[]RuntimeGroupExpected{
				{
					Contains: []string{"BASE"}, // base first at all
				},
				{
					Contains: []string{"NB", "MC", "NC", "NA", "TC", "TA", "TB"}, //anything else at the end unordered
				},
				{
					Contains: []string{"MD"}, // the last is main_d
				},
			},
		}
	}

	assertRuntimeGroup(t, "./../../docs/test/01concurrent", "main_d", test)
}
