package yaclint_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yaclint"
	"github.com/swaros/contxt/module/yamc"
)

const (
	FailIfNotEqual = 0
	FailIfLower    = 1
	FailIfHigher   = 2
)

var (
	filterTags = []string{"MatchToken.getNameOf:", "!No Tag found!"}
	noFilter   = []string{}
	usedFilter = filterTags
)

// Helper function to load a config file and verify it
// subdir: the subdirectory where the config file is located
// file: the config file name
// config: the config struct. the config have to be a pointer to a struct
// expectedLevel: the expected issue level. if the highest issue level is higher than expectedLevel, the test fails
// mode: the mode to use. 0 = value must be equal, 1 = value must be higher or equal, 2 = value must be lower or equal
func assertIssueLevelByConfig(t *testing.T, subdir, file string, config interface{}, expectedLevel int, mode int) *yaclint.Linter {
	t.Helper()
	configHndl := yacl.New(
		config,
		yamc.NewYamlReader(),
	).SetSubDirs("testdata", subdir).
		SetSingleFile(file).
		UseRelativeDir()

	if err := configHndl.Load(); err != nil {
		t.Error("Load failed")
		t.Error(err)
		return nil
	}

	chck := yaclint.NewLinter(*configHndl)
	chck.SetDirtyLogger(yaclint.NewDirtyLogger())
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
		t.Log("\n" + diff + "\n" + chck.GetTrace(usedFilter...))
		t.SkipNow()
		// reset allways to tag filter
		usedFilter = filterTags
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

	chck := yaclint.NewLinter(*configHndl)
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

func TestConfigLower(t *testing.T) {
	type dataSet struct {
		TicketNr int
		Comment  string
	}

	type mConfig struct {
		SourceCode         string    `yaml:"sourceCode"`
		BuildEngine        string    `yaml:"buildEngine"`
		BuildEngineVersion string    `yaml:"buildEngineVersion"`
		Targets            []string  `yaml:"targets"`
		BuildSteps         []string  `yaml:"buildSteps"`
		IsSystem           bool      `yaml:"isSystem"`
		IsDefault          bool      `yaml:"isDefault"`
		MainVersionNr      int       `yaml:"mainVersionNr"`
		DataSet            []dataSet `yaml:"dataSet,omitempty"`
	}
	var testConf mConfig

	yLoader := yamc.NewYamlReader()
	configHndl := yacl.New(
		&testConf,
		yLoader,
	).SetSubDirs("testdata", "testConfig").
		SetSingleFile("valid_lowercase.yml").
		UseRelativeDir()

	if err := configHndl.Load(); err != nil {
		t.Error(err)
	}

	chck := yaclint.NewLinter(*configHndl)
	chck.SetDirtyLogger(yaclint.NewDirtyLogger())
	if chck == nil {
		t.Error("failed to create linter")
	}

	chck.Verify()

	expected := 2
	if chck.GetHighestIssueLevel() > expected {
		t.Error("found errors in valid config. expected issue level not higher than ", expected, ". got", chck.GetHighestIssueLevel())
		t.Log(chck.PrintIssues())
	}

	if chck.HasWarning() {
		t.Error("found warnings in valid config. expected no warnings")
		t.Log(chck.PrintIssues())
		t.Log(chck.GetTrace())
	}

	if !chck.HasInfo() {
		t.Error("found no info in valid config. expected info")

		t.Log(chck.GetTrace())
	}

}

func TestConfigValidJson(t *testing.T) {
	type dataSet struct {
		TicketNr int
		Comment  string
	}

	type mConfig struct {
		SourceCode    string    `json:"SourceCode"`
		BuildEngine   string    `json:"BuildEngine"`
		BuildVersion  string    `json:"BuildVersion"`
		Targets       []string  `json:"Targets"`
		BuildSteps    []string  `json:"BuildSteps"`
		IsSystem      bool      `json:"IsSystem"`
		IsDefault     bool      `json:"IsDefault"`
		MainVersionNr int       `json:"MainVersionNr"`
		DataSet       []dataSet `json:"DataSet,omitempty"`
	}
	var testConf mConfig

	yLoader := yamc.NewJsonReader()
	configHndl := yacl.New(
		&testConf,
		yLoader,
	).SetSubDirs("testdata", "testConfig").
		SetSingleFile("valid.json").
		UseRelativeDir()

	if err := configHndl.Load(); err != nil {
		t.Error(err)

	}

	chck := yaclint.NewLinter(*configHndl)
	if chck == nil {
		t.Error("failed to create linter")
	}

	if err := chck.Verify(); err != nil {
		t.Error(err)
	}

	if chck.GetHighestIssueLevel() > 0 {
		t.Error("found errors in valid config. expected issue level not higher than 0. got", chck.GetHighestIssueLevel())
		t.Log(chck.PrintIssues())

		t.Log(chck.GetTrace())
	}

	if testConf.SourceCode != "module/yaclint/linter_test.go" {
		t.Error("expected SourceCode to be 'module/yaclint/linter_test.go'. got", testConf.SourceCode)
	}

	if testConf.BuildEngine != "go" {
		t.Error("expected BuildEngine to be 'go'. got", testConf.BuildEngine)
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

	assertIssueLevelByConfig(t, "testConfig", "valid.yml", &testConf, yaclint.ValueNotMatch, FailIfNotEqual)
}

func TestConfigNo1DifferentYamlKeywords(t *testing.T) {
	type dataSet struct {
		TicketNr int
		Comment  string
	}

	type testConfig struct {
		SourceCode         string    `yaml:"sourceCode"`
		BuildEngine        string    `yaml:"buildEngine"`
		BuildEngineVersion string    `yaml:"buildEngineVersion"`
		Targets            []string  `yaml:"targets"`
		BuildSteps         []string  `yaml:"buildSteps"`
		IsSystem           bool      `yaml:"isSystem"`
		IsDefault          bool      `yaml:"isDefault"`
		MainVersionNr      int       `yaml:"mainVersionNr"`
		DataSet            []dataSet `yaml:"dataSet,omitempty"`
	}
	var testConf testConfig
	assertIssueLevelByConfig(t, "testConfig", "valid_lowercase.yml", &testConf, yaclint.ValueNotMatch, FailIfNotEqual)
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
	assertIssueLevelByConfig(t, "testConfig", "invalid_types.yml", &testConf, yaclint.UnknownEntry, FailIfNotEqual)

}

// same as TestConfigNo2, but we check more than the issue level
func TestConfigNo2MoreChecks(t *testing.T) {
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
	verifier := assertIssueLevelByConfig(t, "testConfig", "invalid_types.yml", &testConf, yaclint.UnknownEntry, FailIfNotEqual)

	HitUnkonwnEntryCount := 0
	MissingEntryCount := 0
	UnexpectedEntryCount := 0
	verifier.WalkIssues(func(token *yaclint.MatchToken, added bool) {

		if token.Status == yaclint.UnknownEntry {
			HitUnkonwnEntryCount++
		} else if token.Status == yaclint.MissingEntry {
			MissingEntryCount++
		} else {
			UnexpectedEntryCount++
		}

	})

	ExpectHitUnkonwnEntryCount := 1
	if HitUnkonwnEntryCount != ExpectHitUnkonwnEntryCount {
		t.Error("expected to find ", ExpectHitUnkonwnEntryCount, "unknown entries. got", HitUnkonwnEntryCount)
		t.Log("\n" + verifier.PrintIssues())
	}

	ExpectMissingEntryCount := 8
	if MissingEntryCount != ExpectMissingEntryCount {
		t.Error("expected to find ", ExpectMissingEntryCount, "missing entries. got", MissingEntryCount)
		t.Log("\n" + verifier.PrintIssues())
	}

	ExpectUnexpectedEntryCount := 0
	if UnexpectedEntryCount != ExpectUnexpectedEntryCount {
		t.Error("expected to find ", ExpectUnexpectedEntryCount, "unexpected entries. got", UnexpectedEntryCount)
		t.Log("\n" + verifier.PrintIssues())
	}
}

func TestLoadFailure(t *testing.T) {
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
	configHndl := yacl.New(
		&testConf,
		yamc.NewYamlReader(),
	).SetSubDirs("testdata", "testConfig").
		SetSingleFile("valid.yml").
		UseRelativeDir()

	// using linter without loading the config should fail
	chck := yaclint.NewLinter(*configHndl)
	if chck == nil {
		t.Error("failed to create linter")
	}

	if vErr := chck.Verify(); vErr == nil {
		t.Error("expected to fail, because config was not loaded")
	} else {
		if vErr.Error() != "no reader found. the config needs to be loaded first" {
			t.Error("we expected a different error message then: ", vErr.Error())
		}
	}

}

// similar to MatchToken, but with less fields that we can use as expected values
type assertTokenSimplify struct {
	KeyWord    string
	Value      interface{}
	Type       string
	Added      bool
	IndexNr    int
	SequenceNr int
	Status     int
	IsChecked  bool // true if the token was checked
}

func (a *assertTokenSimplify) String() string {
	typeOfValue := reflect.TypeOf(a.Value)
	return fmt.Sprintf(
		"KeyWord: '%s', Value: %v[%v], Type: %s, Added: %v, IndexNr: %d, SequenceNr: %d, Status: %d, IsChecked: %v",
		a.KeyWord, a.Value, typeOfValue, a.Type, a.Added, a.IndexNr, a.SequenceNr, a.Status, a.IsChecked)
}

// helper function to get the expected token from a slice
// the token have to match on keyword, added, Indexnr, sequenceNr and ischecked.
// isChecked is set to any token that is already checked.
func helperGetExpectedFromSlice(name string, added bool, indexNr int, sequenceNr int, from []*assertTokenSimplify) *assertTokenSimplify {
	for _, v := range from {
		if v.KeyWord == name && v.Added == added && indexNr == v.IndexNr && sequenceNr == v.SequenceNr && v.IsChecked == false {
			return v
		}
	}
	return nil
}

func helperGetExpectedFromSliceBynameAnAdded(name string, added bool, from []*assertTokenSimplify) *[]assertTokenSimplify {
	var result []assertTokenSimplify
	for _, v := range from {
		if v.KeyWord == name && v.Added == added && v.IsChecked == false {
			result = append(result, *v)
		}
	}
	return &result
}

// Testing some results in the ReportDiff callback.
// We expect to find the following tokens:
// this test can be akward because of flacky results.
// yaml implementation is not stable and while encode/decode a clear numeric value can be converted to a string.
// this is not a problem for the yaml structure, but it is for us.
// so IF this is sometime changed/fixed, this test would fail, just because it will not report the following issues.
func TestReportDiffStartedAt(t *testing.T) {
	//t.Skip("not implemented yet. have to think about it")
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
	linter := assertIssueLevelByConfig(t, "testConfig", "some_fails.yml", &testConf, yaclint.ValueNotMatch, FailIfNotEqual)
	if linter == nil {
		t.Error("failed to create linter")
		return
	}

	expectedTokens := []*assertTokenSimplify{
		{"BuildEngineVersion", 1.14, "float64", false, 1, 2, yaclint.ValueMatchButTypeDiffers, false},
		{"BuildEngineVersion", "1.14", "string", true, 1, 2, yaclint.ValueMatchButTypeDiffers, false},
		//{"  Comment", "", "string", true, 1, 7, yaclint.ValueNotMatch, false},
		//{"  Comment", "this is a comment", "string", false, 1, 7, yaclint.ValueNotMatch, false},
		//{"  TicketNr", 1, "int", false, 2, 7, yaclint.ValueNotMatch, false},
		//{"  TicketNr", 0, "int", true, 2, 7, yaclint.ValueNotMatch, false},
	}

	checkIndex := 0
	reportNotFound := false // report all tokens that are not found. this helps while setting up the test to focus on value that are found but differs
	reportedTokens := []*assertTokenSimplify{}
	linter.GetIssue(0, func(token *yaclint.MatchToken) {

		reportToken := &assertTokenSimplify{
			KeyWord:    token.KeyWord,
			Value:      token.Value,
			Type:       token.Type,
			Added:      token.Added,
			IndexNr:    token.IndexNr,
			SequenceNr: token.SequenceNr,
			Status:     token.Status,
			IsChecked:  false,
		}
		reportedTokens = append(reportedTokens, reportToken)

		assertToken := helperGetExpectedFromSlice(token.KeyWord, token.Added, token.IndexNr, token.SequenceNr, expectedTokens)
		if assertToken == nil {
			if reportNotFound {
				t.Error("unexpected token", token.ToString())
			}
			return
		}
		assertToken.IsChecked = true
		indexIdent := strconv.Itoa(checkIndex) + "/" + strconv.Itoa(len(expectedTokens)) + " [" + assertToken.KeyWord + "] "
		if token.KeyWord != assertToken.KeyWord {
			t.Error(indexIdent, "expected keyword to be [", assertToken.KeyWord, "] got [", token.KeyWord, "]")
			t.Log(" <-- skip the rest of the test because the keyword is already wrong")
			return // no need to check the rest because if the keyword is already wrong, the other fields are also wrong
		}
		if token.Value != assertToken.Value {
			t.Error(indexIdent, "expected value to be (", assertToken.Value, ") got (", token.Value, ") ", token.ToString())
		}
		if token.Type != assertToken.Type {
			t.Error(indexIdent, "expected type to be ", assertToken.Type, "got", token.Type)
		}

		if token.Added != assertToken.Added {
			t.Error(indexIdent, "expected added to be ", assertToken.Added, "got", token.Added)
		}
		if token.SequenceNr != assertToken.SequenceNr {
			t.Error(indexIdent, "expected sequenceNr to be ", assertToken.SequenceNr, "got", token.SequenceNr)
		}
		if token.Status != assertToken.Status {
			t.Error(indexIdent, "expected status to be ", assertToken.Status, "got", token.Status, "(", token.ToString(), ")")
		}

		checkIndex++

	})
	// now check if we do not check all tokens
	for _, v := range expectedTokens {
		if !v.IsChecked {
			closestToken := helperGetExpectedFromSliceBynameAnAdded(v.KeyWord, v.Added, reportedTokens)
			if closestToken == nil || len(*closestToken) == 0 {
				t.Log(" >> no possible match found")
				outStr := ""
				for _, v := range reportedTokens {
					outStr += v.String() + "\n"
				}
				t.Log(" >> reported tokens:\n", outStr)
			} else {
				for _, v := range *closestToken {
					t.Log(" >> possible match:\n", v.String())
				}
			}

			t.Error(
				"we did not found token\n", v.String(),
				"\nseems it is not reported in the the diff. ARE YOU SURE this token should be have an diff?\n",
				"reported diffs:\n", linter.PrintIssues(),
			)

		}
	}

	diff, linErr := linter.GetDiff()
	if linErr != nil {
		t.Error("failed to get diff", linErr)
		return
	}
	// we did not check the diff, because we are not the diff package. but we can check if the diff is empty
	if diff == "" {
		t.Error("diff is empty")
		return
	}

	reportOut := linter.PrintIssues()
	if reportOut == "" {
		t.Error("report is empty")
		return
	}
}

func TestBasicExample(t *testing.T) {
	type Config struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}
	config := &Config{}
	cfgApp := yacl.New(
		config,
		yamc.NewYamlReader(),
	)
	// load the config file. must be done before the linter can be used
	if err := cfgApp.SetSubDirs("example", "basic").LoadFile("config.yaml"); err != nil {
		t.Error(err)
	}

	// create a new linter instance
	linter := yaclint.NewLinter(*cfgApp)
	// error if remapping is not possible. so no linting error
	if err := linter.Verify(); err != nil {
		t.Error(err)
	}

	// if we found any issues, then the issuelevel is not 0
	if linter.GetHighestIssueLevel() > 0 {
		// just print the issues
		fmt.Println(linter.PrintIssues())

		fmt.Println(linter.GetDiff())
		t.Error("we found issues")
	}
}

func TestBasicExamplePointerError(t *testing.T) {
	type Config struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}
	config := &Config{}
	cfgApp := yacl.New(
		&config,
		yamc.NewYamlReader(),
	)
	// load the config file. must be done before the linter can be used
	if err := cfgApp.SetSubDirs("example", "basic").LoadFile("config.yaml"); err != nil {
		t.Error(err)
	}

	// create a new linter instance
	linter := yaclint.NewLinter(*cfgApp)
	// error if remapping is not possible. so no linting error
	if err := linter.Verify(); err != nil {
		t.Error(err)
	}

	// just print the issues
	if reason, haveError := linter.HaveParsingError(); haveError {
		expectedReason := "pointers are not supported"
		if reason != expectedReason {
			t.Error("we expected to find a parsing error with reason", expectedReason, "got", reason)
		}
	} else {
		t.Error("we expected to find a parsing error")
	}

}

func TestUnexpectedExample(t *testing.T) {
	type Config struct {
		Name    string `yaml:"name"`
		Contact struct {
			Email string `yaml:"email"`
			Phone string `yaml:"phone"`
		} `yaml:"contact"`
		LastName string `yaml:"lastname"`
		Age      int    `yaml:"age"`
	}

	// usual yacl stuff
	config := &Config{}
	cfgApp := yacl.New(
		config,
		yamc.NewYamlReader(),
	)
	if err := cfgApp.SetSubDirs("example", "unexpected01").LoadFile("contact2.yaml"); err != nil {
		panic(err)
	}

	// now the linter
	linter := yaclint.NewLinter(*cfgApp)
	if err := linter.Verify(); err != nil {
		panic(err)
	}

	// do we have any issues?
	if linter.GetHighestIssueLevel() > 1 {
		t.Error("we found issues. but we did not expect any")
		// first we can print the issues
		fmt.Println(linter.PrintIssues())

		linter.GetIssue(0, func(token *yaclint.MatchToken) {
			fmt.Println(token.ToString())
		})
	} else {
		fmt.Println("no issues found")
	}
}

func TestConfigStruct(t *testing.T) {
	type targets struct {
		Name     string `yaml:"name"`
		SureName string `yaml:"surename"`
	}

	type slConfig struct {
		Main    string    `yaml:"main"`
		Targets []targets `yaml:"targets"`
	}
	var testConf slConfig

	configHndl := yacl.New(
		&testConf,
		yamc.NewYamlReader(),
	).SetSubDirs("testdata", "structAsSlice").
		SetSingleFile("test1.yaml").
		UseRelativeDir()

	if err := configHndl.Load(); err != nil {
		t.Error(err)

	}

	chck := yaclint.NewLinter(*configHndl)
	if chck == nil {
		t.Error("failed to create linter")
	}
	chck.SetDirtyLogger(yaclint.NewDirtyLogger().CreateCtxoutTracer())
	chck.Verify()

	expected := 2
	if chck.GetHighestIssueLevel() > expected {
		t.Error("found errors in valid config. expected issue level not higher than ", expected, ". got", chck.GetHighestIssueLevel())
		t.Log(chck.PrintIssues())

		t.Log(chck.GetTrace())
	}

}

func TestConfigStruct2(t *testing.T) {
	// test struct with tags
	type worker struct {
		Name     string `yaml:"name"`
		SureName string `yaml:"surename"`
	}

	type targets struct {
		Worker []worker `yaml:"worker"`
		Labels []string `yaml:"labels"`
	}

	type testConfig struct {
		Main    string  `yaml:"main"`
		Targets targets `yaml:"targets"`
	}
	var testConf testConfig
	usedFilter = noFilter
	assertIssueLevelByConfig(t, "structAsSlice", "test2.yaml", &testConf, yaclint.ValueNotMatch, FailIfHigher)
}

func BenchmarkLinterTest(b *testing.B) {
	// test struct with tags
	type worker struct {
		Name     string `yaml:"name"`
		SureName string `yaml:"surename"`
	}

	type targets struct {
		Worker []worker `yaml:"worker"`
		Labels []string `yaml:"labels"`
	}

	type testConfig struct {
		Main    string  `yaml:"main"`
		Targets targets `yaml:"targets"`
	}
	var config testConfig

	configHndl := yacl.New(
		&config,
		yamc.NewYamlReader(),
	).SetSubDirs("testdata", "structAsSlice").
		SetSingleFile("test2.yaml").
		UseRelativeDir()

	if err := configHndl.Load(); err != nil {
		b.Error("Load failed")
		b.Error(err)
		return
	}

	chck := yaclint.NewLinter(*configHndl)
	chck.SetDirtyLogger(yaclint.NewDirtyLogger())
	if chck == nil {
		b.Error("failed to create linter")
		return
	}

	if err := chck.Verify(); err != nil {
		b.Error("Verify failed")
		b.Error(err)
		return
	}
}

func BenchmarkLinter(b *testing.B) {
	// benchmark TestConfigStruct2
	b.Run("TestConfigStruct2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			BenchmarkLinterTest(b)
		}
	})
}
