package taskrun_test

import (
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/taskrun"
)

func TestCnFind(t *testing.T) {
	configure.UsedConfig.Paths = []string{
		"/home/user/project/testing",
		"/home/user/project/testing/source/backend/server",
		"/home/user/project/testing/source/backend/server/build",
		"/home/user/project/testing/source/frontend/website",
		"/home/user/project/testing/source/frontend/website/build",
		"/home/user/project/testing/source/toolset/someelse/build",
	}

	assertStringEquals(t, taskrun.DirFind([]string{"0"}), "/home/user/project/testing")
	assertStringEquals(t, taskrun.DirFind([]string{"1"}), "/home/user/project/testing/source/backend/server")
	assertStringEquals(t, taskrun.DirFind([]string{"web", "build"}), "/home/user/project/testing/source/frontend/website/build")
	assertStringEquals(t, taskrun.DirFind([]string{"serv", "build"}), "/home/user/project/testing/source/backend/server/build")
}

func TestDecicePath(t *testing.T) {
	paths := []string{
		"/home/user/project/testing",
		"/home/user/project/testing/source/backend/server",
		"/home/user/project/testing/source/backend/server/build",
		"/home/user/project/testing/source/frontend/website",
		"/home/user/project/testing/source/frontend/website/build",
		"/home/user/project/testing/source/toolset/someelse/build",
	}

	pathsReordered := []string{
		"/home/user/project/testing/source/toolset/someelse/build",
		"/home/user/project/testing/source/frontend/website/build",
		"/home/user/project/testing/source/backend/server",
		"/home/user/project/testing/source/frontend/website",
		"/home/user/project/testing",
		"/home/user/project/testing/source/backend/server/build",
	}
	if decidedPath, ok := taskrun.DecidePath([]string{"website"}, paths); ok {
		assertStringEquals(t, decidedPath, "/home/user/project/testing/source/frontend/website")
	} else {
		t.Error("they shoud find a path.")
	}

	if decidedPath, ok := taskrun.DecidePath([]string{"website"}, pathsReordered); ok {
		assertStringEquals(t, decidedPath, "/home/user/project/testing/source/frontend/website")
	} else {
		t.Error("they shoud find a path.")
	}

	if decidedPath, ok := taskrun.DecidePath([]string{"server"}, paths); ok {
		assertStringEquals(t, decidedPath, "/home/user/project/testing/source/backend/server")
	} else {
		t.Error("they shoud find a path.")
	}

	if decidedPath, ok := taskrun.DecidePath([]string{"server"}, pathsReordered); ok {
		assertStringEquals(t, decidedPath, "/home/user/project/testing/source/backend/server")
	} else {
		t.Error("they shoud find a path.")
	}
	// looking for build in booth path slices
	if decidedPath, ok := taskrun.DecidePath([]string{"build"}, paths); ok {
		assertStringEquals(t, decidedPath, "/home/user/project/testing/source/frontend/website/build")
	} else {
		t.Error("they shoud find a path.")
	}

	if decidedPath, ok := taskrun.DecidePath([]string{"build"}, pathsReordered); ok {
		assertStringEquals(t, decidedPath, "/home/user/project/testing/source/frontend/website/build")
	} else {
		t.Error("they shoud find a path.")
	}

	// looking for a specific path by more arguments
	if decidedPath, ok := taskrun.DecidePath([]string{"web", "build"}, paths); ok {
		assertStringEquals(t, decidedPath, "/home/user/project/testing/source/frontend/website/build")
	} else {
		t.Error("they shoud find a path.")
	}

	if decidedPath, ok := taskrun.DecidePath([]string{"server", "build"}, paths); ok {
		assertStringEquals(t, decidedPath, "/home/user/project/testing/source/backend/server/build")
	} else {
		t.Error("they shoud find a path.")
	}

	if decidedPath, ok := taskrun.DecidePath([]string{"build", "server"}, paths); ok {
		assertStringEquals(t, decidedPath, "/home/user/project/testing/source/backend/server/build")
	} else {
		t.Error("they shoud find a path.")
	}

	if decidedPath, ok := taskrun.DecidePath([]string{"build", "tool"}, paths); ok {
		assertStringEquals(t, decidedPath, "/home/user/project/testing/source/toolset/someelse/build")
	} else {
		t.Error("they shoud find a path.")
	}

	if decidedPath, ok := taskrun.DecidePath([]string{"noin", "eventhis"}, paths); ok {
		t.Error("is is impossible to find something that matches with the arguments")
	} else {
		assertStringEquals(t, decidedPath, ".")
	}
}
