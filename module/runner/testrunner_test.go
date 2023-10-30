package runner_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/runner"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yamc"
)

var useLastDir = "./"
var lastExistCode = 0
var testDirectory = ""

type ExpectDef struct {
	ExpectedInOutput []string "yaml:\"output\"" // what should be in the output (contains! not full match)
	ExpectedInError  []string "yaml:\"error\""  // what should be in the error (contains! not full match)
	NotExpected      []string "yaml:\"not\""    // what should not be in the output (contains! not full match)
}

type TestRunExpectation struct {
	TestName     string    "yaml:\"testName\"" // the nameof the test
	RunCmd       string    "yaml:\"runCmd\""   // the run command to execute
	Folder       string    "yaml:\"folder\""   // the folder where the test is located. empty to use the current directory
	Systems      []string  "yaml:\"systems\""  // what system is ment like linux, windows, darwin etc.
	Expectations ExpectDef "yaml:\"expect\""
}

func TestAllIssues(t *testing.T) {
	// walk on every file in the directory starting from ./testdata/issues
	// and run the test if the file is prefixed with 'issue_'

	// change to the testdata directory
	path := ChangeToRuntimeDir(t)

	// walk on every file in the directory
	// look for files prefixed with 'issue_'
	// and run the test
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// check if we have an file and if the file is prefixed with 'issue_' and has the suffix '.yml'
		// but use the filename for test prefix

		baseName := filepath.Base(path)
		if !info.IsDir() && strings.HasPrefix(baseName, "issue_") && strings.HasSuffix(path, ".yml") {

			// load the test definition
			testDef := TestRunExpectation{}
			loader := yacl.New(&testDef, yamc.NewYamlReader()).SetFileAndPathsByFullFilePath(path)
			if err := loader.Load(); err != nil {
				t.Error(err)
			}
			if loader.GetLoadedFile() == "" {
				t.Error("could not load the test definition file: " + path)
			} else {

				if testDef.Systems != nil && len(testDef.Systems) > 0 {
					// check if the current system is in the list of systems
					// if not, we skip the test
					if !systools.StringInSlice(runtime.GOOS, testDef.Systems) {
						return nil
					}
				}

				testDef.Folder, _ = filepath.Abs(filepath.Dir(path))
				IssueTester(t, testDef)
			}
		}
		return nil

	})
	if err != nil {
		panic(err)
	}
}

func IssueTester(t *testing.T, testDef TestRunExpectation) {
	t.Helper()
	tasks.NewGlobalWatchman().ResetAllTaskInfos()
	ChangeToRuntimeDir(t)
	app, output, appErr := SetupTestApp("issues", "ctx_test_config.yml")
	if appErr != nil {
		t.Errorf("Expected no error, got '%v'", appErr)
	}
	cleanAllFiles()
	defer cleanAllFiles()

	// set the log file with an timestamp
	logFileName := testDef.TestName + time.Now().Format(time.RFC3339) + ".log"
	output.SetLogFile(getAbsolutePath(logFileName))
	output.ClearAndLog()
	currentDir := dirhandle.Pushd()
	defer currentDir.Popd()

	// change into the test directory
	if testDef.Folder != "" {
		if err := os.Chdir(testDef.Folder); err != nil {
			t.Errorf("error by changing the directory. check test: '%v'", err)
		}
	}

	assertSomething := 0

	if err := runCobraCmd(app, testDef.RunCmd); err != nil {
		if len(testDef.Expectations.ExpectedInError) > 0 {
			for _, expected := range testDef.Expectations.ExpectedInError {
				assertInMessage(t, output, expected)
				assertSomething++
			}
		} else {
			t.Errorf("Expected no error, got '%v'", err)
		}
	}
	if len(testDef.Expectations.ExpectedInOutput) > 0 {
		for _, expected := range testDef.Expectations.ExpectedInOutput {
			assertInMessage(t, output, expected)
			assertSomething++
		}
	}

	// looking for the not expected
	if len(testDef.Expectations.NotExpected) > 0 {
		for _, expected := range testDef.Expectations.NotExpected {
			assertNotInMessage(t, output, expected)
			assertSomething++
		}
	}
	output.ClearAndLog()
	if assertSomething == 0 {
		t.Error("no expectations was set and tested. please check the test definition")
	}

}

func RuntimeFileInfo(t *testing.T) string {
	_, filename, _, _ := runtime.Caller(0)
	return filename
}

func ChangeToRuntimeDir(t *testing.T) string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	t.Log("change to dir: " + dir)
	if err := os.Chdir(dir); err != nil {
		t.Error(err)
	}
	return dir
}

// shortcut for running a cobra command
// without any other setup
func runCobraCommand(runnCallback func(cobra *runner.SessionCobra, writer io.Writer)) string {
	cobra := runner.NewCobraCmds()
	cmpltn := new(bytes.Buffer)
	if runnCallback != nil {
		runnCallback(cobra, cmpltn)
	}
	return cmpltn.String()
}

// this are some helper functions especially for testing the runner
// Setup the test app
// create the application. set up the config folder name, and the name of the config file.
// the testapp bevavior is afterwards different, because it uses the config
// related to the current directory.
//
//	if the file should remover automatically, it needs prefixed by 'ctx_'.
//
// thats why we have some special helper functions.
//   - getAbsolutePath to get the absolute path to the testdata directory
//   - backToWorkDir to go back to the testdata directory
//   - cleanAllFiles to remove the config file
func SetupTestApp(dir, file string) (*runner.CmdSession, *TestOutHandler, error) {
	tasks.NewGlobalWatchman().ResetAllTaskInfos()
	file = strings.ReplaceAll(file, ":", "_")
	file = strings.ReplaceAll(file, "-", "_")
	file = strings.ReplaceAll(file, "+", "_")

	// first we want to catch the exist codes
	systools.AddExitListener("testing_prevent_exit", func(no int) systools.ExitBehavior {
		lastExistCode = no
		return systools.Interrupt
	})

	configure.USE_SPECIAL_DIR = false   // no special directory like userHome etc.
	configure.CONTXT_FILE = file        // set the configuration file name
	configure.MIGRATION_ENABLED = false // disable the migration
	configure.CONTEXT_DIR = dir         // set the directory name

	// save the current directory
	// and also get back to them (next time)
	popdTestDir()
	// we need to stick to the testdata directory
	// any other directory will not work
	if err := os.Chdir("./testdata"); err != nil {

		panic(err)
	}
	// check if the directory exists, that we want to use in the testdata directory.
	// even if the config package is able to create them, we want avoid this here.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		panic(err.Error() + "| the directory " + dir + " does not exist in the testdata directory")
	}

	// build the absolute path to the testdata directory
	// this is needed to go back to the testdata directory
	// if needed
	if pwd, derr := os.Getwd(); derr == nil {
		useLastDir = pwd
		configure.CONFIG_PATH_CALLBACK = func() string {
			return useLastDir + "/" + configure.CONTEXT_DIR + "/" + configure.CONTXT_FILE
		}
	} else {
		panic(derr)
	}

	app := runner.NewCmdSession()

	// set the TemplateHndl OnLoad function to parse required files
	// like it is done in the real application
	onLoadFn := func(template *configure.RunConfig) error {
		return app.SharedHelper.MergeRequiredPaths(template, app.TemplateHndl)
	}
	app.TemplateHndl.SetOnLoad(onLoadFn)

	functions := runner.NewCmd(app)
	// init the main functions
	functions.MainInit()

	// signs filter
	signsFilter := ctxout.NewSignFilter(ctxout.NewBaseSignSet())
	ctxout.AddPostFilter(signsFilter)
	// tabout filter
	tabouOutFilter := ctxout.NewTabOut()
	ctxout.AddPostFilter(tabouOutFilter)
	info := ctxout.PostFilterInfo{
		Width:      800,   // give us a big width so we can render the whole line
		IsTerminal: false, //no terminal
		Colored:    false, // no colors
		Height:     500,   // give us a big height so we can render the whole line
		Disabled:   true,
	}
	tabouOutFilter.Update(info)
	signsFilter.Update(info)
	signsFilter.ForceEmpty(true)

	if err := app.Cobra.Init(functions); err != nil {
		panic(err)
	}
	ctxout.ForceFilterUpdate(info)

	outputHdnl := NewTestOutHandler()
	app.OutPutHdnl = outputHdnl
	configure.GetGlobalConfig().ResetConfig()
	return app, outputHdnl, nil
}

// helper function to verify the configuration file.
// if the testCallBack is not nil, we will call it with the configuration model
// so we can check the content of the configuration file.
// this is helpfull just because to double check the content of the file itself and
// the current state of the configuration. the configuration can be different from the file.
// just because the configuration is in memory and the file is on the disk.
// this functions is all about checking if the configuration is updated correctly, also in the file content.
func verifyConfigurationFile(t *testing.T, testCallBack func(CFG *configure.ConfigMetaV2)) {
	t.Helper()
	file := ""
	if configure.CONFIG_PATH_CALLBACK != nil {
		file = configure.CONFIG_PATH_CALLBACK()
	}

	if file == "" {
		t.Error("configuration setup failed. could not determine the configuration file.")
		return
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Error("configuration file not found: ", file)
		return
	}
	// if the testCallBack is nil, we dont need to check the content
	if testCallBack == nil {
		return
	}
	// model
	var CFG configure.ConfigMetaV2 = configure.ConfigMetaV2{}
	// load the configuration file
	loader := yacl.New(&CFG, yamc.NewYamlReader()).SetFileAndPathsByFullFilePath(file)
	if err := loader.Load(); err != nil {
		t.Error(err)
	}
	testCallBack(&CFG)

}

// save and go back to the test folder
func popdTestDir() {
	// if not set, we get the current directory
	// and set them once.
	// so the carefully use this function in the first place
	if testDirectory == "" {
		if pwd, derr := os.Getwd(); derr == nil {
			testDirectory = pwd
		} else {
			panic(derr)
		}
	}

	if err := os.Chdir(testDirectory); err != nil {
		panic(err)
	}
}

// helper function to change back to the testdata directory
func backToWorkDir() {
	if err := os.Chdir(useLastDir); err != nil {
		panic(err)
	}
}

// helper function to get the absolute path to the testdata directory
func getAbsolutePath(dir string) string {

	dir = useLastDir + "/" + dir
	dir = filepath.Clean(dir)
	if filepath.IsAbs(dir) {
		return dir
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	return abs
}

// helper function to remove the config files
// from testdata/config folder
func cleanAllFiles() {
	popdTestDir()
	if err := os.Chdir("./testdata/config"); err != nil {
		panic(err)
	}
	// walk on every file in the directory
	// and remove it
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasPrefix(path, "ctx_") && strings.HasSuffix(path, ".yml") {
			return os.Remove(path)
		}
		return nil

	})
	if err != nil {
		panic(err)
	}
	popdTestDir()
}

// helper function to run a cobra command by argument line
func runCobraCmd(app *runner.CmdSession, cmd string) error {
	app.Cobra.RootCmd.SetArgs(strings.Split(cmd, " "))
	return app.Cobra.RootCmd.Execute()
}

// checks if the given string is part of the output buffer
// if not, it will fail the test.
// the special thing about this function is, that it will split the string
// by new line and check if every line is part of the output buffer.
// example:
//
//	output.ClearAndLog()
//	if err := runCobraCmd(app, "workspace new ducktale"); err != nil {
//	   t.Errorf("Expected no error, got '%v'", err)
//	 }
//	assertInMessage(t, output, "ducktale\ncreated\nproject")
func assertSplitTestInMessage(t *testing.T, output *TestOutHandler, msg string) {
	t.Helper()
	parts := strings.Split(msg, "\n")
	errorHappen := false
	for _, part := range parts {
		if part == "" {
			continue
		}
		if !output.Contains(part) {
			errorHappen = true
			t.Errorf("Expected [%s]not found in the output", part)
		}
	}
	if errorHappen {
		t.Error("this is the source output\n", output.String())
	}
}

// assert a string is part of the output buffer
func assertInMessage(t *testing.T, output *TestOutHandler, msg string) {
	t.Helper()
	if !output.Contains(msg) {
		t.Errorf("Expected \n%s\n-- but instead we did not found it in --\n%v\n", msg, output.String())
	}
}

// assert a string is part of the output buffer as regex
func assertRegexmatchInMessage(t *testing.T, output *TestOutHandler, msg string) {
	t.Helper()
	if !output.TestRegexPattern(msg) {
		t.Errorf("Expected \n%s\nbut instead we got\n%v", msg, output.String())
	}
}

// assert a string is not part of the output buffer
func assertNotInMessage(t *testing.T, output *TestOutHandler, msg string) {
	t.Helper()
	if output.Contains(msg) {
		t.Errorf("Expected '%s' is not in the message, but got '%v'", msg, output.String())
	}
}

// assert a cobra command is returning an error
func assertCobraError(t *testing.T, app *runner.CmdSession, cmd string, expectedMessageContains string) {
	t.Helper()
	if err := runCobraCmd(app, cmd); err == nil {
		t.Errorf("Expected error, but got none")
	} else {
		if !strings.Contains(err.Error(), expectedMessageContains) {
			t.Errorf("Expected error message to contain '%s', but got '%s'", systools.PadString(expectedMessageContains, 80), err.Error())
		}
	}
}

// checking if we are in the expected path in the operting system
func assertInOsPath(t *testing.T, path string) {
	t.Helper()
	if currentDir, err := os.Getwd(); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	} else {
		if currentDir != path {
			t.Errorf("Expected to be in '%v', got '%v'", path, currentDir)
		}
	}
}

var (
	AcceptFullMatch          = 1
	AcceptIgnoreLn           = 2
	AcceptContains           = 3
	AcceptContainsNoSpecials = 4
)

func assertStringFind(search, searchIn string, acceptableLevel int) bool {
	if search == "" || searchIn == "" {
		return true
	}
	if acceptableLevel >= AcceptFullMatch && searchIn == search {
		return true
	}
	if acceptableLevel >= AcceptIgnoreLn && searchIn == search+"\n" {
		return true
	}
	if acceptableLevel >= AcceptContains && strings.Contains(searchIn, search) {
		return true
	}

	if acceptableLevel >= AcceptContainsNoSpecials {
		search = strings.Replace(search, " ", "", -1)
		searchIn = strings.Replace(searchIn, " ", "", -1)
		search = strings.Replace(search, "\n", "", -1)
		searchIn = strings.Replace(searchIn, "\n", "", -1)
		search = strings.Replace(search, "\t", "", -1)
		searchIn = strings.Replace(searchIn, "\t", "", -1)
		if strings.Contains(searchIn, search) {
			return true
		}
	}

	return false
}

func assertStringFindInArray(search string, searchIn []string, acceptableLevel int) int {
	for index, s := range searchIn {
		if assertStringFind(search, s, acceptableLevel) {
			return index
		}
	}
	return -1
}

func assertFileExists(t *testing.T, file string) {
	t.Helper()
	file, err := filepath.Abs(file)
	if err != nil {
		t.Errorf("Error while trying to get the absolute path, got '%v'", err)
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Errorf("Expected file '%s' exists, but got '%v'", file, err)
	}
}

func assertFileContent(t *testing.T, file string, expectedContent string, acceptableLevel int) {
	t.Helper()
	if content, err := os.ReadFile(file); err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	} else {
		fileSlice := strings.Split(string(content), "\n")
		expectedSlice := strings.Split(expectedContent, "\n")
		// we want to check anything from the expectations is in the file
		// but we need to make sure if we also have this in order
		lastHit := -1
		for _, expected := range expectedSlice {
			hitAtIndex := assertStringFindInArray(expected, fileSlice, acceptableLevel)
			if hitAtIndex == -1 {
				t.Errorf("Expected file '%s' should contains '%s' what seems not be the case", file, expected)
			}
			if hitAtIndex < lastHit {
				t.Errorf("Expected file '%s' contains '%s' but not in the right order", file, expected)
			}
			// remove the hit from the file slice, so we can check if we have duplicates.
			// this is also nessary to check if we have the same line multiple times and do
			// not fail because we found it on the wrong index
			if hitAtIndex != -1 {
				systools.RemoveFromSliceOnce(fileSlice, fileSlice[hitAtIndex])
			}
			lastHit = hitAtIndex
		}
	}
}

type find_flags int

const (
	FindFlagsNone       find_flags = iota
	IgnoreTabs                     // ignore all tabs in the content and the message
	IgnoreSpaces                   // ignore all spaces in the content and the message
	IgnoreNewLines                 // ignore all new lines in the content and the message
	IgnoreMultiSpaces              // ignore all repeated spaces and tabs in the content and the message
	IgnoreCaseSensitive            // ignore case sensitive
)

// assert a string is part of the output buffer where we can ignore some flags
// like tabs, spaces, new lines, case sensitive
func assertInContent(t *testing.T, content string, msg string, flags ...find_flags) {
	t.Helper()

	if len(flags) > 0 {
		for _, flag := range flags {
			switch flag {
			case IgnoreTabs:
				content = strings.ReplaceAll(content, "\t", "")
				msg = strings.ReplaceAll(msg, "\t", "")
			case IgnoreSpaces:
				content = strings.ReplaceAll(content, " ", "")
				msg = strings.ReplaceAll(msg, " ", "")
			case IgnoreNewLines:
				content = strings.ReplaceAll(content, "\n", "")
				msg = strings.ReplaceAll(msg, "\n", "")
			case IgnoreMultiSpaces:
				content = systools.TrimAllSpaces(content)
				msg = systools.TrimAllSpaces(msg)
			case IgnoreCaseSensitive:
				content = strings.ToLower(content)
				msg = strings.ToLower(msg)
			}
		}
	}
	if !strings.Contains(content, msg) {
		t.Errorf("Expected to find '%s' in string. but did not found", msg)
	}
}

func assertStringSliceInContent(t *testing.T, content string, msg []string, flags ...find_flags) {
	t.Helper()
	for _, line := range msg {
		assertInContent(t, content, line, flags...)
	}
}
