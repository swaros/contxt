package taskrun

import "os"

const (
	PWRSHELL_CMD_VERSION = "$PSVersionTable.PSVersion.Major" // powershell cmd to get actual version
	PWRSHELL_CMD_PROFILE = "$PROFILE"                        // powershell cmd to get actual profile
)

func PwrShellExec(cmd string) string {
	cmdArg := []string{"-nologo", "-noprofile"} // these the arguments for powrshell
	result := ""
	ExecuteScriptLine(GetDefaultCmd(), cmdArg, cmd, func(s string, e error) bool {
		result = s
		return true
	}, func(p *os.Process) {

	})
	return result
}
