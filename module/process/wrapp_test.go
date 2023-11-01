package process_test

import (
	"os"
	"testing"

	"github.com/swaros/contxt/module/process"
)

func TestReadProc(t *testing.T) {
	proc, err := process.ReadProc(os.Getpid())
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("Pid: %d", proc.Pid)
		t.Logf("Cmd: %s", proc.Cmd)
		if proc.Cmd == "" {
			t.Error("Cmd is empty")
		}
		t.Logf("ThreadCount: %d", proc.ThreadCount)
		if proc.Pid != os.Getpid() {
			t.Error("Pid is not correct")
		}

	}
}
