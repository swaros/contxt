package tasks

import (
	"os"
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

func (ts *TaskDef) StopTrackProcess() {
	ts.process = nil
}

func (ts *TaskDef) GetProcess() *ProcessDef {
	return ts.process
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

func (ts *TaskDef) IsProcessRunning() bool {
	if ts.process == nil {
		return false
	}
	if ts.process.processInfo != nil {
		if ts.process.processInfo.Pid > 0 {
			proc, err := os.FindProcess(ts.process.processInfo.Pid)
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
