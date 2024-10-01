// Copyright (c) 2023 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// # Licensed under the MIT License
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
package tasks

import (
	"errors"
	"os"
	"runtime"
	"strings"

	"github.com/swaros/contxt/module/configure"
)

var (
	// ShellCmd is the command to execute shell commands
	// It is set to the default value for the current OS
	// If you want to use a different shell, you can change this value
	// before calling any of the tasks
	ShellCmd shellCmd = shellCmd{}
)

type shellCmd struct{}

// GetMainCmd returns the main command and the arguments to use
func (s shellCmd) GetMainCmd(cfg configure.Options) (string, []string) {
	cfg.Maincmd = strings.TrimSpace(cfg.Maincmd)
	// the case we have no main command and no main params
	if cfg.Maincmd == "" && len(cfg.Mainparams) == 0 {
		lwr := strings.ToLower(runtime.GOOS)
		switch lwr {
		case "darwin": // macos
			return "bash", []string{"-c"}
		case "freebsd": // freebsd
			return "bash", []string{"-c"}
		case "netbsd": // netbsd
			return "bash", []string{"-c"}
		case "openbsd": // openbsd
			return "bash", []string{"-c"}
		case "plan9": // plan9
			return "rc", []string{}
		case "solaris": // solaris
			return "bash", []string{"-c"}
		case "windows": // windows
			return "powershell", []string{"-nologo", "-noprofile"}

		}
		// fallback is bash. This is also the default for linux
		return "bash", []string{"-c"}
	}
	// the case we have a main command and no main params
	if cfg.Maincmd != "" {
		return cfg.Maincmd, s.GetArgsForCmd(cfg.Maincmd)
	}
	// if anything is set already, we just return it
	return cfg.Maincmd, cfg.Mainparams

}

// try to get the arguments for the most common shells
func (s shellCmd) GetArgsForCmd(cmd string) []string {
	switch cmd {
	case "bash":
		return []string{"-c"}
	case "rc":
		return []string{}
	case "powershell":
		return []string{"-nologo", "-noprofile"}

	}
	return []string{}
}

func getCmd(forOs string) (string, []string, error) {
	lwr := strings.ToLower(forOs)
	switch lwr {
	case "linux": // linux
		return "bash", []string{"-c"}, nil

	case "darwin": // macos
		return "bash", []string{"-c"}, nil
	case "freebsd": // freebsd
		return "bash", []string{"-c"}, nil
	case "netbsd": // netbsd
		return "bash", []string{"-c"}, nil
	case "openbsd": // openbsd
		return "bash", []string{"-c"}, nil
	case "plan9": // plan9
		return "rc", []string{}, nil
	case "solaris": // solaris
		return "bash", []string{"-c"}, nil
	case "windows": // windows
		return "powershell", []string{"-nologo", "-noprofile"}, nil

	}

	return "", nil, errors.New("could not detect shell")
}
func detectCmd() (string, []string, error) {
	return getCmd(runtime.GOOS)
}

type shellRunner struct {
	cmd  string
	args []string
}

func GetShellRunner() *shellRunner {
	if shell, args, err := detectCmd(); err == nil {
		return &shellRunner{shell, args}
	} else {
		panic(err)
	}
}

func GetShellRunnerForOs(os string) *shellRunner {
	if shell, args, err := getCmd(os); err == nil {
		return &shellRunner{shell, args}
	} else {
		panic(err)
	}
}

// Exec executes the given command and calls the callback for each line of output
// If the callback returns false, the execution is stopped
func (s *shellRunner) Exec(command string, callback func(string, error) bool, startInfo func(*os.Process)) (int, int, error) {
	return Execute(s.cmd, s.args, command, callback, startInfo)
}

func (s *shellRunner) ExecSilentAndReturnLast(command string) (string, int) {
	last := ""
	_, code, _ := Execute(s.cmd, s.args, command, func(s string, err error) bool {
		last = s
		return true
	}, func(p *os.Process) {})
	return last, code
}

func (s *shellRunner) GetCmd() string {
	return s.cmd
}

func (s *shellRunner) GetArgs() []string {
	return s.args
}
