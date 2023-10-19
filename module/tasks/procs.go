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
	"runtime"
)

func FindChildPid(pid int) ([]int, error) {
	var procList []int
	// only linux is supported
	// but others should not fail
	if runtime.GOOS != "linux" {
		return procList, nil
	}
	command := fmt.Sprintf("ps -o pid --no-headers --ppid %d", pid)
	runner := GetShellRunner()
	// internalExitCode, processExitCode, err
	intCode, realCode, err := runner.Exec(
		command,
		func(msg string, err error) bool {
			if err == nil {
				var code int
				fmt.Sscanf(msg, "%d", &code)
				procList = append(procList, code)
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
