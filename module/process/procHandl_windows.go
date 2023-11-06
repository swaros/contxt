package process

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

// WinProcData is a struct to hold the data of a process on windows
type WinProcData struct {
	Pid         int
	ProcName    string
	SessionName string
	SessionId   int
	MemUsage    string
	Status      string
	UserName    string
	CpuTime     string
	WindowTitle string
}

// for windows we just do nothing with the cmd
// returnning false to indicate that the process group id is not set
func TryPid2Pgid(cmd *exec.Cmd) bool {
	return false
}

func KillProcessTree(pid int) error {
	return syscall.Kill(pid, syscall.SIGKILL)
}

func ReadProc(pid int) (*ProcData, error) {
	proc, err := ProcInfo(pid)
	if err != nil {
		return &ProcData{}, err
	}
	return &ProcData{
		Pid:         proc.Pid,
		Cmd:         proc.ProcName,
		ThreadCount: 0,
		Threads:     []int{},
	}, nil
}

// parsing the csv output of tasklist comming from windows tasklist command
func WinCsvToProcData(csv string) ([]WinProcData, error) {
	var procList []WinProcData

	parts := strings.Split(csv, "\n")
	for _, line := range parts {
		if strings.HasPrefix(line, "\"") {
			line = strings.Trim(line, "\"")
			parts := strings.Split(line, "\",\"")
			if len(parts) == 9 {
				var proc WinProcData
				if _, err := fmt.Sscanf(parts[1], "%d", &proc.Pid); err != nil {
					return procList, err
				}
				proc.ProcName = parts[0]
				proc.SessionName = parts[2]
				if _, err := fmt.Sscanf(parts[3], "%d", &proc.SessionId); err != nil {
					return procList, err
				}
				proc.MemUsage = parts[4]
				proc.Status = parts[5]
				proc.UserName = parts[6]
				proc.CpuTime = parts[7]
				proc.WindowTitle = parts[8]
				procList = append(procList, proc)
			}
		}
	}
	return procList, nil
}

func ProcInfo(pid int) (WinProcData, error) {
	// TASKLIST /V /FO "CSV" /NH /FI "PID eq 2656"
	cmd := exec.Command("TASKLIST", "/V", "/FO", "CSV", "/NH", "/FI", fmt.Sprintf("PID eq %d", pid))
	result, err := cmd.Output()
	if err != nil {
		return WinProcData{}, err
	}
	procInfo := string(result)
	procList, err := WinCsvToProcData(procInfo)
	if err != nil {
		return WinProcData{}, err
	}
	if len(procList) == 0 {
		return WinProcData{}, fmt.Errorf("no process found")
	}
	return procList[0], nil
}
