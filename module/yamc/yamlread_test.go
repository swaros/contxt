package yamc_test

import (
	"testing"

	"github.com/swaros/contxt/module/yamc"
)

type configStruct struct {
	Name     string   `yaml:"name"`
	Path     string   `yaml:"path"`
	Boolflag bool     `yaml:"boolflag"`
	Subs     []string `yaml:"subs"`
}

func TestYamlDecode(t *testing.T) {
	rdr := yamc.NewYamlReader()

	var cfg configStruct
	if err := rdr.FileDecode("testdata/cfgv1.yml", &cfg); err != nil {
		t.Error(err)
	}

	if !cfg.Boolflag {
		t.Error("boolflag is false")
	}

	if cfg.Name != "test" {
		t.Error("name is not test")
	}

	if cfg.Path != "root/path" {
		t.Error("name is not root/path")
	}

	if len(cfg.Subs) < 1 {
		t.Error("wrong count of subs")
	}

	if len(cfg.Subs) > 0 && cfg.Subs[0] != "first" {
		t.Error("sub 0 is not first")
	}

}

func TestYamlDecodeError(t *testing.T) {
	rdr := yamc.NewYamlReader()

	var cfg configStruct
	if err := rdr.FileDecode("testdata/cfgv1.yml", cfg); err == nil {
		t.Error("yaml config used without pointer should end in an error")
	}

}
