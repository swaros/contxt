package runner_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctemplate"
	"github.com/swaros/contxt/module/runner"
)

func TestCreateImport(t *testing.T) {
	template := ctemplate.New()
	// file must exists, so we add a path that exists for a other testing
	added, err := runner.AddPathToIncludeImports(template.GetIncludeConfig(), "testdata/task2/.contxt.yml")
	if err != nil {
		t.Error(err)
	}
	if added == "" {
		t.Error("Path should be added")
	}
	// this is how the output should look like
	expected := `include:
  basedir: false
  folders:
  - testdata/task2/.contxt.yml
`
	if added != expected {
		t.Errorf("Expected \n%s\ngot\n%s", expected, added)
	}

	// the configuration itself should be changed
	// because the template is a pointer
	if template.GetIncludeConfig().Include.Folders[0] != "testdata/task2/.contxt.yml" {
		t.Error("Configuration not changed")
	}

	// now we add another path
	added, err = runner.AddPathToIncludeImports(template.GetIncludeConfig(), "testdata/task3/.contxt.yml")
	if err != nil {
		t.Error(err)
	}
	if added == "" {
		t.Error("Path should be added")
	}
	// this is how the output should look like
	expected = `include:
  basedir: false
  folders:
  - testdata/task2/.contxt.yml
  - testdata/task3/.contxt.yml
`

	if added != expected {
		t.Errorf("Expected \n%s\ngot\n%s", expected, added)
	}

	// the configuration itself should be changed
	// because the template is a pointer
	if template.GetIncludeConfig().Include.Folders[0] != "testdata/task2/.contxt.yml" {
		t.Error("Configuration not changed")
	}
	if template.GetIncludeConfig().Include.Folders[1] != "testdata/task3/.contxt.yml" {
		t.Error("Configuration not changed")
	}

}
