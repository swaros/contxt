package yacl_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/swaros/contxt/module/yacl"
)

type configV1 struct {
	Name     string   `yaml:"name"`
	Path     string   `yaml:"path"`
	Boolflag bool     `yaml:"boolflag"`
	Subs     []string `yaml:"subs"`
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

/* working on this
func TestPropertieChanges(t *testing.T) {
	var cfg configV1
	cfgv1Handl := yacl.NewConfig(&cfg, yamc.NewYamlReader()).SetSubDirs("v1").SetSingleFile("cfgv1.yml")

	if err := cfgv1Handl.Load(); err != nil {
		t.Error(err)
	}

	if !cfg.Boolflag {
		t.Error("boolflag should be true")
	}

	if cfg.Name != "test" {
		t.Error("name should be test. got ", cfg.Name)
	}

}
*/
