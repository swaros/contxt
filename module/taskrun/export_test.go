package taskrun_test

import (
	"testing"

	"github.com/swaros/contxt/module/taskrun"
)

func TestExport(t *testing.T) {
	folderRunner("./../../docs/test/01multi", t, func(t *testing.T) {
		cmdLine, err := taskrun.ExportTask("task")
		if err != nil {
			t.Error(err)
		}

		if cmdLine == "" {
			t.Error("unexpected empty result")
		}
		xpectedCmdLine := `echo "hello 1"
        echo "hello 2"
`
		if clearStrings(cmdLine) != clearStrings(xpectedCmdLine) {
			t.Error("not equals [", clearStrings(xpectedCmdLine), "] and [", clearStrings(cmdLine), "]")
		}

	})
}
