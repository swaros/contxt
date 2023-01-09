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
