package taskrun

import "os"

const (
	PWRSHELL_CMD_VERSION        = "$PSVersionTable.PSVersion.Major"                                      // powershell cmd to get actual version
	PWRSHELL_CMD_PROFILE        = "$PROFILE"                                                             // powershell cmd to get actual profile
	PWRSHELL_CMD_TEST_PROFILE   = `Test-Path -Path $PROFILE.CurrentUserCurrentHost`                      // powershell cmd to test if profile exists
	PWRSHELL_CMD_PROFILE_CREATE = `New-Item -Path $PROFILE.CurrentUserCurrentHost -ItemType File -Force` // powershell cmd to create profile
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
