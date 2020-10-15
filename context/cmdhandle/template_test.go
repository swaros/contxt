package cmdhandle_test

import (
	"os"
	"testing"

	"github.com/swaros/contxt/context/cmdhandle"
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
	os.Chdir("./../../docs/test/")
	template, exists := cmdhandle.GetTemplate()
	if !exists {
		t.Error("could not found the test template file")
	} else {
		if template.Config.Sequencially == true {
			t.Error("expected template.Config.Sequencially is set to true")
		}
	}
}
