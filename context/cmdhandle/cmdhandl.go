package cmdhandle

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/dirhandle"
	"github.com/swaros/contxt/context/systools"
)

const (
	// DefaultExecFile is the filename of the script defaut file
	DefaultExecFile = "/.context.json"

	// TargetScript is script default target
	TargetScript = "script"

	// InitScript is script default target
	InitScript = "init"

	// ClearScript is script default target
	ClearScript = "clear"

	// TestScript is teh test target
	TestScript = "test"

	// DefaultCommandFallBack is used if no command is defined
	DefaultCommandFallBack = "bash"
)

// Shellout executes a command by using defined shell
func Shellout(ShellToUse string, command string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// ShelloutTrace executes a command by using defined shell
func ShelloutTrace(ShellToUse string, command string, callback func(string) bool) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	stdoutPipe, _ := cmd.StdoutPipe()
	err := cmd.Start()
	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		keepRunning := callback(m)
		if keepRunning == false {
			cmd.Process.Kill()
			return "", "", err
		}

	}
	cmd.Wait()

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	return stdout.String(), stderr.String(), err
}

// WriteTemplate create path based execution file
func WriteTemplate() {
	dir, error := dirhandle.Current()
	if error != nil {
		fmt.Println("error getting current directory")
		return
	}
	var template configure.ExecuteDefinition
	var dummyTest configure.CommandLine

	dummyTest.Command = "bash"
	dummyTest.Params = "echo 'Hallo Welt'"
	dummyTest.Comment = "just a example about the structure of a command"
	dummyTest.StopOnError = true

	dummyTest.Require.FileExists = make([]string, 0)
	dummyTest.Require.FileNotExists = make([]string, 0)

	template.Script = append(template.Script, dummyTest)
	template.TestScript = make([]configure.CommandLine, 0)
	template.InitScript = make([]configure.CommandLine, 0)
	template.CleanScript = make([]configure.CommandLine, 0)

	var path = dir + DefaultExecFile

	already, errEx := dirhandle.Exists(path)
	if errEx != nil {
		fmt.Println("error ", errEx)
		return
	}

	if already {
		fmt.Println("file already exists. aborted")
		return
	}

	fmt.Println("write execution template to ", path)

	b, _ := json.MarshalIndent(template, "", " ")
	err := ioutil.WriteFile(path, b, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

// ExecCurrentPathTemplate looks for template in current folder and executes them if exists
func ExecCurrentPathTemplate(target string) {
	dir, error := dirhandle.Current()
	if error != nil {
		fmt.Println("error getting current directory")
		return
	}
	var path = dir + DefaultExecFile

	already, errEx := dirhandle.Exists(path)
	if errEx != nil {
		fmt.Println("error ", errEx)
		return
	}
	if already {
		ExecPathFile(path, target)
	}
}

// ExecPathFile executes the default exec file
func ExecPathFile(path string, target string) {
	existing, fileerror := dirhandle.Exists(path)
	if fileerror != nil {
		fmt.Println("filecheck error: ", fileerror)
		return
	}

	if existing {
		fmt.Println(systools.Purple("exec "), systools.Teal(target), systools.White(path))
		file, _ := os.Open(path)
		defer file.Close()
		decoder := json.NewDecoder(file)
		var template configure.ExecuteDefinition
		err := decoder.Decode(&template)
		if err != nil {
			fmt.Println("error:", err)
		}
		execTemplate(template, target)
	}
}

func execTemplate(template configure.ExecuteDefinition, target string) bool {
	var result = true
	switch target {
	case TargetScript:
		result = runCommand(template.Script)
	case TestScript:
		result = runCommand(template.TestScript)
	case InitScript:
		result = runCommand(template.InitScript)
	case ClearScript:
		result = runCommand(template.CleanScript)

	}
	return result
}

func checkRequirements(command configure.CommandLine) bool {
	for _, fileExists := range command.Require.FileExists {
		fexists, err := dirhandle.Exists(fileExists)
		if err != nil || fexists == false {
			fmt.Println("required file (", fileExists, ") not found ")
			return false
		}
	}

	for _, fileNotExists := range command.Require.FileNotExists {
		fexists, err := dirhandle.Exists(fileNotExists)
		if err != nil || fexists == true {
			fmt.Println("unexpected file (", fileNotExists, ")  found ")
			return false
		}
	}

	return true
}

func testResultForReasonToStop(out string, command configure.CommandLine) bool {
	var stopReason = false
	if command.StopOnOutCountLess > 0 && command.StopOnOutCountLess > len(out) {
		fmt.Println("stopped because output len (", len(out), ") is less then ", command.StopOnOutCountLess)
		stopReason = true
	}
	if command.StopOnOutCountMore > 0 && command.StopOnOutCountMore < len(out) {

		fmt.Println("stopped because output len (", len(out), ") is more then ", command.StopOnOutCountMore)
		stopReason = true
	}

	if command.StopOnOutContains != "" && strings.Contains(out, command.StopOnOutContains) {
		fmt.Println("stopped because output contains (", command.StopOnOutContains, ")")
		stopReason = true
	}

	return stopReason
}

func runCommand(commands []configure.CommandLine) bool {
	for _, command := range commands {

		if command.Command == "" {
			command.Command = DefaultCommandFallBack
		}

		// first check requirements
		reqOk := checkRequirements(command)
		var out, errout string
		var err error
		var stopReason = false
		if reqOk == false {
			return false
		}
		if command.TraceOutput {
			out, errout, err = ShelloutTrace(command.Command, command.Params, func(message string) bool {
				fmt.Println(message)
				stopReason = testResultForReasonToStop(message, command)
				if stopReason {
					return false
				}
				return true
			})
		} else {
			out, errout, err = Shellout(command.Command, command.Params)
		}

		if err != nil {
			if command.StopOnError {
				log.Fatal(err, errout)
			}

			fmt.Println("\t", systools.Red(" Error:"), systools.Yellow(errout))
		}
		stopReason = testResultForReasonToStop(out, command)
		if stopReason {
			fmt.Println(systools.Teal("execution stopped by rule"), systools.White(command.Command), systools.Purple(command.Params))
			return false
		}
		fmt.Printf("%s", out)
	}
	return true
}
