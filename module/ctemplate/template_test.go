package ctemplate_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaros/contxt/module/ctemplate"
)

var (
	pushdir = "."
)

func pushDir(todir string) {
	if dir, err := os.Getwd(); err == nil {
		pushdir = dir
	}
	os.Chdir(todir)
}

func popDir() {
	os.Chdir(pushdir)
}

func TestTemplate(t *testing.T) {
	pushDir("testdata/basic")
	tmplte := ctemplate.New()
	if err := tmplte.Init(); err != nil {
		t.Error(err)
	}

	if _, ok := tmplte.FindTemplateFileName(); !ok {
		t.Error("Template file not found")
	} else {

		if ctxTmpl, err := tmplte.LoadV2(); err != nil {
			t.Error("Template not loaded", err)
		} else {
			assert.Equal(t, ctxTmpl.Task[0].ID, "task1")
		}
	}

	popDir()
}

func TestTemplateInclude(t *testing.T) {
	pushDir("testdata/withInclude")
	tmplte := ctemplate.New()
	if err := tmplte.Init(); err != nil {
		t.Error(err)
	}

	cfg, _, err := tmplte.Load()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, cfg.Task[0].ID, "mars")

	popDir()
}

func TestTemplateV2Include(t *testing.T) {
	pushDir("testdata/withInclude")
	tmplte := ctemplate.New()
	if err := tmplte.Init(); err != nil {
		t.Error(err)
	}

	cfg, err := tmplte.LoadV2()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, cfg.Task[0].ID, "mars")

	popDir()
}

func TestTemplateWithLinter(t *testing.T) {
	pushDir("testdata/basic")
	tmplte := ctemplate.New()
	if err := tmplte.Init(); err != nil {
		t.Error(err)
	}

	if _, ok := tmplte.FindTemplateFileName(); !ok {
		t.Error("Template file not found")
	} else {
		tmplte.SetLinting(true) // enable linting
		if ctxTmpl, err := tmplte.LoadV2(); err != nil {
			t.Error("Template not loaded", err)
		} else {
			assert.Equal(t, ctxTmpl.Task[0].ID, "task1")
			assert.Equal(t, "echo hello", ctxTmpl.Task[0].Script[0], "Task 1")
		}

		if linter, err := tmplte.GetLinter(); err != nil {
			t.Error(err)
		} else {
			if linter.HasError() {
				t.Error("Linter should not have errors")
				t.Log("\n", strings.Join(linter.Errors(), "\n"))
			}
		}
	}

	popDir()
}

func TestTemplateV2KeydMap(t *testing.T) {
	pushDir("testdata/valueMaps")
	tmplte := ctemplate.New()
	if err := tmplte.Init(); err != nil {
		t.Error(err)
		return
	}

	cfg, err := tmplte.LoadV2()
	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "mars", cfg.Task[0].ID)
	assert.Equal(t, "echo valueA + valueB + valueC", cfg.Task[0].Script[0])
	assert.Equal(t, "echo postgreSQL", cfg.Task[0].Script[1])
	assert.Equal(t, "echo valueB", cfg.Task[0].Script[2])

	popDir()
}

func TestWithIgnoredValues(t *testing.T) {
	pushDir("testdata/ignore")
	tmplte := ctemplate.New()
	if err := tmplte.Init(); err != nil {
		t.Error(err)
	}
	tmplte.SetIgnoreHndl(true)
	cfg, err := tmplte.LoadV2()
	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "testing", cfg.Task[0].ID)
	assert.Equal(t, `echo "docker inspect \-f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' container_name_or_id"`, cfg.Task[0].Script[0])

	popDir()
}
