package yacl_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yamc"
)

// structure of the old configuration
type Configuration struct {
	CurrentSet string
	Paths      []string
	LastIndex  int
	PathInfo   map[string]WorkspaceInfo
}
type WorkspaceInfo struct {
	Project string
	Role    string
	Version string
}

// new version of the configuration starts here
type ConfigMetaV2 struct {
	CurrentSet string
	Configs    map[string]ConfigurationV2
}

type WorkspaceInfoV2 struct {
	Path    string
	Project string
	Role    string
	Version string
}

type ConfigurationV2 struct {
	Name  string
	Paths map[string]WorkspaceInfoV2
}

func TestMigrateToV2(t *testing.T) {
	var cfgV1 Configuration // the old config
	var cfgV2 ConfigMetaV2  // the new configuration structure

	tim := time.Now()

	newConfig := yacl.NewConfig(&cfgV2, yamc.NewYamlReader()).
		Init(func(strct *any) {
			cfgV2.Configs = make(map[string]ConfigurationV2)

		}, nil).
		SetSubDirs("tmpfiles").
		SetSingleFile(tim.Format("2006-01-02_15_04_05_") + "newConfig.yml"). // create a file that should be uniue enough
		Empty()

	contxtCfg := yacl.NewConfig(&cfgV1, yamc.NewJsonReader()).
		SetSubDirs("testdata").
		SupportMigrate(func(path string, cfg interface{}) {

			// special configs by there name.

			// this one contains the name of the current used config
			if filepath.Clean(path) == filepath.Clean("testdata/contxt_current_config.json") {
				cfgV2.CurrentSet = cfgV1.CurrentSet
				return // just get out. we do not need something else from this file
			}

			// this one is not needed in any case
			if filepath.Clean(path) == filepath.Clean("testdata/default_contxt_ws.json") {
				return // just get out. we do not need something else from this file
			}

			// here we check for any reported path, if the reference of cfgV1 is currently set to them
			// so we get the path and we know what the loaded configshould contains
			// so we check the contxt config
			if filepath.Clean(path) == filepath.Clean("testdata/contxt.json") {
				if cfgV1.CurrentSet != "contxt" {
					t.Error("expect content is different to", cfgV1.CurrentSet)
				}
			}
			// just to check another one
			if filepath.Clean(path) == filepath.Clean("testdata/develop.json") {
				if cfgV1.CurrentSet != "develop" {
					t.Error("expect content is different to", cfgV1.CurrentSet)
				}
			}

			// here we have the logic to convert the configuration
			var cfgEntrie ConfigurationV2 = ConfigurationV2{Name: cfgV1.CurrentSet}
			for index, inPath := range cfgV1.Paths {
				var cfgPaths WorkspaceInfoV2 = WorkspaceInfoV2{
					Path: inPath,
				}
				if cfgEntrie.Paths == nil {
					cfgEntrie.Paths = make(map[string]WorkspaceInfoV2)
				}
				keyStr := fmt.Sprintf("%d", index)
				cfgEntrie.Paths[keyStr] = cfgPaths
			}
			// we make this already in the Init func
			cfgV2.Configs[cfgV1.CurrentSet] = cfgEntrie
		})

	contxtCfg.Load() // process the obsolet configs

	// check if they ar loaded
	// markup-string is the last file they should used. so we should have this as current config
	if cfgV1.CurrentSet != "markup-string" {
		t.Error("we expect the current set is markup-string. got ", cfgV1.CurrentSet)
	}
	l := len(contxtCfg.GetAllParsedFiles())
	if l != 7 {
		t.Error("unexpected amount of loaded files", l, " ", strings.Join(contxtCfg.GetAllParsedFiles(), ";"))
	}

	// check the new config. depending the migrate function, we should have contxt now
	if cfgV2.CurrentSet != "contxt" {
		t.Error("the new config to not get the expected contxt. we got [", cfgV2.CurrentSet, "] instead")
	}

	// using gjson paths to check the content
	if entry, err := newConfig.GetValue("CurrentSet"); err != nil || entry != "contxt" {
		if err != nil {
			t.Error(err)
		} else { // no error ? then the value is not matching
			t.Error("did not match with contxt... ", entry)
		}
	}

	cfgStr, toStrErr := newConfig.ToString(yamc.NewYamlReader())
	if toStrErr != nil {
		t.Error(toStrErr)
	}

	if cfgStr == "" {
		t.Error("config should not being empty")
	}

	// save the file of the new config in the temp storage
	if err := newConfig.Save(); err != nil {
		t.Error(err)
	}
}

func TestContxtObsoleteCfg(t *testing.T) {
	var cfg Configuration

	contxtCfg := yacl.NewConfig(&cfg, yamc.NewJsonReader()).SetSingleFile("contxt.json").SetSubDirs("testdata")

	if err := contxtCfg.Load(); err != nil {
		t.Error(err)
	} else {
		if filepath.Clean(contxtCfg.GetLoadedFile()) != filepath.Clean("testdata/contxt.json") {
			t.Error("this file should not being used ", contxtCfg.GetLoadedFile())
		}

		if len(cfg.Paths) != 2 {
			t.Error("there should be 2 paths stored. but got ", len(cfg.Paths))
		} else {
			if cfg.Paths[0] != "C:\\Users\\thoma\\OneDrive\\Dokumente\\code\\projects\\contxt\\contxt" {
				t.Error("should be not ", cfg.Paths[0])
			}
			if cfg.Paths[1] != "C:\\Users\\thoma\\OneDrive\\Dokumente\\code\\projects\\managed-out" {
				t.Error("should be not ", cfg.Paths[1])
			}
		}

	}

}
