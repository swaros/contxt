package cmdhandle

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"

	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/dirhandle"
	"github.com/swaros/contxt/context/systools"
	"gopkg.in/yaml.v2"
)

const (
	// DefaultExecFile is the filename of the script defaut file
	DefaultExecFile = "/.context.json"

	// DefaultExecYaml is the default yaml configuration file
	DefaultExecYaml = "/.contxt.yml"

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

// ExecuteTemplateWorker runs ExecCurrentPathTemplate in context of a waitgroup
func ExecuteTemplateWorker(waitGroup *sync.WaitGroup, path string) {
	defer waitGroup.Done()
	ExecCurrentPathTemplate(path)
}

// ExecuteWorker runs ExecuteScriptLine in context of a waitgroup
func ExecuteWorker(waitGroup *sync.WaitGroup, ShellToUse string, command string, callback func(string) bool, startInfo func(*os.Process)) {
	defer waitGroup.Done()
	ExecuteScriptLine(ShellToUse, command, callback, startInfo)
}

// ExecuteScriptLine executes a simple shell script
func ExecuteScriptLine(ShellToUse string, command string, callback func(string) bool, startInfo func(*os.Process)) error {
	cmd := exec.Command(ShellToUse, "-c", command)
	stdoutPipe, _ := cmd.StdoutPipe()
	err := cmd.Start()
	startInfo(cmd.Process)
	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		keepRunning := callback(m)
		if keepRunning == false {
			cmd.Process.Kill()
			return err
		}

	}
	cmd.Wait()
	return err
}

// WriteTemplate create path based execution file
func WriteTemplate() {
	dir, error := dirhandle.Current()
	if error != nil {
		fmt.Println("error getting current directory")
		return
	}

	var path = dir + DefaultExecYaml
	already, errEx := dirhandle.Exists(path)
	if errEx != nil {
		fmt.Println("error ", errEx)
		return
	}

	if already {
		fmt.Println("file already exists.", path, " aborted")
		return
	}

	fmt.Println("write execution template to ", path)

	var demoContent = "task:\n    - id: script\n      script:\n        - echo 'hallo welt'\n        - ls -ga"
	err := ioutil.WriteFile(path, []byte(demoContent), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

// ExecCurrentPathTemplate looks for template in current folder and executes them if exists
func ExecCurrentPathTemplate(target string) {
	execCurrentYaml(target)
}

func execCurrentYaml(target string) {
	dir, error := dirhandle.Current()
	if error != nil {
		fmt.Println("error getting current directory")
		return
	}
	var path = dir + DefaultExecYaml
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
		file, ferr := ioutil.ReadFile(path)
		if ferr != nil {
			fmt.Println("file loading error: ", fileerror)
		}
		var template configure.RunConfig
		err := yaml.Unmarshal(file, &template)

		if err != nil {
			fmt.Println("error:", err)
		} else {
			executeTemplate(template, target)
		}
	}
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
