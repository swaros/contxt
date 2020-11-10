package cmdhandle

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/dirhandle"
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
func ExecuteTemplateWorker(waitGroup *sync.WaitGroup, useWaitGroup bool, target string, template configure.RunConfig) {
	if useWaitGroup {
		defer waitGroup.Done()
	}
	//ExecCurrentPathTemplate(path)
	ExecPathFile(waitGroup, useWaitGroup, template, target)

}

// ExecuteScriptLine executes a simple shell script
func ExecuteScriptLine(ShellToUse string, cmdArg []string, command string, callback func(string) bool, startInfo func(*os.Process)) (int, int, error) {

	// default behavior. -c param is not set by default
	if cmdArg == nil && ShellToUse == DefaultCommandFallBack {
		cmdArg = []string{"-c"}
	}

	cmdArg = append(cmdArg, command)
	cmd := exec.Command(ShellToUse, cmdArg...)

	stdoutPipe, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	err := cmd.Start()
	if err != nil {
		GetLogger().Warn("execution error: ", err)
		return ExitCmdError, 0, err
	}

	GetLogger().WithFields(logrus.Fields{
		"env":        cmd.Env,
		"args":       cmd.Args,
		"dir":        cmd.Dir,
		"extrafiles": cmd.ExtraFiles,
		"pid":        cmd.Process.Pid,
	}).Info("command")

	startInfo(cmd.Process)
	scanner := bufio.NewScanner(stdoutPipe)

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()

		keepRunning := callback(m)

		GetLogger().WithFields(logrus.Fields{
			"keep-running": keepRunning,
			"out":          m,
		}).Info("handle-result")
		if keepRunning == false {
			cmd.Process.Kill()
			return ExitByStopReason, 0, err
		}

	}
	err = cmd.Wait()
	if err != nil {
		errRealCode := 0
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				GetLogger().Warn("(maybe expected...) Exit Status reported: ", status.ExitStatus())
				errRealCode = status.ExitStatus()
			}

		} else {
			GetLogger().Warn("execution error: ", err)
		}
		return ExitCmdError, errRealCode, err
	}

	return ExitOk, 0, err
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

	var demoContent = `task:
  - id: script
    script:
      - echo 'hallo welt'
      - ls -ga`
	err := ioutil.WriteFile(path, []byte(demoContent), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

// ExecPathFile executes the default exec file
func ExecPathFile(waitGroup *sync.WaitGroup, useWaitGroup bool, template configure.RunConfig, target string) {
	executeTemplate(waitGroup, useWaitGroup, template, target)
}
