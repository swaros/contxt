package process

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/swaros/contxt/module/systools"
)

var (
	// usual prcesses path
	procPath = "/proc/%d"                     // general path for process info
	cmdPath  = path.Join(procPath, "cmdline") // path for process command line
	threads  = path.Join(procPath, "task")    // path for process tasks

)

// on linux we need to set the process group id to kill the whole process tree
// now any kill command have to add -pgid to kill the whole process tree
// like: syscall.Kill(-ts.process.processInfo.Pid, syscall.SIGKILL)
// returnning true to indicate that the process group id is set
func TryPid2Pgid(cmd *exec.Cmd) bool {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return true
}

// KillProcessTree kills the process by using the process group id
func KillProcessTree(pid int) error {
	return syscall.Kill(-pid, syscall.SIGKILL)
}

func ReadProc(pid int) (*ProcData, error) {
	if pid == 0 {
		return nil, errors.New("ReadProc: Error. pid can not be 0")
	}
	// get the comand line of the process
	proc := &ProcData{Pid: pid}
	cmdline, err := os.ReadFile(fmt.Sprintf(cmdPath, pid))
	if err != nil {
		return &ProcData{}, err
	} else {
		proc.Cmd = string(cmdline)
	}

	// read the threads of the process
	threads, err := os.ReadDir(fmt.Sprintf(threads, pid))
	if err != nil {
		return &ProcData{}, err
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
		if proc.Childs, err = GetChildPIDs(pid); err != nil {
			return &ProcData{}, err
		}
		for _, child := range proc.Childs {
			if child == 0 {
				continue
			}
			if childProc, err := ReadProc(child); err != nil {
				return &ProcData{}, err
			} else {
				proc.ChildProcs = append(proc.ChildProcs, childProc)
			}
		}
	}
	return proc, nil

}

func GetChildPIDs(pid int) ([]int, error) {
	if pid == 0 {
		return nil, errors.New("GetChildPIDs: Error. pid can not be 0")
	}
	// Read the children file
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/task/%d/children", pid, pid))
	if err != nil {
		return nil, err
	}
	dataStr := systools.TrimAllSpaces(string(data))

	// Split the data by spaces to get the child PIDs
	pids := strings.Split(strings.TrimSpace(dataStr), " ")

	// Convert the PIDs to integers
	childPIDs := make([]int, len(pids))
	for i, pid := range pids {
		if pid == "" {
			continue
		}

		pidInt, err := strconv.Atoi(pid)
		if err != nil {
			return nil, err
		}
		if pidInt != 0 {
			childPIDs[i] = pidInt
		}

	}

	return childPIDs, nil
}
