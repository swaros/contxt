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

func TestLoadWithoutLoaders(t *testing.T) {
	var cfg *configV0
	cfgv1 := yacl.New(cfg).SetSubDirs("test", "version")
	err := cfgv1.Load()
	if err != nil {
		if !strings.Contains(err.Error(), "no loaders assigned") {
			t.Error("error do not contains expected part of the error message ", err.Error())
		}
	} else {
		t.Error("an error should happen, because we do not used any reader")
	}

	// same for save
	err = cfgv1.Save()
	if err != nil {
		if !strings.Contains(err.Error(), "we need at least one") {
			t.Error("error do not contains expected part of the error message ", err.Error())
		}
	} else {
		t.Error("an error should happen, because we do not used any reader")
	}
}

func TestComposePath(t *testing.T) {
	var cfg *configV0

	// relative path just with two subfolders
	cfgv1 := yacl.New(cfg).SetSubDirs("test", "version")
	path := cfgv1.GetConfigPath()
	if path != filepath.Clean("test/version") {
		t.Error("did not get expected path ", path)
	}

	// recreate by using the home dir
	cfgv1 = yacl.New(&cfg).UseHomeDir().SetSubDirs("test", "version")
	path = cfgv1.GetConfigPath()

	if usrHome, err := os.UserHomeDir(); err != nil {
		t.Error(err)
	} else {
		if path != filepath.Clean(usrHome+"/test/version") {
			t.Error("did not get expected path ", path)
		}

	}

	// recreate by using the user config dir
	cfgv1 = yacl.New(&cfg).UseConfigDir().SetSubDirs("test", "version")
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
	cfgv1Handl := yacl.New(&cfg, yamc.NewYamlReader()).SetSubDirs("v1").SetSingleFile("cfgv2.yml")

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
	cfgv1Handl := yacl.New(&cfg, yamc.NewYamlReader()).
		SetSubDirs("v1").
		SetSingleFile("notThere.yml")

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
	cfgv1Handl := yacl.New(&cfg, yamc.NewYamlReader()).
		SetSubDirs("someDir").
		SetSingleFile("cfgv2.yml")

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
	cfgv1Handl := yacl.New(&cfg, yamc.NewYamlReader()).
		SetSubDirs("v1").
		SetSingleFile("cfgv1.yml")

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

// testing loading behavior

type chainNode struct {
	Loglevel string            `yaml:"loglevel"`
	Env      map[string]string `yaml:"env"`
	Users    []string          `yaml:"users"`
	Host     string            `yaml:"host"`
	InPort   int               `yaml:"inport"`
	OutPort  int               `yaml:"outport"`
}

type chainConfig struct {
	Config chainNode `yaml:"config"`
}

func TestLoadingOverride(t *testing.T) {
	var cfg chainConfig
	chainCfg := yacl.New(&cfg, yamc.NewYamlReader()).
		SetSubDirs("v2")
		//SetFolderBlackList([]string{"v2/deployEu", "v2/deployUs"})

	loadErr := chainCfg.Load()
	if loadErr != nil {
		t.Error(loadErr)
		return // no need to test anything else if this was failing already
	}

	filesAll := strings.Join(chainCfg.GetAllParsedFiles(), ", ")
	if filesAll != "v2/001-test.base.yml, v2/002-local-gitignored.yml" {
		t.Error("error on loading the expected files ", filesAll)
	}

	if cfg.Config.Loglevel != "DEBUG" {
		t.Error("got unexpected ", cfg.Config.Loglevel)
	}

	if cfg.Config.InPort != 8089 {
		t.Error("unset properties should stay overwritten ", cfg.Config.InPort)
	}

	users := strings.Join(cfg.Config.Users, ", ")
	if users != "$USER, FakeHost" {
		t.Error("got unexpected user list ", users)
	}

}

// in this test the local is not excluded. in real worls
// this file would not being deployd because config files like them
// should be gitgnored
func TestLoadingOverrideAnUseEu(t *testing.T) {
	var cfg chainConfig
	chainCfg := yacl.New(&cfg, yamc.NewYamlReader()).
		SetSubDirs("v2").
		AllowSubdirsByRegex("deployEu")

	loadErr := chainCfg.Load()
	if loadErr != nil {
		t.Error(loadErr)
		return // no need to test anything else if this was failing already
	}

	filesAll := strings.Join(chainCfg.GetAllParsedFiles(), ", ")
	if filesAll != "v2/001-test.base.yml, v2/002-local-gitignored.yml, v2/deployEu/001-test.base.yml" {
		t.Error("error on loading the expected files ", filesAll)
	}

	if cfg.Config.Loglevel != "DEBUG" {
		t.Error("got unexpected ", cfg.Config.Loglevel)
	}

	if cfg.Config.InPort != 8089 {
		t.Error("unset properties should stay overwritten ", cfg.Config.InPort)
	}

	users := strings.Join(cfg.Config.Users, ", ")
	if users != "$USER, FakeHost" {
		t.Error("got unexpected user list ", users)
	}

	if cfg.Config.Host != "europe.deploy.de" {
		t.Error("got unexpected host", cfg.Config.Host)
	}

}

func TestLoadingOverrideAndUseEuByBlackList(t *testing.T) {
	var cfg chainConfig
	chainCfg := yacl.New(&cfg, yamc.NewYamlReader()).
		SetSubDirs("v2").
		SetFolderBlackList([]string{"v2/deployUs"})

	loadErr := chainCfg.Load()
	if loadErr != nil {
		t.Error(loadErr)
		return // no need to test anything else if this was failing already
	}

	filesAll := strings.Join(chainCfg.GetAllParsedFiles(), ", ")
	if filesAll != "v2/001-test.base.yml, v2/002-local-gitignored.yml, v2/deployEu/001-test.base.yml" {
		t.Error("error on loading the expected files ", filesAll)
	}

	if cfg.Config.Loglevel != "DEBUG" {
		t.Error("got unexpected ", cfg.Config.Loglevel)
	}

	if cfg.Config.InPort != 8089 {
		t.Error("unset properties should stay overwritten ", cfg.Config.InPort)
	}

	users := strings.Join(cfg.Config.Users, ", ")
	if users != "$USER, FakeHost" {
		t.Error("got unexpected user list ", users)
	}

	if cfg.Config.Host != "europe.deploy.de" {
		t.Error("got unexpected host", cfg.Config.Host)
	}

}

// explizit ignoring the local test config by file pattern
func TestLoadingOverrideUseUsNoDEv(t *testing.T) {
	var cfg chainConfig
	chainCfg := yacl.New(&cfg, yamc.NewYamlReader()).
		SetSubDirs("v2").
		AllowSubdirsByRegex("deployUs").
		SetFilePattern("00([0-9])-(....).base.yml") // sets regex so it ignores the local

	loadErr := chainCfg.Load()
	if loadErr != nil {
		t.Error(loadErr)
		return // no need to test anything else if this was failing already
	}

	filesAll := strings.Join(chainCfg.GetAllParsedFiles(), ", ")
	if filesAll != "v2/001-test.base.yml, v2/deployUs/001-test.base.yml" {
		t.Error("error on loading the expected files ", filesAll)
	}

	if cfg.Config.Loglevel != "ERROR" {
		t.Error("got unexpected ", cfg.Config.Loglevel)
	}

	if cfg.Config.InPort != 5001 {
		t.Error("unset properties should stay overwritten ", cfg.Config.InPort)
	}

	users := strings.Join(cfg.Config.Users, ", ")
	if users != "root" {
		t.Error("got unexpected user list ", users)
	}

	if cfg.Config.Host != "us-east.deploy.com" {
		t.Error("got unexpected host", cfg.Config.Host)
	}

}

// here we loads any related config for an us deploy one
// by one.
// first we load the default setup, and afterwards the overwrite
// by us.
// here we avoid any folder magic.
func TestLoadingOverrideUseUsByChainLoad(t *testing.T) {
	var cfg chainConfig
	chainCfg := yacl.New(&cfg, yamc.NewYamlReader())
	chainCfg.LoadFile("v2/001-test.base.yml")
	chainCfg.LoadFile("v2/deployUs/001-test.base.yml")

	loadErr := chainCfg.Load()
	if loadErr != nil {
		t.Error(loadErr)
		return // no need to test anything else if this was failing already
	}

	filesAll := strings.Join(chainCfg.GetAllParsedFiles(), ", ")
	if filesAll != "v2/001-test.base.yml, v2/deployUs/001-test.base.yml" {
		t.Error("error on loading the expected files ", filesAll)
	}

	if cfg.Config.Loglevel != "ERROR" {
		t.Error("got unexpected ", cfg.Config.Loglevel)
	}

	if cfg.Config.InPort != 5001 {
		t.Error("unset properties should stay overwritten ", cfg.Config.InPort)
	}

	users := strings.Join(cfg.Config.Users, ", ")
	if users != "root" {
		t.Error("got unexpected user list ", users)
	}

	if cfg.Config.Host != "us-east.deploy.com" {
		t.Error("got unexpected host", cfg.Config.Host)
	}

}

func TestNoConfigFilesFailExpected(t *testing.T) {
	var cfg chainConfig
	chainCfg := yacl.New(&cfg, yamc.NewYamlReader()).
		SetSubDirs("v10")

	loadErr := chainCfg.Load()
	if loadErr == nil {
		t.Error("load without ignore nof files exists, should be end up in an error")
	}
}

func TestNoConfigFiles(t *testing.T) {
	var cfg chainConfig
	chainCfg := yacl.New(&cfg, yamc.NewYamlReader()).
		SetSubDirs("v2").
		SetSingleFile("not_existing.yml").
		SetExpectNoConfigFiles() // it is okay for not having any configuration yet.

	loadErr := chainCfg.Load()
	if loadErr != nil {
		t.Error(loadErr)
	}
}

// if we expect files, we can still handle not existing
// configs by using the Init handler. if we return nil there,
// we do not run in an error
func TestNoConfigFilesCheckInit(t *testing.T) {
	initIsCalled := false
	notExistCalled := false
	var cfg chainConfig
	chainCfg := yacl.New(&cfg, yamc.NewYamlReader()).
		SetSubDirs("v2").
		SetSingleFile("not_existing.yml").
		Init(nil, func(errCode int) error {
			if errCode == yacl.NO_CONFIG_FILES_FOUND {
				initIsCalled = true
			}
			if errCode == yacl.ERROR_PATH_NOT_EXISTS {
				notExistCalled = true
			}
			return nil
		})

	loadErr := chainCfg.Load()
	if loadErr != nil {
		t.Error(loadErr)
	}

	if chainCfg.GetLoadedFile() != "" {
		t.Error("whoops. why is this file used? ", chainCfg.GetLoadedFile())
	}

	if notExistCalled {
		t.Error("the dir v2 should exists (please check) so this error should not happen")
	}
	if !initIsCalled {
		t.Error("there is no matching file. we should get a Init call with NO_CONFIG_FILES_FOUND")
	}
}

func TestNoConfigFilesByMissingFolder(t *testing.T) {
	var cfg chainConfig
	chainCfg := yacl.New(&cfg, yamc.NewYamlReader()).
		SetSubDirs("v10").
		SetExpectNoConfigFiles() // it is okay for not having any configuration yet.

	loadErr := chainCfg.Load()
	if loadErr != nil {
		if !strings.Contains(loadErr.Error(), "the path v10 not exists.") {
			t.Error("enexpected error message ", loadErr.Error())
		}

	}
}

func TestNoConfigFilesByMissingFolderFixed(t *testing.T) {
	var cfg chainConfig
	didWeHit := false
	chainCfg := yacl.New(&cfg, yamc.NewYamlReader()).
		SetSubDirs("v10").
		SetExpectNoConfigFiles(). // it is okay for not having any configuration yet.
		Init(nil, func(errCode int) error {
			if errCode == yacl.ERROR_PATH_NOT_EXISTS {
				didWeHit = true
			}
			return nil
		})

	loadErr := chainCfg.Load()
	if loadErr != nil {
		t.Error(loadErr)
	}
	if !didWeHit {
		t.Error("we should get the ERROR_PATH_NOT_EXISTS message")
	}
}

func TestOneFileNameusage(t *testing.T) {
	var cfg chainConfig

	chainCfg := yacl.New(&cfg, yamc.NewYamlReader()).
		SetSubDirs("v2").
		AllowSubdirs().
		SetSingleFile("001-test.base.yml")

	loadErr := chainCfg.Load()
	if loadErr != nil {
		t.Error(loadErr)
	}

	filesAll := strings.Join(chainCfg.GetAllParsedFiles(), ", ")
	if filesAll != "v2/001-test.base.yml, v2/deployEu/001-test.base.yml, v2/deployUs/001-test.base.yml" {
		t.Error("error on loading the expected files ", filesAll)
	}
}
