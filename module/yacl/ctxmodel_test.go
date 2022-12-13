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

	}

}
