package yacl_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yamc"
)

type configV0 struct {
}

type configV1 struct {
	Name     string   `yaml:"name"`
	Path     string   `yaml:"path"`
	Boolflag bool     `yaml:"boolflag"`
	Subs     []string `yaml:"subs"`
}

type configV2 struct {
	Doe          string   `yaml:"doe"`
	Ray          string   `yaml:"ray"`
	Pi           float64  `yaml:"pi"`
	Xmas         bool     `yaml:"xmas"`
	FrenchHens   int      `yaml:"french-hens"`
	CallingBirds []string `yaml:"calling-birds"`
}

func TestComposePath(t *testing.T) {
	var cfg *configV0

	// relative path just with two subfolders
	cfgv1 := yacl.NewConfig(cfg).SetSubDirs("test", "version")
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

func TestPropertieLoads(t *testing.T) {
	var cfg configV2
	cfgv1Handl := yacl.NewConfig(&cfg, yamc.NewYamlReader()).SetSubDirs("v1").SetSingleFile("cfgv2.yml")

	if err := cfgv1Handl.Load(); err != nil {
		t.Error(err)
	}

	if cfgv1Handl.GetLoadedFile() != filepath.Clean("v1/cfgv2.yml") {
		t.Error("wrong file loaded", cfgv1Handl.GetLoadedFile())
	}

	if !cfg.Xmas {
		t.Error("boolflag should be true")
	}

	if cfg.Doe != "a deer, a female deer" {
		t.Error("name should be test. got ", cfg.Doe)
	}

	if cfg.Pi != 3.14159 {
		t.Error("Pi is wrong")
	}

	if cfg.FrenchHens != 3 {
		t.Error("french-hens is wrong")
	}

	list := cfg.CallingBirds
	if len(list) != 4 {
		t.Error("wrong count of list entries")
	}
}

func TestPropertieFailLoads(t *testing.T) {
	var cfg configV2
	cfgv1Handl := yacl.NewConfig(&cfg, yamc.NewYamlReader()).SetSubDirs("v1").SetSingleFile("notThere.yml")

	if err := cfgv1Handl.Load(); err == nil {
		t.Error("the file not exists. this should result in a error")
	} else {
		if err.Error() != "at least one Configuration should exists. but found nothing" {
			t.Error("we expected a different error message then ", err.Error())
		}
	}
}

func TestPropertieFailLoadsBecauseOfDirectoryNotExists(t *testing.T) {
	var cfg configV2
	cfgv1Handl := yacl.NewConfig(&cfg, yamc.NewYamlReader()).SetSubDirs("someDir").SetSingleFile("cfgv2.yml")

	if err := cfgv1Handl.Load(); err == nil {
		t.Error("the file not exists. this should result in a error")
	} else {
		if !strings.Contains(err.Error(), "the path someDir not exists") {
			t.Error("we expected a different error message then ", err.Error())
		}
	}
}

func TestPropertieChanges(t *testing.T) {
	var cfg configV1 = configV1{
		Name:     "Unset",
		Boolflag: false,
		Path:     "none",
	}
	cfgv1Handl := yacl.NewConfig(&cfg, yamc.NewYamlReader()).SetSubDirs("v1").SetSingleFile("cfgv1.yml")

	if err := cfgv1Handl.Load(); err != nil {
		t.Error(err)
	}

	if cfgv1Handl.GetLoadedFile() != filepath.Clean("v1/cfgv1.yml") {
		t.Error("wrong file loaded", cfgv1Handl.GetLoadedFile())
	}

	if !cfg.Boolflag {
		t.Error("boolflag should be true")
	}

	if cfg.Name != "test" {
		t.Error("name should be test. got ", cfg.Name)
	}

	// now change the properties ones
	cfg.Path = "helloWorldPath"
	cfg.Name = "new Day"

	if newStr, err := cfgv1Handl.ToString(yamc.NewJsonReader()); err != nil {
		t.Error(err)
	} else {
		checkStr := `{"Boolflag":true,"Name":"new Day","Path":"helloWorldPath","Subs":["first","second"]}`
		if newStr != checkStr {
			t.Error("wrong created string[", newStr, "]")
		}
	}

	// now change the properties ones
	cfg.Path = "an other name"
	cfg.Name = "changed again"
	cfg.Boolflag = false

	if newStr, err := cfgv1Handl.ToString(yamc.NewJsonReader()); err != nil {
		t.Error(err)
	} else {
		checkStr := `{"Boolflag":false,"Name":"changed again","Path":"an other name","Subs":["first","second"]}`
		if newStr != checkStr {
			t.Error("wrong created string[", newStr, "]")
		}
	}
}
