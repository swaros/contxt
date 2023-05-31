package runner_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/runner"
	"github.com/swaros/contxt/module/systools"
)

var useLastDir = "./"
var lastExistCode = 0
var testDirectory = ""

// this are some helper functions especially for testing the runner

// Setup the test app
// create the application. set up the config folder name, and the name of the config file.
// the testapp bevavior is afterwards different, because it uses the config
// related to the current directory.
// thats why we have some special helper functions.
// - getAbsolutePath to get the absolute path to the testdata directory
// - backToWorkDir to go back to the testdata directory
// - cleanAllFiles to remove the config file
func SetupTestApp(dir, file string) (*runner.CmdSession, *TestOutHandler, error) {

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

		return nil, nil, err
	}
	// check if the directory exists, that we want to use in the testdata directory.
	// even if the config package is abel to create them, we want avoid this here.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil, err
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
		return nil, nil, derr
	}

	app := runner.NewCmdSession()

	functions := runner.NewCmd(app)

	ctxout.AddPostFilter(ctxout.NewTabOut())

	if err := app.Cobra.Init(functions); err != nil {
		return nil, nil, err
	}

	outputHdnl := NewTestOutHandler()
	app.OutPutHdnl = outputHdnl
	return app, outputHdnl, nil
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
	return useLastDir + "/" + dir
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

// assert a string is part of the output buffer
func assertInMessage(t *testing.T, output *TestOutHandler, msg string) {
	t.Helper()
	if !output.Contains(msg) {
		t.Errorf("Expected '%s', got '%v'", msg, output.String())
	}
}

// assert a string is not part of the output buffer
func assertNotInMessage(t *testing.T, output *TestOutHandler, msg string) {
	t.Helper()
	if output.Contains(msg) {
		t.Errorf("Expected '%s' is not in the message, but got '%v'", msg, output.String())
	}
}
