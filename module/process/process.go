package process

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/swaros/contxt/module/systools"
)

type ProcCallback func(string, error) bool
type ProcInfoCallback func(*os.Process)

type Process struct {
	cmd      string
	args     []string
	runArgs  []string
	command  string
	OnOutput ProcCallback
	OnInit   ProcInfoCallback
	outPipe  io.ReadCloser
	inPipe   io.WriteCloser
	stopped  bool
	started  bool
	cmdLock  sync.Mutex
	stayOpen bool
}

func NewProcess(cmd string, args ...string) *Process {
	return &Process{
		cmd:  cmd,
		args: args,
	}
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

func (p *Process) Stop() error {
	p.stopped = true
	if p.inPipe != nil {
		return p.inPipe.Close()
	}
	return nil
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
