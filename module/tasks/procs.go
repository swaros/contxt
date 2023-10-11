package tasks

import (
	"fmt"
	"os"
	"runtime"
)

func FindChildPid(pid int) ([]int, error) {
	var procList []int
	// only linux is supported
	// but others should not fail
	if runtime.GOOS != "linux" {
		return procList, nil
	}
	command := fmt.Sprintf("ps -o pid --no-headers --ppid %d", pid)
	runner := GetShellRunner()
	// internalExitCode, processExitCode, err
	intCode, realCode, err := runner.Exec(
		command,
		func(msg string, err error) bool {
			if err == nil {
				var code int
				fmt.Sscanf(msg, "%d", &code)
				procList = append(procList, code)
			}
			return true
		},
		func(proc *os.Process) {

		},
	)
	if err != nil {
		return procList, err
	}
	if intCode != 0 {
		return procList, fmt.Errorf("command %s failed with internal exit code %d", command, intCode)
	}
	if realCode != 0 {
		return procList, fmt.Errorf("command %s failed with exit code %d", command, realCode)
	}
	return procList, nil
}

func HandleAllMyPid(handlePid func(pid int) error) error {
	//mainProcs, err := FindChildPid(os.Getppid())
	procs, err := FindChildPid(os.Getpid())
	if err != nil {
		return err
	}

	for _, pid := range procs {
		if err := handlePid(pid); err != nil {
			return err
		}
	}
	return nil

}
