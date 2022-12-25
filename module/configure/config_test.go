package configure_test

import (
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yamc"
)

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
