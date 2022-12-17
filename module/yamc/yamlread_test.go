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

func TestYamlDecodecode(t *testing.T) {
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

	if cfg.Subs[0] != "first" {
		t.Error("sub 0 is not first")
	}

}
