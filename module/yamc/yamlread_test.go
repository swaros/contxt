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

func TestStructRead(t *testing.T) {
	type testStruct struct {
		Name  string   `yaml:"name"`
		Age   int      `yaml:"age"`
		IsVip bool     `yaml:"isvip"`
		Tags  []string `yaml:"tags,omitempty"`
	}

	rdr := yamc.NewYamlReader()

	var cfg testStruct
	source := `
name: test
age: 42
isvip: true
tags:
  - first
  - second
`
	stringToRead := []byte(source)
	if err := rdr.Unmarshal(stringToRead, &cfg); err != nil {
		t.Error(err)
	}

	if cfg.Name != "test" {
		t.Error("name is not test")
	}

	if cfg.Age != 42 {
		t.Error("age is not 42")
	}

	if !cfg.IsVip {
		t.Error("isvip is not true")
	}

	if len(cfg.Tags) != 2 {
		t.Error("wrong count of tags")
	}

	if len(cfg.Tags) > 0 && cfg.Tags[0] != "first" {
		t.Error("tag 0 is not first")
	}

	if len(cfg.Tags) > 1 && cfg.Tags[1] != "second" {
		t.Error("tag 1 is not second")
	}

	// we should have fields.
	if !rdr.HaveFields() {
		t.Error("have fields is false")
	}

	// we should then also get the fields
	fields := rdr.GetFields()
	if !fields.Init {
		t.Error("init is false. should be true after reporting we have fields")
	}

	// getting a specific field, by the name, we defined on the begining,  should not fail
	// here we have to use the name in the Struct (first upper case) and not what we defined for the yaml data
	if fieldName, err := fields.GetField("Name"); err != nil {
		t.Error(err)
	} else {
		// then
		if fieldName.Name != "Name" {
			t.Error("field name is not Name")
		}
		// here we get how the the field is read from the yaml data
		if fieldName.OrginalTag.TagRenamed != "name" {
			t.Error("field tag is not name")
		}
	}

	// like the name test, we should also get the tags field
	if fieldTags, err := fields.GetField("Tags"); err != nil {
		t.Error(err)
	} else {
		// then
		if fieldTags.Name != "Tags" {
			t.Error("field name is not Tags")
		}
		// here we get how the the field is read from the yaml data
		if fieldTags.OrginalTag.TagRenamed != "tags" {
			t.Error("field tag is not tags")
		}
	}
}

func TestFailLoadTab(t *testing.T) {
	type testStruct struct {
		Name  string   `yaml:"name"`
		Age   int      `yaml:"age"`
		IsVip bool     `yaml:"isvip"`
		Tags  []string `yaml:"tags,omitempty"`
	}

	rdr := yamc.NewYamlReader()

	var cfg testStruct
	source := `
name: test
age: 42
isvip: true
tags:
	- first
  - second
`
	stringToRead := []byte(source)
	if err := rdr.Unmarshal(stringToRead, &cfg); err == nil {
		t.Error("should fail on tab")
	}
}
