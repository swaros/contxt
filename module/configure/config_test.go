package configure_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yamc"
)

// prepareTempConfigFileFromBase is a helper function to use base.yml as config
// by copy base.yml to the temp folder and setup using them for the configuration
func prepareTempConfigFileFromBase(target string) error {
	if err := systools.CopyFile("test/base.yml", "test/temp/"+target); err != nil {
		return err
	}

	// redefine the targets
	configure.USE_SPECIAL_DIR = false
	configure.CONTEXT_DIR = "test/temp"
	configure.CONTXT_FILE = target
	return nil
}

func pathCompare(left, right string) bool {
	l := filepath.FromSlash(left)
	r := filepath.FromSlash(right)

	return l == r
}

func lazyHelperFindConfigEntry(t *testing.T, conMd *yacl.ConfigModel, expectContains string) bool {
	t.Helper()
	if err := helperFindConfigEntry(conMd, expectContains); err != nil {
		_, fnmane, lineNo, _ := runtime.Caller(1)
		t.Error(fnmane+":"+strconv.Itoa(lineNo), err)
		return false
	} else {
		return true
	}
}

func lazyHelperFindNotConfigEntry(t *testing.T, conMd *yacl.ConfigModel, expectContains string) bool {
	if err := helperFindConfigEntry(conMd, expectContains); err == nil {
		_, fnmane, lineNo, _ := runtime.Caller(1)
		t.Error(fnmane+":"+strconv.Itoa(lineNo), "found unexpected entry in config ", expectContains)
		return false
	} else {
		return true
	}
}

func helperFindConfigEntry(conMd *yacl.ConfigModel, expectContains string) error {
	if str, err := conMd.ToString(yamc.NewJsonReader()); err != nil {
		return err
	} else {
		if !strings.Contains(str, expectContains) {
			return errors.New("string[" + expectContains + "] is not in the content: " + str)
		}
	}
	return nil
}

func helperFindConfigEntryInFile(filename, expectContains string) (bool, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}
	str := string(file)
	if !strings.Contains(str, expectContains) {
		return false, errors.New("string[" + expectContains + "] not found in file content: " + str)
	}
	return true, nil
}

func lazyHelperFindFileContent(t *testing.T, filename string, content ...string) {
	for _, expect := range content {
		_, err := helperFindConfigEntryInFile(filename, expect)
		if err != nil {
			_, fnmane, lineNo, _ := runtime.Caller(1)
			t.Error(fnmane+":"+strconv.Itoa(lineNo), err)
			t.Error("do not found content ")
		}
	}
}

func lazyHelperFindFileHaveNoContent(t *testing.T, filename string, content ...string) {
	for _, expect := range content {
		_, err := helperFindConfigEntryInFile(filename, expect)
		if err == nil {
			_, fnmane, lineNo, _ := runtime.Caller(1)
			t.Error(fnmane+":"+strconv.Itoa(lineNo), "the file have the unexpected content "+expect)
		}
	}
}

// TestConfigLoadYaml testing basic read
func TestConfigLoadYaml(t *testing.T) {
	var cfg configure.ConfigMetaV2
	defaultV2Yacl := yacl.New(&cfg, yamc.NewYamlReader()).
		UseRelativeDir().
		SetSubDirs("test").
		SetSingleFile("case1.yml")

	if err := defaultV2Yacl.Load(); err != nil {
		t.Error("error while load ", err)
	}

	if cfg.CurrentSet != "contxt" {
		t.Error("current set should be contxt")
	}

	if len(cfg.Configs) != 5 {
		t.Error("invalid count of configurations ", len(cfg.Configs))
	}
}

// TestLoadWorkspaceData tests change current used path
func TestLoadWorkspaceData(t *testing.T) {
	// redefine the targets
	configure.USE_SPECIAL_DIR = false
	configure.CONTEXT_DIR = "test"
	configure.CONTXT_FILE = "case1.yml"

	conf := configure.NewContxtConfig()

	if !pathCompare(conf.DefaultV2Yacl.GetLoadedFile(), "test/case1.yml") {
		t.Error("load the wrong file. check test setup", conf.DefaultV2Yacl.GetLoadedFile())
	}

	hits := []string{}
	names := []string{}
	conf.ExecOnWorkSpaces(func(index string, cfg configure.ConfigurationV2) {
		hits = append(hits, index)
		names = append(names, cfg.Name)
	})

	for _, check := range []string{"lima", "manout", "version", "contxt", "fixed-line-mout"} {
		if !systools.SliceContains(hits, check) {
			t.Error("missing key ", check, " in the result")
		}
		if !systools.SliceContains(names, check) {
			t.Error("missing name ", check, " in the result")
		}
	}

	path := conf.GetActivePath(".")
	if !pathCompare(path, "/home/deep/development/markup-string") {
		t.Error("unexpected path [", path, "]")
	}

	if err := conf.ChangeActivePath("10"); err == nil {
		t.Error("path 10 should not be possible, because it is out of range")
	} else if !strings.Contains(err.Error(), "could not change the index.") {
		t.Error("wrong error reported:", err)
	}

	if err := conf.ChangeActivePath("0"); err != nil {
		t.Error("this should work")
	} else {
		path := conf.GetActivePath(".")
		if !pathCompare(path, "/home/deep/development/contxt") {
			t.Error("unexpected path [", path, "]")
		}
	}
}

func TestChangeWorksSpace(t *testing.T) {

	if err := prepareTempConfigFileFromBase("case002.yml"); err != nil {
		t.Error(err)
	}
	conf := configure.NewContxtConfig()
	if !pathCompare(conf.DefaultV2Yacl.GetLoadedFile(), "test/temp/case002.yml") {
		t.Error("load the wrong file. check test setup", conf.DefaultV2Yacl.GetLoadedFile())
	}

	if conf.UsedV2Config.CurrentSet != "contxt" {
		t.Error("the wrong workspace is set ", conf.UsedV2Config.CurrentSet)
	}

	wsErr := conf.ChangeWorkspace("lima", func(s string) bool { return true }, func(s string) {})
	if wsErr != nil {
		t.Error(wsErr)
	}

	if conf.UsedV2Config.CurrentSet != "lima" {
		t.Error("the wrong workspace is not set to the new one  ", conf.UsedV2Config.CurrentSet)
	}

	if configure.CFG.CurrentSet != "lima" {
		t.Error("main config did also not change to lima", configure.CFG.CurrentSet)
	}

	if str, err := conf.DefaultV2Yacl.ToString(yamc.NewJsonReader()); err != nil {
		t.Error(err)
	} else {
		if !strings.Contains(str, `"CurrentSet":"lima"}`) {
			t.Error(" no match ", str)
		}
	}

	// testing relation between the the config-handler and the current config model
	// if the relation is lost, any save will write outdated content
	if y, err := conf.DefaultV2Yacl.CreateYamc(yamc.NewJsonReader()); err != nil {
		t.Error("error on convert data")
	} else {
		if v, err := y.FindValue("CurrentSet"); err != nil {
			t.Error("error while trying to get value", err)
		} else {
			if v.(string) != "lima" {
				t.Error("unexpected content of the value ", v)
			}
		}
	}

	if err := conf.AddPath("/home/deep/development/lima"); err == nil {
		t.Error("this path already exists, so this should be an error")
	} else {
		if !strings.Contains(err.Error(), "the path is already stored in lima") {
			t.Error("unexpected error message ", err)
		}
	}

	if err := conf.AddPath("/home/deep/development/test-lima"); err != nil {
		t.Error("this should work")
	} else {
		if err := helperFindConfigEntry(conf.DefaultV2Yacl, `/home/deep/development/test-lima`); err != nil {
			t.Error(err)
		} else {

			found, _ := helperFindConfigEntryInFile(conf.DefaultV2Yacl.GetLoadedFile(), "/home/deep/development/test-lima")
			if found {
				t.Error("the path should not already in the config, because the function do not save by his own")
			}

			conf.SaveConfiguration()
			found, _ = helperFindConfigEntryInFile(conf.DefaultV2Yacl.GetLoadedFile(), "/home/deep/development/test-lima")
			if !found {
				t.Error("now the path should exists in the config file")
			}
		}

	}

	if !conf.RemovePath("/home/deep/development/test-lima") {
		t.Error("error on removing path from config")
	} else {
		lazyHelperFindConfigEntry(t, conf.DefaultV2Yacl, "/home/deep/development/lima")         // should still be there
		lazyHelperFindNotConfigEntry(t, conf.DefaultV2Yacl, "/home/deep/development/test-lima") // but this should no longer be there

		// but it should not save yet in the config file
		lazyHelperFindFileContent(t, conf.DefaultV2Yacl.GetLoadedFile(), "/home/deep/development/lima", "/home/deep/development/test-lima")

		// now save it
		conf.SaveConfiguration()
		// now lets check the config file again
		lazyHelperFindFileContent(t, conf.DefaultV2Yacl.GetLoadedFile(), "/home/deep/development/lima")            // should be still there
		lazyHelperFindFileHaveNoContent(t, conf.DefaultV2Yacl.GetLoadedFile(), "/home/deep/development/test-lima") // but this should be gone

	}

	conf.ClearPaths() // removing all paths

	lazyHelperFindNotConfigEntry(t, conf.DefaultV2Yacl, "/home/deep/development/lima") // now this path should also be gone

	lazyHelperFindFileContent(t, conf.DefaultV2Yacl.GetLoadedFile(), "/home/deep/development/contxt") // just check if we touch nothing else
	// all paths should be gone
	lazyHelperFindFileHaveNoContent(t, conf.DefaultV2Yacl.GetLoadedFile(), "/home/deep/development/test-lima", "/home/deep/development/lima")

	// now try removing the workspace. this should not work because we are currently in this workspace
	if err := conf.RemoveWorkspace("lima"); err == nil {
		t.Error("removing the current workspace should not work")
	} else {
		// first we switch to an different workspace. this should work
		if err := conf.ChangeWorkspace("manout", func(s string) bool { return true }, func(s string) {}); err != nil {
			t.Error(err)
		} else {
			// now the removal should work
			if err := conf.RemoveWorkspace("lima"); err != nil {
				t.Error("removing the lima workspace should work now")
			} else {
				if conf.HaveWorkSpace("lima") {
					t.Error("lima should no longer exists")
				}
				// check if lima is in the config. should not be the case
				lazyHelperFindNotConfigEntry(t, conf.DefaultV2Yacl, "lima:")
				// save the config and check again
				conf.SaveConfiguration()
				// lima should also removed from config file
				lazyHelperFindFileHaveNoContent(t, conf.DefaultV2Yacl.GetLoadedFile(), "lima:")

			}
		}
	}

	// ---- testing adding a new workspace ------

	// did we handle invalid chars?
	if err := conf.AddWorkSpace("this shouldnot valid {} :", func(s string) bool { return true }, func(s string) {}); err == nil {
		t.Error("this should not work because of the weird naming")
	} else {
		if !strings.Contains(err.Error(), "string contains not accepted chars") {
			t.Error("unexpected error message:", err)
		}
	}

	// some other chars, that will be translated, should also not accepted
	if err := conf.AddWorkSpace("slashes/also/not/allowed", func(s string) bool { return true }, func(s string) {}); err == nil {
		t.Error("this should not work because of the weird naming")
	} else {
		if !strings.Contains(err.Error(), "is invalid") {
			t.Error("unexpected error message:", err)
		}
	}

	if err := conf.AddWorkSpace("test-space", func(s string) bool { return true }, func(s string) {}); err != nil {
		t.Error(err)
	} else {
		// check config file and config entries
		lazyHelperFindFileContent(t, conf.DefaultV2Yacl.GetLoadedFile(), "test-space:", "CurrentSet: test-space")
		lazyHelperFindConfigEntry(t, conf.DefaultV2Yacl, `"test-space":{"CurrentIndex":"","Name":"","Paths":{}}`)

		// add a path
		if err := conf.AddPath("/home/deep/development/pathNo1"); err != nil {
			t.Error(err)
		} else {
			conf.SaveConfiguration()
			lazyHelperFindConfigEntry(t, conf.DefaultV2Yacl,
				`"test-space":{"CurrentIndex":"0","Name":"","Paths":{"0":{"Path":"/home/deep/development/pathNo1","Project":"","Role":"","Version":""}}}`)
		}
	}
}
