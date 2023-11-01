package process

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/swaros/contxt/module/systools"
)

type ProcCallback func(string, error) bool
type ProcInfoCallback func(*os.Process)

type Process struct {
	cmd        string
	args       []string
	runArgs    []string
	command    string
	OnOutput   ProcCallback
	OnInit     ProcInfoCallback
	outPipe    io.ReadCloser
	inPipe     io.WriteCloser
	stopped    bool
	started    bool
	cmdLock    sync.Mutex
	wait       sync.WaitGroup
	stayOpen   bool
	timeOut    time.Duration
	timoutSet  bool
	errorCode  int
	internCode int
	runError   error
}

var (
	ExitInBackGround = 601
)

func NewProcess(cmd string, args ...string) *Process {
	return &Process{
		cmd:  cmd,
		args: args,
	}
}

func (p *Process) SetTimeout(timeout time.Duration) {
	p.timoutSet = true
	p.timeOut = timeout
}

func (p *Process) SetRunArgs(args ...string) {
	p.runArgs = args
}

func (p *Process) SetStayOpen(stayOpen bool) {
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
	p.command = cmd
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

func (p *Process) Run() error {
	cmd := exec.Command(p.cmd, p.args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		for _, arg := range p.runArgs {
			io.WriteString(stdin, arg)
		}
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", out)
	return nil
}

func (p *Process) Exec() (int, int, error) {

	cmd := exec.Command(p.cmd, p.args...)
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
	go func() {
		if !p.stayOpen {
			defer p.inPipe.Close()
		}
		for _, arg := range p.runArgs {
			io.WriteString(p.inPipe, arg+"\n")
		}
	}()

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

func (p *Process) startWait(cmd *exec.Cmd) (int, int, error) {
	var err error
	err = cmd.Start()
	p.started = true
	if err != nil {
		return systools.ExitCmdError, 0, err
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
			cmd.Process.Kill()
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
