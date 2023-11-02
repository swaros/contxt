package process

import "os"

// basic struct to hold the data of a process
type ProcData struct {
	Pid         int    // process id
	Cmd         string // command line
	ThreadCount int    // number of threads
	Threads     []int  // list of threads pids
}

type ProcessDef struct {
	pData       *ProcData
	processInfo *os.Process
}

// ReadProc reads the process data of a process with the given pid
// and returns a ProcData struct
func NewProc(pid int) (*ProcData, error) {
	return ReadProc(pid)
}

// NewProcessWatcher creates a new ProcessDef struct
// and returns a pointer to it
// the ProcessDef struct holds the process data of the given process
func NewProcessWatcher(proc *os.Process) (*ProcessDef, error) {
	if pdef, err := NewProc(proc.Pid); err != nil {
		return nil, err
	} else {
		return &ProcessDef{
			pData:       pdef,
			processInfo: proc,
		}, nil
	}
}
