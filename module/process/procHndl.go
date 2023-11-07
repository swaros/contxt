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

package process

import (
	"errors"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/swaros/contxt/module/mimiclog"
)

// basic struct to hold the data of a process
type ProcData struct {
	Pid         int         // process id
	Cmd         string      // command line
	ThreadCount int         // number of threads
	Threads     []int       // list of threads pids
	Childs      []int       // list of child pids
	ChildProcs  []*ProcData // list of child processes
}

type ProcessWatch struct {
	pData       *ProcData
	processInfo *os.Process
	stopLock    sync.Mutex
	logger      mimiclog.Logger
}

// ReadProc reads the process data of a process with the given pid
// and returns a ProcData struct
func NewProc(pid int) (*ProcData, error) {
	return ReadProc(pid)
}

// NewProcessWatcherByCmd creates a new ProcessDef struct
// and returns a pointer to it
// the ProcessDef struct holds the process data of the given process
func NewProcessWatcherByCmd(cmd *exec.Cmd) (*ProcessWatch, error) {
	if pdef, err := NewProc(cmd.Process.Pid); err != nil {
		return nil, err
	} else {
		return &ProcessWatch{
			pData:       pdef,
			processInfo: cmd.Process,
			logger:      mimiclog.NewNullLogger(),
		}, nil
	}
}

func NewProcessWatcherByProcessInfo(proc *os.Process) (*ProcessWatch, error) {
	if proc == nil {
		return nil, errors.New("NewProcessWatcherByProcessInfo: process is nil")
	}
	if proc.Pid == 0 {
		return nil, errors.New("NewProcessWatcherByProcessInfo: process pid is 0")
	}
	if pdef, err := NewProc(proc.Pid); err != nil {
		return nil, err
	} else {
		return &ProcessWatch{
			pData:       pdef,
			processInfo: proc,
			logger:      mimiclog.NewNullLogger(),
		}, nil
	}
}

func NewProcessWatcherByPid(pid int) (*ProcessWatch, error) {
	if pid == 0 {
		return nil, errors.New("NewProcessWatcherByPid: can not handle pid = 0")
	}
	if pdef, err := os.FindProcess(pid); err != nil {
		return nil, err
	} else {
		return NewProcessWatcherByProcessInfo(pdef)
	}
}

func (proc *ProcessWatch) SetLogger(logger mimiclog.Logger) {
	proc.logger = logger
}

func (proc *ProcessWatch) GetPid() int {
	return proc.pData.Pid
}

func (proc *ProcessWatch) GetCmd() string {
	return proc.pData.Cmd
}

func (proc *ProcessWatch) GetThreadCount() int {
	return proc.pData.ThreadCount
}

func (proc *ProcessWatch) GetThreads() []int {
	return proc.pData.Threads
}

func (proc *ProcessWatch) GetChilds() []int {
	return proc.pData.Childs
}

func (proc *ProcessWatch) WalkChildProcs(startLevel int, f func(p *ProcData, parentPid int, level int) bool) {
	level := startLevel + 1
	for _, child := range proc.pData.ChildProcs {
		if proc.logger.IsTraceEnabled() {
			proc.logger.Trace("WalkChildProcs: ", child.Pid, " ", child.Cmd, " ", child.ThreadCount, " ", child.Threads, " ", child.Childs)
		}
		if f(child, proc.pData.Pid, level) {
			child.WalkChildProcs(level, f)
		}
	}

}

func (pd *ProcData) WalkChildProcs(startLevel int, f func(p *ProcData, parentPid int, level int) bool) {
	level := startLevel + 1
	for _, child := range pd.ChildProcs {
		if f(child, pd.Pid, level) {
			child.WalkChildProcs(level, f)
		}
	}
}

func (proc *ProcessWatch) GetProcessInfo() *os.Process {
	return proc.processInfo
}

func (proc *ProcessWatch) StopWithDefaultSigOrder() error {
	proc.logger.Debug("StopWithDefaultSigOrder: ", DefaultInterruptSignal, " ", DefaultKillSignal)
	return proc.StopChilds(DefaultInterruptSignal, DefaultKillSignal)
}

// ProcessWatch.StopChilds sends the given signals to the child processes
func (proc *ProcessWatch) StopChilds(signals ...Signal) error {
	proc.logger.Debug("StopChilds: ", signals)
	// if there are no child processes then just return
	// before locking the mutex
	if len(proc.pData.ChildProcs) == 0 {
		proc.logger.Debug("StopChilds: no child processes")
		return nil
	}
	proc.stopLock.Lock()
	defer proc.stopLock.Unlock()
	proc.Update() // update the process data
	for _, child := range proc.pData.ChildProcs {
		if err := child.Stop(signals...); err != nil {
			return err
		}
	}
	return nil
}

// ProcData.Stop sends the given signals to the child processes
// any of these child processes can have child processes
// they will be stopped too
func (pd *ProcData) StopChilds(signals ...Signal) error {
	for _, child := range pd.ChildProcs {
		if err := child.Stop(signals...); err != nil {
			return err
		}
	}
	return nil
}

// Stop sends the given signals to the child processes
// and then to the process itself
func (pd *ProcData) Stop(signals ...Signal) error {
	if err := pd.StopChilds(signals...); err != nil {
		return err
	}
	for _, signal := range signals {
		if processInfo, err := os.FindProcess(pd.Pid); err != nil {
			return err
		} else {
			if pd.isRunning() {
				if err := processInfo.Signal(signal.Signal); err != nil {
					return err
				}
				// now wait for the given time period
				if signal.ThenWait > 0 {
					time.Sleep(signal.ThenWait)
				}
			}
		}
	}
	return nil
}

// check if still running by requesting the process data again
// if an error occurs then the process is not running anymore
func (pd *ProcData) isRunning() bool {
	tmp, err := NewProc(pd.Pid)
	if err != nil {
		return false
	}
	return tmp.Pid > 0

}

func (proc *ProcessWatch) Kill() error {
	return KillProcessTree(proc.pData.Pid)
}

func (proc *ProcessWatch) Update() error {
	proc.logger.Debug("Update: ", proc.pData.Pid)
	if pdef, err := NewProc(proc.pData.Pid); err != nil {
		proc.logger.Debug("Update Error (maybe expected): ", err)
		return err
	} else {
		proc.logger.Trace("Update: updated process data")
		proc.pData = pdef
		return nil
	}
}

// IsRunning checks if the process is still running.
// this is done by requesting the process data again.
// if an error occurs then the process is not running anymore
// because the process data can not be read.
// that means that we will return false if an error occurs so this is
// you should check instead the error itself.
// the error is useful if you want to know why the process data can not be read.
// but again: for a process that is stopped, you will get always an error.
func (proc *ProcessWatch) IsRunning() (bool, error) {
	if err := proc.Update(); err != nil {
		return false, err
	}
	return proc.pData.Pid > 0, nil
}

func (proc *ProcessWatch) WaitForStop(timeout, waitTick time.Duration) (time.Duration, error) {
	proc.logger.Debug("WaitForStop: ", timeout)
	if timeout == 0 {
		return 0, errors.New("WaitForStop: timeout is 0")
	}
	start := time.Now()
	for {
		if running, _ := proc.IsRunning(); running {
			if time.Since(start) > timeout {
				return 0, errors.New("WaitForStop: timeout")
			}
		} else {
			return time.Since(start), nil
		}
		time.Sleep(waitTick)
	}
}

func (proc *ProcessWatch) WaitForStart(timeout, waitTick time.Duration) error {
	proc.logger.Debug("WaitForStart: ", timeout)
	if timeout == 0 {
		return errors.New("WaitForStart: timeout is 0")
	}
	start := time.Now()
	for {
		if running, _ := proc.IsRunning(); !running {
			if time.Since(start) > timeout {
				return errors.New("WaitForStart: timeout")
			}
		} else {
			return nil
		}
		time.Sleep(waitTick)
	}
}
