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
			if canRun, message := checkRequirements(task.Requires); canRun {
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
			} else {
				out = out + "\n# --- -----------------------------------------------------------------------------------  ---- \n"
				out = out + "# --- a  sequence of the target " + target + " is ignored because of a failed requirement  ---- \n"
				out = out + "# --- this is might be an usual case. The reported reason to skip: " + message + "  \n"
				out = out + "# --- -----------------------------------------------------------------------------------  ---- \n"
			}
		}
	}
	return out, nil
}
