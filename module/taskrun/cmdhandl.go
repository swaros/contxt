// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Licensed under the MIT License
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package taskrun

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/configure"
	"github.com/swaros/contxt/dirhandle"
)

const (
	// DefaultExecFile is the filename of the script defaut file
	DefaultExecFile = string(os.PathSeparator) + ".context.json"

	// DefaultExecYaml is the default yaml configuration file
	DefaultExecYaml     = string(os.PathSeparator) + ".contxt.yml"
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

	// On windows we have a different default
	DefaultCommandFallBackWindows = "powershell"
)

// ExecuteTemplateWorker runs ExecCurrentPathTemplate in context of a waitgroup
func ExecuteTemplateWorker(waitGroup *sync.WaitGroup, useWaitGroup bool, target string, template configure.RunConfig) {
	if useWaitGroup {
		defer waitGroup.Done()
	}
	//ExecCurrentPathTemplate(path)
	exitCode := ExecPathFile(waitGroup, useWaitGroup, template, target)
	GetLogger().WithField("exitcode", exitCode).Info("ExecuteTemplateWorker done with exitcode")

}

func GetExecDefaults() (string, []string) {
	cmd := GetDefaultCmd()
	var args []string
	args = GetDefaultCmdOpts(cmd, args)
	return cmd, args
}

func GetDefaultCmd() string {

	envCmd := os.Getenv("CTX_DEFAULT_CMD")
	if envCmd != "" {
		GetLogger().WithField("defaultcmd", envCmd).Info("Got default cmd from environment")
		return envCmd
	}

	if configure.GetOs() == "windows" {
		return DefaultCommandFallBackWindows
	}
	return DefaultCommandFallBack
}

func GetDefaultCmdOpts(ShellToUse string, cmdArg []string) []string {
	if configure.GetOs() == "windows" {
		if envCmd := os.Getenv("CTX_DEFAULT_CMD_ARGUMENTS"); envCmd != "" {
			GetLogger().WithField("arguments", envCmd).Info("Got cmd arguments form environment")
			return strings.Split(envCmd, " ")
		}
		if cmdArg == nil && ShellToUse == DefaultCommandFallBackWindows {
			cmdArg = []string{"-nologo", "-noprofile"}
		}
	} else {
		if cmdArg == nil && ShellToUse == DefaultCommandFallBack {
			cmdArg = []string{"-c"}
		}
	}
	return cmdArg
}

// ExecuteScriptLine executes a simple shell script
// returns internal exitsCode, process existcode, error
func ExecuteScriptLine(ShellToUse string, cmdArg []string, command string, callback func(string) bool, startInfo func(*os.Process)) (int, int, error) {
	cmdArg = GetDefaultCmdOpts(ShellToUse, cmdArg)
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
	}).Info(":::EXEC")

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
		if !keepRunning {
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
      - echo "hello world"
`
	err := ioutil.WriteFile(path, []byte(demoContent), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

// ExecPathFile executes the default exec file
func ExecPathFile(waitGroup *sync.WaitGroup, useWaitGroup bool, template configure.RunConfig, target string) int {
	var scopeVars map[string]string = make(map[string]string)
	return executeTemplate(waitGroup, useWaitGroup, template, target, scopeVars)
}
