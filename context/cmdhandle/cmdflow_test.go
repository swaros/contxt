package cmdhandle_test

import (
	"os"
	"testing"

	"github.com/swaros/contxt/context/dirhandle"

	"github.com/swaros/contxt/context/cmdhandle"
)

func TestRunTarget(t *testing.T) {
	cmdhandle.ClearAll()
	old, derr := dirhandle.Current()
	if derr != nil {
		t.Error(derr)
	}
	os.Chdir("./../../docs/test/")
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
	os.Chdir(old)
}

func TestRunTargetCase1(t *testing.T) {
	cmdhandle.ClearAll()
	old, derr := dirhandle.Current()
	if derr != nil {
		t.Error(derr)
	}
	os.Chdir("./../../docs/test/case1")
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
	os.Chdir(old)
}
