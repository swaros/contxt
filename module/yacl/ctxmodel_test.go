package yacl_test

import (
	"path/filepath"
	"testing"

	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yamc"
)

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
