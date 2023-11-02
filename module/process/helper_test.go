package process_test

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/swaros/contxt/module/process"
)

func BashPs(cmd string) ([]string, error) {
	cmdString := fmt.Sprintf(`ps -eo cmd | grep "%s"`, cmd)
	process := process.NewProcess("bash", "-c", cmdString)
	outputs := []string{}
	process.SetOnOutput(func(msg string, err error) bool {
		outputs = append(outputs, msg)
		return true
	})
	if _, _, err := process.Exec(); err != nil {
		return []string{}, err
	}
	return outputs, nil
}

func CmdIsRunning(t *testing.T, cmd string) bool {
	t.Helper()
	if runtime.GOOS == "windows" {
		return true // TODO: implement this
	}
	output, err := BashPs(cmd)
	if err != nil {
		t.Error(err)
		return false
	}
	for _, line := range output {
		// we just need to check if we hit the grep command
		// if we hit something without the grep command
		// it is the process we are looking for
		if !strings.Contains(line, "grep") {
			return true
		}
	}
	return false
}
