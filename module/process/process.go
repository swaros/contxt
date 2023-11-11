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
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/process/terminal"
	"github.com/swaros/contxt/module/systools"
)

type ProcCallback func(string, error) bool // the callback to call when output is received
type ProcInfoCallback func(*os.Process)    // the callback to call when the process is started
type ProcHndlCallback func(error)          // a callback for the process handler
type ProcChildCntChangeCallback func(int)  // a callback for the process handler

type Process struct {
	cmd              string                     // the base command to run
	args             []string                   // the arguments to pass to the command
	startCommands    []string                   // the commands to send to the process on startup
	onOutput         ProcCallback               // the callback to call when output is received
	onInit           ProcInfoCallback           // the callback to call when the process is started
	onWaitDone       ProcHndlCallback           // the callback to call when the process is stopped after a regular wait for cmd execution
	outPipe          io.ReadCloser              // the pipe to read output from
	errorPipe        io.ReadCloser              // the pipe to read errors from
	inPipe           io.WriteCloser             // the pipe to write input to
	combinePipes     bool                       // whether or not to combine the output and error pipes
	stopped          bool                       // whether or not the process has been stopped
	started          bool                       // whether or not the process has been started
	cmdLock          sync.Mutex                 // a lock for the command
	wait             sync.WaitGroup             // a wait group for the process
	stayOpen         bool                       // whether or not the process should stay open. If it does, it will not be stopped after the process handles the startup commands
	timeOut          time.Duration              // the timeout for the process. If the process is not stopped after the timeout, it will be stopped
	timoutSet        bool                       // whether or not the timeout is set
	commandExitCode  int                        // the exit code of the process
	internalExitCode int                        // the internal exit code of the process
	runtimeError     error                      // the error of the process
	procWatch        *ProcessWatch              // the process watcher for the process
	pipesClosed      bool                       // whether or not the pipes are closed
	logger           mimiclog.Logger            // the logger to use
	reportChildCount bool                       // whether or not to report the child count
	onChildCntChange ProcChildCntChangeCallback // a callback for the process handler if the amount of processes changed
}

var (
	ExitInBackGround   = 601 // the exit code when the process is running in the background
	ExitTimeout        = 602 // the exit code when the process is stopped by timeout
	RealCodeNotPresent = -1  // the real exit code is not present because the process was killed or never started or timed out
)

// NewProcess creates a new process with the given command and arguments
func NewProcess(cmd string, args ...string) *Process {
	return &Process{
		cmd:    cmd,
		args:   args,
		logger: mimiclog.NewNullLogger(),
	}
}

// Create a new process by using the default terminal
// as defined in the terminal package
// if no arguments are given, is seems to be intended to be used
// for interactive processes
func NewTerminal(args ...string) *Process {
	term, err := terminal.GetTerminal()
	if err != nil {
		panic(err)
	}
	// no arguments given. seems to be an interactive process.
	// so we use the keep open arguments instead
	if len(args) == 0 {
		return NewProcess(term.GetCmd(), term.GetArgsToKeepOpen()...)
	}
	return NewProcess(term.GetCmd(), term.CombineArgs(args...)...)
}

// SetLogger sets the logger for the process
// Fullfilling the mimiclog.Logger interface
func (p *Process) SetLogger(logger mimiclog.Logger) {
	p.logger = logger
	if p.procWatch != nil {
		p.procWatch.SetLogger(logger)
	}
}

// GetLogger returns the logger for the process
func (p *Process) GetLogger() mimiclog.Logger {
	return p.logger
}

// enable reporting the child count if changed. this is mostly usefull for processes that stay open.
// and is ment togehther with the SetOnChildCountChange callback.
// but you can also use it without the callback just to get the child count by the logger output (debug level needed)
func (p *Process) SetReportChildCount(report bool) {
	p.reportChildCount = report
}

// SetOnChildCountChange sets the callback to call when the child count of the process changes.
// this is only usefull for processes that stay open.
func (p *Process) SetOnChildCountChange(callback ProcChildCntChangeCallback) {
	p.onChildCntChange = callback
	p.reportChildCount = callback != nil
}

// returns if the process is started
func (p *Process) IsStarted() bool {
	return p.started
}

// SetCombinePipes sets whether or not to combine the output and error pipes
// this will change the behavior of the process because error messages will be handled as output.
// only errors that are returned by the process itself will be handled as errors and pushed to the onOutput callback.
// this can be usefull for command they runs once so you do not have to handle the error pipe, because you should get the error then
// anyway. for processes that stay open, you should not use this because you will not get the errors while runtime.
func (p *Process) SetCombinePipes(combine bool) {
	p.combinePipes = combine
}

// SetTimeout sets the timeout for the process. If the process is not stopped after the timeout, it will be stopped
func (p *Process) SetTimeout(timeout time.Duration) {
	p.timoutSet = true
	p.timeOut = timeout
}

// AddStartCommands sets the arguments to pass to the command without waiting for any other setup.
// other than Command(string) you do not need to setup the whole environment and control structures.
func (p *Process) AddStartCommands(args ...string) {
	p.startCommands = args
}

// SetKeepRunning sets whether or not the process should stay open.
// If it does, it will not be stopped after the process handles the startup commands.
// this will change the behavior of the process because the started process will wait for inputs
// or beeing stopped by the Stop() method.
func (p *Process) SetKeepRunning(stayOpen bool) {
	p.stayOpen = stayOpen
}

// GetProcessWatcher returns the process watcher for the process.
func (p *Process) GetProcessWatcher() (*ProcessWatch, error) {
	if p.procWatch == nil {
		return nil, errors.New("process watcher is nil")
	}
	return p.procWatch, nil
}

// SetOnWaitDone sets the callback to call when the process is stopped after a regular wait for cmd execution.
// this is not called if the process is stopped by any other reason like timeout or killing the process.
func (p *Process) SetOnWaitDone(callback ProcHndlCallback) {
	p.onWaitDone = callback
}

// SetOnOutput sets the callback to call when output is received.
// Depending on the combinePipes flag, the error messages will be handled as output.
// the callback: func(string, error) bool
//   - the string is the output of the process
//   - the error is the error of the process
//   - the bool is the return value of the callback. if false is returned, the process will be stopped.
//     if true is returned, the process will continue to run.
//
// while runtime and an not combined pipe, anything that will be written to the error pipe will be handled as error.
// but also as message. so there is no need to handle booth messages in case of error.
//
//	process.SetOunOutput(func(msg string, err error) bool {
//	  if err != nil {
//	    // error.Error() is the same as msg. so no need to handle it twice
//	    return false // stop the process in this example. you can also return true to keep the process running
//	  }
//	  // do something with the message
//	  return true
//	})
func (p *Process) SetOnOutput(callback ProcCallback) {
	p.onOutput = callback
}

// SetOnInit sets the callback to call when the process is started.
// you will get the process object as argument.
// you can use this to get the process id and do something with it.
// but be carefull to not kill the process by accident.
// this package should handle the process for you. so you should not need to handle it by yourself.
func (p *Process) SetOnInit(callback ProcInfoCallback) {
	p.onInit = callback
}

// Command sends a command to the process. this command is send to the process by using the inPipe.
// the command will be send to the process as a string with a newline at the end.
// to get the response of the process, you need to set the onOutput callback.
// this is only possible if the process is set to stay open.
// if the process is not set to stay open, this will return an error.
// if the process is not started, this will return an error.
// if the process is stopped, this will return an error.
// if the inPipe is nil, this will return an error.
// if the command could not be send to the process, this will return an error.
func (p *Process) Command(cmd string) error {
	if p.inPipe == nil || !p.stayOpen || !p.started || p.stopped {
		if !p.started {
			return errors.New("process is not started")
		}
		if p.stopped {
			return errors.New("process is stopped")
		}
		if !p.stayOpen {
			return errors.New("process is not set to stay open")
		}
		return errors.New("inPipe is nil")
	}
	p.cmdLock.Lock()
	defer p.cmdLock.Unlock()
	// if we have a pipe to write to, write the command to the pipe so
	// the process can handle it
	if p.inPipe != nil {
		p.logger.Debug("sending command to process: ", cmd)
		if pcount, perr := io.WriteString(p.inPipe, cmd+"\n"); perr != nil {
			p.logger.Error("error while sending command to process: ", perr)
			return perr
		} else {
			p.logger.Debug("wrote ", pcount, " bytes to process")
		}
	}
	return nil
}

func (p *Process) closeInPipe() {
	p.logger.Debug("closing inPipe")
	if p.pipesClosed {
		p.logger.Warn("pipes are already closed")
		return
	}
	p.cmdLock.Lock()
	defer p.cmdLock.Unlock()
	// we do not care about the error here
	// because there is no easy way to check if the pipe is already closed.
	// and we just need to make sure the inPipe is closed
	if p.outPipe != nil {
		if err := p.outPipe.Close(); err != nil {
			p.logger.Error("error while closing outPipe: ", err)
		}
	}
	p.pipesClosed = true
}

// Stop stops the process.
// for this the default stop procedure is used.
// that means first the process willget an interrupt signal. then we wait for a short time.
// if the process is still running, we will kill it.
// this is the soft way to stop a process.
// if you want to kill a process in a more hard way, use the Kill() method instead.
// it returns
//   - the internal exit code of the process
//   - the real exit code of the process (if we have one. if not then -1 on error or 0 for some expected states like killed)
//   - an error if one occured
func (p *Process) Stop() (int, int, error) {
	p.logger.Debug("Stop: is called")
	// then use the process watcher to kill the process if he is still running
	if p.procWatch != nil {
		if running, _ := p.procWatch.IsRunning(); running {
			p.logger.Debug("Stop: process is running. stopping it with default signal.")
			if err := p.procWatch.StopWithDefaultSigOrder(); err != nil {
				p.logger.Error("Stop: error while stopping process: ", err)
				return systools.ErrorBySystem, 0, err
			}
		} else {
			p.logger.Debug("Stop: process is not running. not stopping it.")
		}
	}
	p.done()
	p.logger.Debug("Stop: process is stopped. current exist codes: ", p.internalExitCode, p.commandExitCode, p.runtimeError)
	return p.internalExitCode, p.commandExitCode, p.runtimeError
}

// done sets the stopped flag to true and closes the inPipe
func (p *Process) done() (int, int, error) {
	// first wait if the process may be existing by itself in 100 milliseconds
	if p.stayOpen && p.started && !p.stopped {
		if p.procWatch != nil {
			p.logger.Debug("Done: process is set to stay open. waiting for process to stop by itself")
			p.procWatch.WaitForStop(100*time.Millisecond, 10*time.Millisecond)
		}
	}

	p.logger.Debug("Done: setting stopped flag to true")
	p.stopped = true
	p.logger.Debug("Done: closing inPipe")
	p.closeInPipe()
	// give the process a chance to get done
	p.logger.Debug("Done: process done is also done. current exist codes: ", p.internalExitCode, p.commandExitCode, p.runtimeError)
	return p.internalExitCode, p.commandExitCode, p.runtimeError
}

// Kill kills the process and all its childs
// it uses the DefaultKillSignal to kill the processes. So this is the Hard way to kill a process.
// if you want to stop a process in a more graceful way use the Stop() method instead.
// it returns
//   - the internal exit code of the process
//   - the real exit code of the process (if we have one. if not then -1 on error or 0 for some expected states like killed)
//   - an error if one occured
func (p *Process) Kill() (int, int, error) {
	p.logger.Debug("kill is called")
	// then use the process watcher to kill the process if he is still running
	if p.procWatch != nil {
		if running, _ := p.procWatch.IsRunning(); running {
			p.logger.Debug("Kill: process is running. stopping it with kill signal.")
			if err := p.procWatch.StopChilds(DefaultKillSignal); err != nil {
				p.logger.Error("Kill: error while killing process: ", err)
				return systools.ErrorBySystem, 0, err
			}
		}
	}
	p.done()
	p.logger.Debug("Kill: process is killed. current exist codes: ", p.internalExitCode, p.commandExitCode, p.runtimeError)
	return p.internalExitCode, p.commandExitCode, p.runtimeError
}

// BlockWait blocks the current thread until the process is stopped.
// it uses a tick time to check if the process is stopped in intervals.
// this if meant to be used by tests. not in production code except you know what you are doing.
// most of the time you want to use the OnOutput callback to handle the output of the process
// and stop the process by using the Stop() method.
// or use go routines to handle the process output and stop the process if needed.
// using this method is a sign you do not need a background process.
func (p *Process) BlockWait(tickTime time.Duration) error {
	if !p.stayOpen {
		return errors.New("BlockWait: process is not set to stay open")
	}

	if !p.started {
		return errors.New("BlockWait: process is not started")
	}

	for {
		time.Sleep(tickTime)
		if p.stopped {
			break
		}
	}
	return nil
}

// WaitUntilRunning blocks the current thread until the process is running.
// it uses a tick time to check if the process is running in intervals.
// this can be usefull to make sure the process is running before you send commands to it.
func (p *Process) WaitUntilRunning(tickTime time.Duration) (time.Duration, error) {
	// messure time we are waiting for the process to start
	waitFailHitCount := 0
	start := time.Now()
	if !p.stayOpen {
		return 0, errors.New("WaitUntilRunning is not supported for processes that are not set to stay open")
	}

	if !p.started {
		return 0, errors.New("WaitUntilRunning: process is not started. you need to run Exec() first")
	}

	for {
		time.Sleep(tickTime)
		if p.procWatch != nil {
			if running, _ := p.procWatch.IsRunning(); running {
				break
			}
		} else {
			waitFailHitCount++
		}
	}
	p.logger.Debug("WaitUntilRunning: process is running. ", time.Since(start), waitFailHitCount)
	return time.Since(start), nil

}

// Exec starts the process onece or in background depending on the stayOpen flag.
// if its started in background, it will return directly after the process is started.
// so the return codes are not the real exit codes of the process. instead they are
// an internalcode that indicates the process is running in background.
//   - the internal exit code of the process
//   - the real exit code of the process (if we have one. if not then -1 on error or 0 for some expected states like killed)
//   - an error if one occured
func (p *Process) Exec() (int, int, error) {

	cmd := exec.Command(p.cmd, p.args...)
	p.logger.Debug("starting process: ", cmd, cmd.Args)
	// set the process group id to kill the whole process tree if possible
	TryPid2Pgid(cmd)
	var err error
	p.outPipe, err = cmd.StdoutPipe()
	if err != nil {
		return systools.ErrorBySystem, 0, err
	}
	// handling error pipe depending on the combinePipes flag
	if p.combinePipes {
		// combine the error and output pipes
		cmd.Stderr = cmd.Stdout
	} else {
		// use a seperate error pipe
		p.errorPipe, err = cmd.StderrPipe()
		if err != nil {
			return systools.ErrorBySystem, 0, err
		}
	}

	p.inPipe, err = cmd.StdinPipe()
	if err != nil {
		return systools.ErrorBySystem, 0, err
	}

	// send the startup commands to the process
	// if there are any
	if len(p.startCommands) > 0 {
		p.logger.Debug("sending startup commands to process.", len(p.startCommands))
		go func() {
			if !p.stayOpen {
				defer p.inPipe.Close()
				defer p.logger.Debug("closing inPipe becaue process is not set to stay open")
			}
			for _, arg := range p.startCommands {
				p.logger.Debug("sending startup command to process: ", arg)
				pCount, werr := io.WriteString(p.inPipe, arg+"\n")
				if werr != nil {
					p.logger.Error("error while sending startup command to process: ", werr)
				} else {
					p.logger.Debug("wrote ", pCount, " bytes to process")
				}
			}
			p.logger.Debug("sending startup commands to process done")
		}()
	}
	// now set the started flag to true
	p.started = true
	// start the process loop if the process should stay open
	if p.stayOpen {
		p.logger.Debug("process is set to stay open. starting process loop")
		p.wait.Add(1)
		go func() {
			// set timeout for the process
			// if the process is not stopped after the timeout
			nowTime := time.Now()
			p.logger.Debug("entering process loop")
			for {
				if p.timoutSet && p.timeOut > 0 {
					if nowTime.Add(p.timeOut).Before(time.Now()) {
						p.logger.Debug("process timed out. setting stopped flag to true")
						p.runtimeError = errors.New("process stopped by timeout")
						p.commandExitCode = RealCodeNotPresent
						p.internalExitCode = ExitTimeout
						p.stopped = true
					}
				}
				if p.stopped {
					break
				}
			}
			p.logger.Debug("leaving process loop")
			p.wait.Done()
		}()

		go func() {
			// wait for the process to finish
			p.internalExitCode, p.commandExitCode, p.runtimeError = p.startWait(cmd)
			p.logger.Debug("process finished in background by startWait. ", p.internalExitCode, p.commandExitCode, p.runtimeError)
		}()
		return systools.ExitOk, ExitInBackGround, nil
	} else {
		return p.startWait(cmd)
	}
}

// startWait starts the command and waits for it to finish
// it is also the mainloop for the process if the process is set to stay open
// it returns
//   - the internal exit code of the process
//   - the real exit code of the process (if we have one. if not then -1 on error or 0 for some expected states like killed)
//   - an error if one occured
//
// the return values are the same as in Exec()
func (p *Process) startWait(cmd *exec.Cmd) (int, int, error) {
	p.logger.Debug("startWait: entering")
	var err error
	err = cmd.Start()
	if err != nil {
		p.logger.Debug("startWait: error while starting command: ", err)
		return systools.ExitCmdError, RealCodeNotPresent, err
	}

	p.procWatch, err = NewProcessWatcherByCmd(cmd)
	if err != nil {
		return systools.ExitCmdError, RealCodeNotPresent, err
	}
	p.procWatch.SetLogger(p.logger)

	childCounts := 0
	if p.reportChildCount {
		go func(p *Process) {
			for {
				if p.started && !p.stopped {
					chldCnt := p.procWatch.CountChildsAll()
					time.Sleep(100 * time.Millisecond)
					if chldCnt != childCounts {
						p.logger.Debug("startWait: childs changed. ", mimiclog.Fields{"no of childs": chldCnt})
						if p.onChildCntChange != nil {
							p.onChildCntChange(chldCnt)
						}
						childCounts = chldCnt
					}
				}
			}
		}(p)
	}

	p.logger.Debug("startWait: process watcher created")
	// if we have a callback for the process info, call it
	if p.onInit != nil {
		p.logger.Debug("startWait: calling onInit callback")
		p.onInit(cmd.Process)
	}

	if !p.combinePipes {
		p.logger.Trace("startWait: assigning errorPipe to scanner")
		go func() {
			scanner := bufio.NewScanner(p.errorPipe)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				m := scanner.Text()
				keepRunning := true
				if p.onOutput != nil {
					p.logger.Debug("startWait (errorpipe): calling onOutput callback")
					if p.logger.IsTraceEnabled() {
						p.logger.Trace("startWait (errorpipe): calling onOutput callback with message: ", m)
					}
					keepRunning = p.onOutput(m, errors.New(m))
					p.logger.Debug("startWait (errorpipe): onOutput callback returned: ", keepRunning)

					if !keepRunning {
						p.logger.Debug("startWait: getting out of process loop because onOutput returned false")
						// try to kill the process by using group id if possible
						p.procWatch.Kill()

					}
				}
			}
		}()
	}

	scanner := bufio.NewScanner(p.outPipe)

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		keepRunning := true
		if p.onOutput != nil {
			p.logger.Debug("startWait: calling onOutput callback")
			if p.logger.IsTraceEnabled() {
				p.logger.Trace("startWait: calling onOutput callback with message: ", m)
			}
			keepRunning = p.onOutput(m, nil)
			p.logger.Debug("startWait: onOutput callback returned: ", keepRunning)
		}

		if !keepRunning {
			p.logger.Debug("startWait: getting out of process loop because onOutput returned false")
			// try to kill the process by using group id if possible
			err = p.procWatch.Kill()
			return systools.ExitByStopReason, 0, err
		}
	}
	// wait for the command to finish
	p.logger.Debug("startWait: waiting for command to finish")
	err = cmd.Wait()
	// if we have a callback for the wait done, call it
	if p.onWaitDone != nil {
		p.logger.Debug("startWait: calling onWaitDone callback")
		p.onWaitDone(err)
	} else {
		p.logger.Debug("startWait: no onWaitDone callback set")
	}
	// handle the error
	if err != nil {
		p.logger.Debug("startWait: error while waiting for command to finish: ", err)
		// if we have a callback for the output, call it with the error message as output
		// and the the original error
		if p.onOutput != nil {
			p.onOutput(err.Error(), err)
		}
		errRealCode := 0
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				errRealCode = status.ExitStatus()
			}

		}
		p.logger.Debug("startWait: returning error: ", err, errRealCode)
		return systools.ExitCmdError, errRealCode, err
	}
	p.logger.Debug("startWait: command finished")
	return systools.ExitOk, 0, err
}
