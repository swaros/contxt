package tasks

import (
	"runtime"
	"strings"
)

var (
	// ShellCmd is the command to execute shell commands
	ShellCmd shellCmd = shellCmd{}
)

type shellCmd struct{}

func (s shellCmd) GetMainCmd() (string, []string) {
	lwr := strings.ToLower(runtime.GOOS)

	switch lwr {
	case "darwin":
		return "bash", []string{"-c"}
	case "freebsd":
		return "bash", []string{"-c"}
	case "netbsd":
		return "bash", []string{"-c"}
	case "openbsd":
		return "bash", []string{"-c"}
	case "plan9":
		return "rc", []string{}
	case "solaris":
		return "bash", []string{"-c"}
	case "windows":
		return "powershell", []string{"-nologo", "-noprofile"}

	}
	// fallback is bash. This is also the default for linux
	return "bash", []string{"-c"}
}
