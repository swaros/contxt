package yaclint_test

import (
	"fmt"
	"testing"

	"github.com/swaros/contxt/module/yacl"
	ctxlint "github.com/swaros/contxt/module/yaclint"
	"github.com/swaros/contxt/module/yamc"
)

const (
	FailIfNotEqual = 0
	FailIfLower    = 1
	FailIfHigher   = 2
)

// Helper function to load a config file and verify it
// subdir: the subdirectory where the config file is located
// file: the config file name
// config: the config struct. the config have to be a pointer to a struct
// expectedLevel: the expected issue level. if the highest issue level is higher than expectedLevel, the test fails
// mode: the mode to use. 0 = value must be equal, 1 = value must be higher or equal, 2 = value must be lower or equal
func assertIssueLevelByConfig(t *testing.T, subdir, file string, config interface{}, expectedLevel int, mode int) *ctxlint.Linter {
	t.Helper()
	configHndl := yacl.New(
		config,
		yamc.NewYamlReader(),
	).SetSubDirs("testdata", subdir).
		SetSingleFile(file).
		UseRelativeDir()

	if err := configHndl.Load(); err != nil {
		t.Error(err)
		return nil
	}

	chck := ctxlint.NewLinter(*configHndl)
	if chck == nil {
		t.Error("failed to create linter")
		return nil
	}

	chck.Verify()
	errormsg := ""
	isFailed := false
	switch mode {
	case 0:
		isFailed = chck.GetHighestIssueLevel() != expectedLevel

		errormsg = fmt.Sprintf("the highest issue level is not equal to expected. expected %d got %d", expectedLevel, chck.GetHighestIssueLevel())
	case 1:
		isFailed = chck.GetHighestIssueLevel() < expectedLevel
		errormsg = fmt.Sprintf("the highest issue level is lower than expected. expected %d got %d", expectedLevel, chck.GetHighestIssueLevel())
	case 2:
		isFailed = chck.GetHighestIssueLevel() > expectedLevel
		errormsg = fmt.Sprintf("the highest issue level is higher than expected. expected %d got %d", expectedLevel, chck.GetHighestIssueLevel())
	}

	if isFailed {
		t.Error(errormsg)
		t.Log("\n" + chck.PrintIssues())
		diff, err := chck.GetDiff()
		if err != nil {
			t.Error(err)
			return nil
		}
		t.Log("\n" + diff + "\n")
		return nil
	}
	return chck

}

// TestConfig1 tests a valid config file
// this test is similar to TestConfigNo1
// the difference is, that we use the helper function lintAssertByFile in TestConfigNo1.
// depending on the additional layer of abstraction, it might be that based on some internal changes, the test fails.
// so it is important to have a test that tests the helper function itself. (booth tests fails or just one of them!?)
func TestConfig1(t *testing.T) {
	type dataSet struct {
		TicketNr int
		Comment  string
	}

	type mConfig struct {
		SourceCode         string    `yaml:"SourceCode"`
		BuildEngine        string    `yaml:"BuildEngine"`
		BuildEngineVersion string    `yaml:"BuildEngineVersion"`
		Targets            []string  `yaml:"Targets"`
		BuildSteps         []string  `yaml:"BuildSteps"`
		IsSystem           bool      `yaml:"IsSystem"`
		IsDefault          bool      `yaml:"IsDefault"`
		MainVersionNr      int       `yaml:"MainVersionNr"`
		DataSet            []dataSet `yaml:"DataSet,omitempty"`
	}
	var testConf mConfig

	configHndl := yacl.New(
		&testConf,
		yamc.NewYamlReader(),
	).SetSubDirs("testdata", "testConfig").
		SetSingleFile("valid.yml").
		UseRelativeDir()

	if err := configHndl.Load(); err != nil {
		t.Error(err)

	}

	chck := ctxlint.NewLinter(*configHndl)
	if chck == nil {
		t.Error("failed to create linter")
	}

	chck.Verify()

	expected := 2
	if chck.GetHighestIssueLevel() > expected {
		t.Error("found errors in valid config. expected issue level not higher than ", expected, ". got", chck.GetHighestIssueLevel())
		t.Log(chck.PrintIssues())
	}

}

func TestConfigNo1(t *testing.T) {
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
		DataSet            []dataSet `yaml:"DataSet,omitempty"`
	}
	var testConf testConfig

	assertIssueLevelByConfig(t, "testConfig", "valid.yml", &testConf, ctxlint.ValueNotMatch, FailIfNotEqual)
}

func TestConfigNo2(t *testing.T) {
	type dataSet struct {
		TicketNr int
		Comment  string
	}

	type tConfig struct {
		SourceCode         string    `yaml:"SourceCode"`
		BuildEngine        string    `yaml:"BuildEngine"`
		BuildEngineVersion string    `yaml:"BuildEngineVersion"`
		Targets            []string  `yaml:"Targets"`
		BuildSteps         []string  `yaml:"BuildSteps"`
		IsSystem           bool      `yaml:"IsSystem"`
		IsDefault          bool      `yaml:"IsDefault"`
		MainVersionNr      int       `yaml:"MainVersionNr"`
		DataSet            []dataSet `yaml:"DataSet,omitempty"`
	}
	var testConf tConfig
	// we expect to fail, because the config file contains unknown fields
	assertIssueLevelByConfig(t, "testConfig", "invalid_types.yml", &testConf, ctxlint.UnknownEntry, FailIfNotEqual)

}
