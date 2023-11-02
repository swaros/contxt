package process

import (
	"os"
	"os/exec"
)

// basic struct to hold the data of a process
type ProcData struct {
	Pid         int    // process id
	Cmd         string // command line
	ThreadCount int    // number of threads
	Threads     []int  // list of threads pids
}

type ProcessWatch struct {
	pData       *ProcData
	processInfo *os.Process
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
		}, nil
	}
}

func NewProcessWatcherByProcessInfo(proc *os.Process) (*ProcessWatch, error) {
	if pdef, err := NewProc(proc.Pid); err != nil {
		return nil, err
	} else {
		return &ProcessWatch{
			pData:       pdef,
			processInfo: proc,
		}, nil
	}
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

func (proc *ProcessWatch) GetProcessInfo() *os.Process {
	return proc.processInfo
}

func (proc *ProcessWatch) Kill() error {
	return KillProcessTree(proc.pData.Pid)
}

func (proc *ProcessWatch) Update() error {
	if pdef, err := NewProc(proc.pData.Pid); err != nil {
		return err
	} else {
		proc.pData = pdef
		return nil
	}
}

func (proc *ProcessWatch) IsRunning() (bool, error) {
	if err := proc.Update(); err != nil {
		return false, err
	}
	return proc.pData.Pid > 0, nil
}
