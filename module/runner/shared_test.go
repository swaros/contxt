package runner_test

import (
	"os"
	"testing"

	"github.com/swaros/contxt/module/runner"
)

func TestGetPath(t *testing.T) {
	shared := runner.NewSharedHelperWithPath("/home/user")
	if shared.GetSharedPath("test") != "/home/user/.contxt/shared/test" {
		t.Error("unexpected path")
	}
}

func TestGetPathWithSub(t *testing.T) {
	shared := runner.NewSharedHelperWithPath("/home/user")
	if shared.GetBasePath() != "/home/user" {
		t.Error("unexpected path", shared.GetBasePath())
	}

}

func TestCheckOrCreateUseConfig(t *testing.T) {
	if path, err := os.MkdirTemp("", "testCase*"); err != nil {
		t.Error(err)
	} else {
		defer os.RemoveAll(path)
		shared := runner.NewSharedHelperWithPath(path)
		if libPaqth, gitError := shared.CheckOrCreateUseConfig("swaros/ctx-git"); gitError != nil {
			t.Error(gitError)
		} else {
			expectedPath := path + "/.contxt/shared/swaros/ctx-git/source" // like /tmp/testCase4133459873/.contxt/shared/swaros/ctx-git/source
			if libPaqth != expectedPath {
				t.Error("unexpected path", libPaqth, "expected:", expectedPath)
			}
		}
	}
}

func TestCheckOrCreateUseConfigNotExists(t *testing.T) {
	t.Skip("not working on github without prompt for password")
	if path, err := os.MkdirTemp("", "testCase*"); err != nil {
		t.Error(err)
	} else {
		defer os.RemoveAll(path)
		shared := runner.NewSharedHelperWithPath(path)
		if libPaqth, gitError := shared.CheckOrCreateUseConfig("swaros/ctx-notThere"); gitError == nil {
			if libPaqth != "" {
				t.Error("should return an empty path")
			}
			t.Error("should return an error, because the repo does not exists")
		}
	}
}
