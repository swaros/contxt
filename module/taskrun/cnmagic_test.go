package taskrun_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/taskrun"
)

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

func TestCnHandle(t *testing.T) {
	// setup the temp folder for the test
	rendomTimeBasedName := fmt.Sprintf("test-%d", time.Now().UnixNano())
	configure.USE_SPECIAL_DIR = false
	configure.CONTEXT_DIR = "temp"
	configure.CONTXT_FILE = rendomTimeBasedName + "fake_contxt.yml"
	configure.MIGRATION_ENABLED = false

	err := os.MkdirAll(configure.CONTEXT_DIR, 0777)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(configure.CONTEXT_DIR)

	// create some dirs
	createTestDirs := []string{"project1", "project2", "project3", "project4"}
	createDirsInProject := []string{"role1", "role2", "role3", "role4", "frontend/build", "backend/build", "toolset/assets/build"}

	for _, project := range createTestDirs {
		for _, role := range createDirsInProject {
			err = os.MkdirAll(configure.CONTEXT_DIR+"/"+project+"/"+role, 0777)
			if err != nil {
				t.Error(err)
			}
		}
	}

	// init the config
	conf := configure.NewContxtConfig()
	// add the projects
	for id, project := range createTestDirs {
		strId := strconv.Itoa(id)
		if cerr := conf.AddWorkSpace("project_"+strId, func(s string) bool { return true }, func(s string) {}); cerr != nil {
			t.Error(cerr)
		} else {
			// add the paths
			for _, role := range createDirsInProject {
				if err := conf.AddPath(configure.CONTEXT_DIR + "/" + project + "/" + role); err != nil {
					t.Error(err)
				}
			}
		}
	}
	if cerr := conf.SaveConfiguration(); cerr != nil {
		t.Error(cerr)
	}

	// change the workspace to project 0
	berr := conf.ChangeWorkspace("project_0", func(s1 string) bool {
		return true
	}, func(origin string) {
		if origin != "project_0" {
			t.Error("unexpected workspace", origin)
		}
	})
	if berr != nil {
		t.Error(berr)
	}

	// check the paths in a loop with the DirFind function and using a slice of strings
	// the first element is the path to find and the second is the expected result
	// the result is the path relative to the project
	// the first project is project1 (id -1)
	// the second project is project2 (id 0)
	// the third project is project3 (id 1)
	// the fourth project is project4 (id 2)

	sliceList := [][]string{
		{"le1", "temp/project1/role1"},
		{"le2", "temp/project1/role2"},
		{"le3", "temp/project1/role3"},
		{"le4", "temp/project1/role4"},
		{"frontend", "temp/project1/frontend/build"},
		{"backend", "temp/project1/backend/build"},
		{"toolset", "temp/project1/toolset/assets/build"},
	}

	for _, slice := range sliceList {
		check := taskrun.DirFind([]string{slice[0]})
		if check != slice[1] {
			t.Error("DirFind failed. got: " + check + " expected: " + slice[1])
		}
	}

	check := taskrun.DirFind([]string{"le2"})
	if check != "temp/project1/role2" { // the path for project 0 is project1 (id -1)
		t.Error("DirFind failed. got: " + check)
	}

	check = taskrun.DirFind([]string{"build"})
	if check != "temp/project1/toolset/assets/build" { // the path should used where the search word should stay at least in the path
		t.Error("DirFind failed. got: " + check)
	}

	check = taskrun.DirFind([]string{"back", "build"})
	if check != "temp/project1/backend/build" { // backend/build is the only one they matches to the search words
		t.Error("DirFind failed. got: " + check)
	}
}
