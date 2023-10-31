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

package tasks

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// windows Get-CimInstance Win32_Process | Where-Object { $_.ParentProcessId -eq $ppid } | Select-Object ProcessId
func FindChildPid(pid int) ([]int, error) {
	var procList []int
	// only linux is supported
	// but others should not fail

	shellCmd := "ps -o pid --no-headers --ppid %d"

	if runtime.GOOS == "windows" {
		shellCmd = "Get-CimInstance Win32_Process | Where-Object { $_.ParentProcessId -eq %d } | Select-Object ProcessId"
	}
	command := fmt.Sprintf(shellCmd, pid)
	runner := GetShellRunner()
	// internalExitCode, processExitCode, err
	intCode, realCode, err := runner.Exec(
		command,
		func(msg string, err error) bool {
			if err == nil {
				var code int
				if _, nerr := fmt.Sscanf(msg, "%d", &code); nerr == nil {
					procList = append(procList, code)
				}
			}
			return true
		},
		func(proc *os.Process) {

		},
	)
	if err != nil {
		return procList, err
	}
	if intCode != 0 {
		return procList, fmt.Errorf("command %s failed with internal exit code %d", command, intCode)
	}
	if realCode != 0 {
		return procList, fmt.Errorf("command %s failed with exit code %d", command, realCode)
	}
	return procList, nil
}

func HandleAllMyPid(handlePid func(pid int) error) error {
	//mainProcs, err := FindChildPid(os.Getppid())
	procs, err := FindChildPid(os.Getpid())
	if err != nil {
		return err
	}

	for _, pid := range procs {
		if err := handlePid(pid); err != nil {
			return err
		}
	}
	return nil

}

// WinProcData is a struct to hold the data of a process on windows
type WinProcData struct {
	Pid         int
	ProcName    string
	SessionName string
	SessionId   int
	MemUsage    string
	Status      string
	UserName    string
	CpuTime     string
	WindowTitle string
}

// parsing the csv output of tasklist comming from windows tasklist command
func WinCsvToProcData(csv string) ([]WinProcData, error) {
	var procList []WinProcData

	parts := strings.Split(csv, "\n")
	for _, line := range parts {
		if strings.HasPrefix(line, "\"") {
			line = strings.Trim(line, "\"")
			parts := strings.Split(line, "\",\"")
			if len(parts) == 9 {
				var proc WinProcData
				if _, err := fmt.Sscanf(parts[1], "%d", &proc.Pid); err != nil {
					return procList, err
				}
				proc.ProcName = parts[0]
				proc.SessionName = parts[2]
				if _, err := fmt.Sscanf(parts[3], "%d", &proc.SessionId); err != nil {
					return procList, err
				}
				proc.MemUsage = parts[4]
				proc.Status = parts[5]
				proc.UserName = parts[6]
				proc.CpuTime = parts[7]
				proc.WindowTitle = parts[8]
				procList = append(procList, proc)
			}
		}
	}
	return procList, nil
}

// Getting process info on windows is a bit tricky
// we use the tasklist command to get the info
// and parse the csv output
func WinProcInfo(pid int) (WinProcData, error) {
	// TASKLIST /V /FO "CSV" /NH /FI "PID eq 2656"
	cmd := exec.Command("TASKLIST", "/V", "/FO", "CSV", "/NH", "/FI", fmt.Sprintf("PID eq %d", pid))
	result, err := cmd.Output()
	if err != nil {
		return WinProcData{}, err
	}
	procInfo := string(result)
	procList, err := WinCsvToProcData(procInfo)
	if err != nil {
		return WinProcData{}, err
	}
	if len(procList) == 0 {
		return WinProcData{}, fmt.Errorf("no process found")
	}
	return procList[0], nil
}
