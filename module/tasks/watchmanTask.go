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
	"syscall"
)

// TaskDef holds information about running
// and finished tasks
type TaskDef struct {
	uuid       string
	started    bool
	count      int
	done       bool
	doneCount  int
	process    *ProcessDef
	processLog []ProcessLog
}

type ProcessDef struct {
	handlingDone bool // task wise done. did not mean it is not running anymore
	processInfo  *os.Process
}

type ProcessLog struct {
	Cmd     string
	Args    []string
	Command string
	Pid     int
}

func (ts *TaskDef) StartTrackProcess(proc *os.Process) {
	ts.process = &ProcessDef{
		handlingDone: false,
		processInfo:  proc,
	}
}

func (ts *TaskDef) GetProcess() *ProcessDef {
	return ts.process
}

func (ts *TaskDef) GetProcessPid() (int, bool) {
	if ts.process != nil && ts.process.processInfo != nil {
		return ts.process.processInfo.Pid, true
	}
	return 0, false
}

func (ts *TaskDef) LogCmd(cmd string, args []string, command string) {
	pid := 0
	if ts.process != nil && ts.process.processInfo != nil {
		pid = ts.process.processInfo.Pid
	}
	ts.processLog = append(ts.processLog, ProcessLog{
		Cmd:     cmd,
		Args:    args,
		Pid:     pid,
		Command: command,
	})
}

func PidExists(pid int32) (bool, error) {
	if pid <= 0 {
		return false, fmt.Errorf("invalid pid %v", pid)
	}
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return false, err
	}
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true, nil
	}
	if err.Error() == "os: process already finished" {
		return false, nil
	}
	errno, ok := err.(syscall.Errno)
	if !ok {
		return false, err
	}
	switch errno {
	case syscall.ESRCH:
		return false, nil
	case syscall.EPERM:
		return true, nil
	}
	return false, err
}

func (ts *TaskDef) IsProcessRunning() bool {
	if ts.process != nil && ts.process.processInfo != nil {
		if ts.process.processInfo.Pid > 0 {
			proc, err := os.FindProcess(ts.process.processInfo.Pid)
			if runtime.GOOS == "windows" {
				_, pErr := WinProcInfo(ts.process.processInfo.Pid)
				return pErr == nil
			}
			if err == nil {
				if err := proc.Signal(syscall.Signal(0)); err != nil {
					return false
				}
				return true
			}
		}
	}
	return false
}

func (ts *TaskDef) GetProcessLog() []ProcessLog {
	return ts.processLog
}

func (ts *TaskDef) KillProcess() error {
	if ts.process != nil && ts.process.processInfo != nil {
		if ts.IsProcessRunning() {
			return ts.process.processInfo.Kill()
		} else {
			return fmt.Errorf("process %d is not running", ts.process.processInfo.Pid)
		}
	}
	return fmt.Errorf("no process to kill")
}

// StopProcessIfRunning sends an interrupt signal to the process
// if the process is running
// if the process is not running, nothing happens. will not reported as error
func (ts *TaskDef) StopProcessIfRunning() error {
	if ts.process != nil && ts.process.processInfo != nil {
		if ts.IsProcessRunning() {
			return ts.process.processInfo.Signal(os.Interrupt)
		}
	} else {
		fmt.Println("no process to stop")
	}

	return nil
}
