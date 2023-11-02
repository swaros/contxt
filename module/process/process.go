package process

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/swaros/contxt/module/systools"
)

type ProcCallback func(string, error) bool // the callback to call when output is received
type ProcInfoCallback func(*os.Process)    // the callback to call when the process is started

type Process struct {
	cmd           string           // the base command to run
	args          []string         // the arguments to pass to the command
	startCommands []string         // the commands to send to the process on startup
	OnOutput      ProcCallback     // the callback to call when output is received
	OnInit        ProcInfoCallback // the callback to call when the process is started
	outPipe       io.ReadCloser    // the pipe to read output from
	inPipe        io.WriteCloser   // the pipe to write input to
	stopped       bool             // whether or not the process has been stopped
	started       bool             // whether or not the process has been started
	cmdLock       sync.Mutex       // a lock for the command
	wait          sync.WaitGroup   // a wait group for the process
	stayOpen      bool             // whether or not the process should stay open. If it does, it will not be stopped after the process handles the startup commands
	timeOut       time.Duration    // the timeout for the process. If the process is not stopped after the timeout, it will be stopped
	timoutSet     bool             // whether or not the timeout is set
	errorCode     int              // the exit code of the process
	internCode    int              // the internal exit code of the process
	runError      error            // the error of the process
}

var (
	ExitInBackGround = 601 // the exit code when the process is running in the background
)

// NewProcess creates a new process with the given command and arguments
func NewProcess(cmd string, args ...string) *Process {
	return &Process{
		cmd:  cmd,
		args: args,
	}
}

// SetTimeout sets the timeout for the process. If the process is not stopped after the timeout, it will be stopped
func (p *Process) SetTimeout(timeout time.Duration) {
	p.timoutSet = true
	p.timeOut = timeout
}

// AddStartCommands sets the arguments to pass to the command
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

func (p *Process) SetOnOutput(callback ProcCallback) {
	p.OnOutput = callback
}

func (p *Process) SetOnInit(callback ProcInfoCallback) {
	p.OnInit = callback
}

func (p *Process) GetOutPipe() io.ReadCloser {
	return p.outPipe
}

func (p *Process) GetInPipe() io.WriteCloser {
	return p.inPipe
}

func (p *Process) Command(cmd string) {
	p.cmdLock.Lock()
	defer p.cmdLock.Unlock()
	if p.inPipe != nil {
		io.WriteString(p.inPipe, cmd+"\n")
	}
}

func (p *Process) Stop() (int, int, error) {
	p.stopped = true
	p.wait.Wait()
	if p.inPipe != nil {
		if err := p.inPipe.Close(); err != nil {
			return systools.ErrorBySystem, 0, err
		}
	}
	return p.errorCode, p.internCode, p.runError
}

// Exec executes the process and returns
//   - the internal exit code of the process
//   - the real exit code of the process (if we have one. if not then -1 on error or 0 for some expected states like killed)
//   - an error if one occured
func (p *Process) Exec() (int, int, error) {

	cmd := exec.Command(p.cmd, p.args...)
	// set the process group id to kill the whole process tree if possible
	PidWorkerForCmd(cmd)
	var err error
	p.outPipe, err = cmd.StdoutPipe()
	if err != nil {
		return systools.ErrorBySystem, 0, err
	}
	cmd.Stderr = cmd.Stdout
	p.inPipe, err = cmd.StdinPipe()
	if err != nil {
		return systools.ErrorBySystem, 0, err
	}

	// send the startup commands to the process
	// if there are any
	if len(p.startCommands) > 0 {
		go func() {
			if !p.stayOpen {
				defer p.inPipe.Close()
			}
			for _, arg := range p.startCommands {
				io.WriteString(p.inPipe, arg+"\n")
			}
		}()
	}

	if p.stayOpen {
		p.wait.Add(1)
		go func() {
			// set timeout for the process
			// if the process is not stopped after the timeout
			nowTime := time.Now()

			for {
				if p.timoutSet && p.timeOut > 0 {
					if nowTime.Add(p.timeOut).Before(time.Now()) {
						p.stopped = true
					}
				}
				if p.stopped {
					break
				}
			}
			p.wait.Done()
		}()

		go func() {
			p.errorCode, p.internCode, p.runError = p.startWait(cmd)
		}()

		return systools.ExitOk, ExitInBackGround, nil
	} else {
		return p.startWait(cmd)
	}
}

// startWait starts the command and waits for it to finish
// it returns
//   - the internal exit code of the process
//   - the real exit code of the process (if we have one. if not then -1 on error or 0 for some expected states like killed)
//   - an error if one occured
//
// the return values are the same as in Exec()
func (p *Process) startWait(cmd *exec.Cmd) (int, int, error) {
	var err error
	err = cmd.Start()
	p.started = true
	if err != nil {
		return systools.ExitCmdError, -1, err
	}

	if p.OnInit != nil {
		p.OnInit(cmd.Process)
	}
	scanner := bufio.NewScanner(p.outPipe)

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()
		keepRunning := true
		if p.OnOutput != nil {
			keepRunning = p.OnOutput(m, nil)
		}

		if !keepRunning {
			// try to kill the process by using group id if possible
			KillProcessTree(cmd.Process.Pid)
			return systools.ExitByStopReason, 0, err
		}
	}
	// wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		if p.OnOutput != nil {
			p.OnOutput(err.Error(), err)
		}
		errRealCode := 0
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				errRealCode = status.ExitStatus()
			}

		}
		return systools.ExitCmdError, errRealCode, err
	}

	return systools.ExitOk, 0, err
}
