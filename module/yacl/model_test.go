package yacl_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/swaros/contxt/module/yacl"
)

type configV1 struct {
	/*
		name     string
		path     string
		boolflag bool
		subs     []string
	*/
}

func TestComposePath(t *testing.T) {
	var cfg configV1

	// relative path just with two subfolders
	cfgv1 := yacl.NewConfig(&cfg).SetSubDirs("test", "version")
	path := cfgv1.GetConfigPath()
	if path != filepath.Clean("test/version") {
		t.Error("did not get expected path ", path)
	}

	// recreate by using the home dir
	cfgv1 = yacl.NewConfig(&cfg).UseHomeDir().SetSubDirs("test", "version")
	path = cfgv1.GetConfigPath()

	if usrHome, err := os.UserHomeDir(); err != nil {
		t.Error(err)
	} else {
		if path != filepath.Clean(usrHome+"/test/version") {
			t.Error("did not get expected path ", path)
		}

	}

	// recreate by using the user config dir
	cfgv1 = yacl.NewConfig(&cfg).UseConfigDir().SetSubDirs("test", "version")
	path = cfgv1.GetConfigPath()

	if usrCfgHome, err := os.UserConfigDir(); err != nil {
		t.Error(err)
	} else {
		if path != filepath.Clean(usrCfgHome+"/test/version") {
			t.Error("did not get expected path ", path, " expected ", filepath.Clean(usrCfgHome+"/test/version"))
		}

	}

}
