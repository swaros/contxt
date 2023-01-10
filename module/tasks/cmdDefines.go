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
	"runtime"
	"strings"
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
func (s shellCmd) GetMainCmd() (string, []string) {
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
