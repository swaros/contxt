// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

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
