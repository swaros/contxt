package cmdhandle_test

import (
	"os"
	"testing"

	"github.com/swaros/contxt/context/cmdhandle"
	"github.com/swaros/contxt/context/dirhandle"
)

func getTestTemplate(filename string) string {
	path := "./../../docs/test/" + filename
	return path
}

func TestGetPwdTemplate(t *testing.T) {

	path := getTestTemplate("testcase1.yml")

	template, terr := cmdhandle.GetPwdTemplate(path)
	if terr != nil {
		t.Error("could not get the template from path:", path, "\n error:", terr)
	} else {
		if template.Config.Sequencially == false {
			t.Error("expected template.Config.Sequencially is set to false")
		}
	}
}

func TestGetTemplate(t *testing.T) {
	old, derr := dirhandle.Current()
	if derr != nil {
		t.Error(derr)
	}
	os.Chdir("./../../docs/test/case0")
	template, _, exists, _ := cmdhandle.GetTemplate()
	if !exists {
		t.Error("could not found the test template file")
	} else {
		if template.Config.Sequencially == true {
			t.Error("expected template.Config.Sequencially is set to true")
		}
	}
	os.Chdir(old)
}

func TestGetVarImport(t *testing.T) {
	caseRunner("10", t, func(t *testing.T) {
		template, path, exists, _ := cmdhandle.GetTemplate()
		if !exists {
			t.Error("could not found the test template file", path)
		} else {
			if template.Config.Sequencially == false {
				t.Error("expected template.Config.Sequencially is set to false")
			}
			if template.Config.Coloroff == false {
				t.Error("expected template.Config.Coloroff is set to false")
			}
		}

		cmdhandle.RunTargets("script", true)
		test1Result := cmdhandle.GetPH("RUN.script.LOG.LAST")
		if test1Result == "" {
			t.Error("result 1 should not be empty.", test1Result)
		}

	})
}

// test to verfiy content that might be build in step 10
func TestGetVarImport11(t *testing.T) {
	caseRunner("11", t, func(t *testing.T) {
		template, path, exists, _ := cmdhandle.GetTemplate()
		if !exists {
			t.Error("could not found the test template file", path)
		} else {
			if template.Config.Sequencially == false {
				t.Error("expected template.Config.Sequencially is set to false")
			}
		}

	})
}
