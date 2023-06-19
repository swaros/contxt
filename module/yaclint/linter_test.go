package yaclint_test

import (
	"testing"

	"github.com/swaros/contxt/module/yacl"
	ctxlint "github.com/swaros/contxt/module/yaclint"
	"github.com/swaros/contxt/module/yamc"
)

func TestLintOkSimple(t *testing.T) {

	type dataSet struct {
		TicketNr int
		Comment  string
	}

	type testConfig struct {
		SourceCode         string    `yaml:"SourceCode"`
		BuildEngine        string    `yaml:"BuildEngine"`
		BuildEngineVersion string    `yaml:"BuildEngineVersion"`
		Targets            []string  `yaml:"Targets"`
		BuildSteps         []string  `yaml:"BuildSteps"`
		IsSystem           bool      `yaml:"IsSystem"`
		IsDefault          bool      `yaml:"IsDefault"`
		MainVersionNr      int       `yaml:"MainVersionNr"`
		DataSet            []dataSet `yaml:"DataSet"`
	}
	// here we load the config.
	// lint depends on a yacl instance, so we need to create one first
	var testConf testConfig
	configHndl := yacl.New(
		&testConf,
		yamc.NewYamlReader(),
	).SetSubDirs("testdata", "testConfig").
		SetSingleFile("valid.yml").
		UseRelativeDir()

	if err := configHndl.Load(); err != nil {
		t.Errorf("failed to load config: %v", err)
		t.Log(configHndl.GetLoadedFile())
	}
	// -- end of config loading

	chck := ctxlint.NewLinter(*configHndl)
	if chck == nil {
		t.Error("failed to create linter")
	}

	chck.Verify()
	chck.PrintDiff()
	t.Error("not implemented")

}
