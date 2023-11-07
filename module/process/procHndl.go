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
	pData       *ProcData       // process data
	processInfo *os.Process     // process info
	stopLock    sync.Mutex      // mutex to lock the stop process
	logger      mimiclog.Logger // logger
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

// NewProcessWatcherByProcessInfo creates a new ProcessDef struct
// and returns a pointer to it
// same as NewProcessWatcherByCmd but with a os.Process struct as parameter
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

// NewProcessWatcherByPid creates a new ProcessDef struct
// and returns a pointer to it
// same as NewProcessWatcherByCmd but with a pid as parameter
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

// SetLogger sets the logger for the process
// if no logger is set, a null logger will be used
func (proc *ProcessWatch) SetLogger(logger mimiclog.Logger) {
	proc.logger = logger
}

// GetPid returns the pid of the process
func (proc *ProcessWatch) GetPid() int {
	return proc.pData.Pid
}

// GetCmd returns the command line of the process
func (proc *ProcessWatch) GetCmd() string {
	return proc.pData.Cmd
}

// GetThreadCount returns the number of threads of the process
// these are NOT the child processes. these are the threads of the process itself
func (proc *ProcessWatch) GetThreadCount() int {
	return proc.pData.ThreadCount
}

// GetThreads returns the list of PID's of threads of the process
func (proc *ProcessWatch) GetThreads() []int {
	return proc.pData.Threads
}

// GetChilds returns the list of PID's of child processes of the process
func (proc *ProcessWatch) GetChilds() []int {
	return proc.pData.Childs
}

// WalkChildProcs walks through the child processes of the process
// and calls the given function for each child process
// the function gets the child process data, the parent pid and the level as parameter
// the level is the level of the child process in the process tree
// the function returns a bool. if the bool is true, the child processes of the child process will be walked too
func (proc *ProcessWatch) WalkChildProcs(f func(p *ProcData, parentPid int, level int) bool) {
	level := 1
	for _, child := range proc.pData.ChildProcs {
		if proc.logger.IsTraceEnabled() {
			proc.logger.Trace("WalkChildProcs: ", child.Pid, " ", child.Cmd, " ", child.ThreadCount, " ", child.Threads, " ", child.Childs)
		}
		if f(child, proc.pData.Pid, level) {
			child.WalkChildProcs(level, f)
		}
	}

}

// WalkChildProcs walks through the child processes of the process. this is a recursive function
// and calls the given function for each child process.
// the function gets the child process data, the parent pid and the level as parameter
// the level is the level of the child process in the process tree.
// this is mostly used internally by calling WalkChildProcs from ProcessWatch. but can also be used to get
// any childs starting from a different level, if needed.
func (pd *ProcData) WalkChildProcs(startLevel int, f func(p *ProcData, parentPid int, level int) bool) {
	level := startLevel + 1
	for _, child := range pd.ChildProcs {
		if f(child, pd.Pid, level) {
			child.WalkChildProcs(level, f)
		}
	}
}

// GetProcessInfo returns the os.Process struct of the process
func (proc *ProcessWatch) GetProcessInfo() *os.Process {
	return proc.processInfo
}

// StopWithDefaultSigOrder sends the default signals to the child processes
// DefaultInterruptSignal and DefaultKillSignal
// this is the same as calling StopChilds(DefaultInterruptSignal, DefaultKillSignal)
// instead of just stopping the current process, we also taking care about the child processes.
// this way we can make sure that the process tree is stopped, and we do not have any zombie processes.
func (proc *ProcessWatch) StopWithDefaultSigOrder() error {
	proc.logger.Debug("StopWithDefaultSigOrder: ", DefaultInterruptSignal, " ", DefaultKillSignal)
	return proc.StopChilds(DefaultInterruptSignal, DefaultKillSignal)
}

// ProcessWatch.StopChilds sends the given signals to the child processes
// any of these child processes can have child processes.
// they will be stopped too.
// the signal order is important. you can use one of the default Signnals, the containing the regular signal
// and the ThenWait time, that is used to give the process time to stop.
// or you can send your own signals.
// like so:
//
//	proc.StopChilds(process.Signal{Signal: syscall.SIGINT, ThenWait: 1 * time.Second}, process.Signal{Signal: syscall.SIGKILL, ThenWait: 10 * time.Millisecond})
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
// and then to the process itself.
// any of these child processes can have child processes
// they will be stopped too
// the signal order is important. you can use one of the default Signnals, the containing the regular signal
// and the ThenWait time, that is used to give the process time to stop.
// or you can send your own signals.
// like so:
//
//	proc.Stop(process.Signal{Signal: syscall.SIGINT, ThenWait: 1 * time.Second}, process.Signal{Signal: syscall.SIGKILL, ThenWait: 10 * time.Millisecond})
//
// if the process is not running anymore, we will return nil
// this is also used by the ProcessWatch.Stop() function
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
// here we ignore the error and just return false, becaue we are only interested
// if the process is running or not. the error itself is not important.
// errors depending on the try to get get the process data again.
// and any of these failures depending on a process that is not running anymore.
// if you need to know what exactly happened, you can use the Update() function for checking
// and handle the error yourself.
func (pd *ProcData) isRunning() bool {
	tmp, err := NewProc(pd.Pid)
	if err != nil {
		return false
	}
	return tmp.Pid > 0

}

// Kill sends the kill signal to the process
// it uses the KillProcessTree function to kill the process tree.
// this way we can include some os specific code to kill the process tree.
func (proc *ProcessWatch) Kill() error {
	return KillProcessTree(proc.pData.Pid)
}

// Update updates the process data of the process.
// this is done by requesting the process data again.
// if an error occurs then the process is not running anymore. at least this is the usual case.
// because the process data can not be read.
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

// WaitForStop waits until the process is stopped. or until the timeout is reached.
// this is different to the usual Timeout function, because this will not count for the Timeout.
// this function is ment for use in cases, we just want to wait until the process is stopped, without forcing them being killed.
// if the (local) timeout is reached, we will return an error, but the process will still be running.
// this can be combined with the Timeout function, to force the process to stop after the timeout is reached.
// but then make sure to set the timeout to a higher value than the WaitForStop timeout.
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

// WaitForStart waits until the process is started. or until the timeout is reached.
// this is ment for use in cases, we just want to wait until the process is started, before we start working with them.
// here we do not check any internal flags or something like that. we just check if the process is running in the system.
// so this would also return true if the process is running, but not able to handle some inputs.
// this is depending on the process itself.
// for checking if the process handle inputs, you need to check the output of the process. (if the application outputs some text on start)
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
