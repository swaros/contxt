package taskrun_test

import (
	"testing"

	"github.com/swaros/contxt/module/taskrun"
)

func TestGetUseCaseMain(t *testing.T) {
	testpath := "./../../docs/test/02shared"

	usecase, version := taskrun.GetUseInfo("swaros/ctx-git", testpath)
	if usecase != "swaros/ctx-git" {
		t.Error("unexpected usecase:", usecase)
	}

	if version != "refs/heads/main" {
		t.Error("unexpected version:", version)
	}

}

func TestGetUseCaseVersion(t *testing.T) {
	testpath := "./../../docs/test/02shared"

	usecase, version := taskrun.GetUseInfo("swaros/ctx-git@v0.0.1", testpath)
	if usecase != "swaros/ctx-git" {
		t.Error("unexpected usecase:", usecase)
	}

	if version != "refs/tags/v0.0.1" {
		t.Error("unexpected version:", version)
	}

}
