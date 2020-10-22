package cmdhandle

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/swaros/contxt/context/output"

	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/dirhandle"
	"gopkg.in/yaml.v2"
)

const (
	// DefaultExecFile is the filename of the script defaut file
	DefaultExecFile = "/.context.json"

	// DefaultExecYaml is the default yaml configuration file
	DefaultExecYaml     = "/.contxt.yml"
	defaultExecYamlName = ".contxt.yml"

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
func ExecuteTemplateWorker(waitGroup *sync.WaitGroup, useWaitGroup bool, target string, templatePath string) {
	if useWaitGroup {
		defer waitGroup.Done()
	}
	//ExecCurrentPathTemplate(path)
	ExecPathFile(waitGroup, useWaitGroup, templatePath, target)

}

// ExecuteScriptLine executes a simple shell script
func ExecuteScriptLine(ShellToUse string, command string, callback func(string) bool, startInfo func(*os.Process)) (int, error) {

	cmd := exec.Command(ShellToUse, "-c", command)

	stdoutPipe, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	err := cmd.Start()
	if err != nil {
		return ExitCmdError, err
	}

	startInfo(cmd.Process)
	scanner := bufio.NewScanner(stdoutPipe)

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		keepRunning := callback(m)
		if keepRunning == false {
			cmd.Process.Kill()
			return ExitByStopReason, err
		}

	}
	err = cmd.Wait()
	if err != nil {

		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				log.Printf("Exit Status reported: %d", status.ExitStatus())
			}

		} else {
			log.Printf("execution error: %v", err)
		}
		return ExitCmdError, err
	}

	return ExitOk, err
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

// ExecPathFile executes the default exec file
func ExecPathFile(waitGroup *sync.WaitGroup, useWaitGroup bool, path string, target string) {
	existing, fileerror := dirhandle.Exists(path)
	if fileerror != nil {
		fmt.Println("filecheck error: ", fileerror)
		return
	}

	if existing {
		fmt.Println(output.MessageCln(output.ForeBlue, "[exec] ", output.BoldTag, target, output.ResetBold, " ", output.ForeWhite, path))
		file, ferr := ioutil.ReadFile(path)
		if ferr != nil {
			fmt.Println("file loading error: ", fileerror)
		}
		var template configure.RunConfig
		err := yaml.Unmarshal(file, &template)

		if err != nil {
			fmt.Println("error:", err)
		} else {
			executeTemplate(waitGroup, useWaitGroup, template, target)
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
