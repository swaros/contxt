package cmdhandle

import (
	"errors"
	"strings"
)

func ExportTask(target string) (string, error) {
	template, _, exists := GetTemplate()
	if !exists {
		return "", errors.New("template not exists")
	}
	var out string = ""
	for _, task := range template.Task {
		if task.ID == target {
			for _, need := range task.Needs {
				out = out + "\n# --- target " + need + " included ---- this is a need of " + target + "\n\n"
				if needtask, nErr := ExportTask(need); nErr == nil {
					out = out + needtask + "\n"
				}
			}
			out = out + strings.Join(task.Script, "\n") + "\n"
			for _, next := range task.Next {
				out = out + "\n# --- target " + next + " included ---- this is a next-task of " + target + "\n\n"
				if nextJob, sErr := ExportTask(next); sErr == nil {
					out = out + nextJob + "\n"
				}
			}
		}
	}
	return out, nil
}
