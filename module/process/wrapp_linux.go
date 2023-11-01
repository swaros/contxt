package process

import (
	"fmt"
	"os"
	"path"
)

var (
	// usual prcesses path
	procPath = "/proc/%d"                     // general path for process info
	cmdPath  = path.Join(procPath, "cmdline") // path for process command line
	threads  = path.Join(procPath, "task")    // path for process tasks

)

func ReadProc(pid int) (ProcData, error) {
	// get the comand line of the process
	proc := ProcData{Pid: pid}
	cmdline, err := os.ReadFile(fmt.Sprintf(cmdPath, pid))
	if err != nil {
		return ProcData{}, err
	} else {
		proc.Cmd = string(cmdline)
	}

	// read the threads of the process
	threads, err := os.ReadDir(fmt.Sprintf(threads, pid))
	if err != nil {
		return ProcData{}, err
	} else {
		proc.ThreadCount = len(threads)
		for _, thread := range threads {
			// convert the thread name to an int
			// and add it to the list of threads
			tPid := 0
			if _, err := fmt.Sscanf(thread.Name(), "%d", &tPid); err == nil && tPid != 0 && tPid != pid {
				proc.Threads = append(proc.Threads, tPid)
			}

		}
	}
	return proc, nil

}
