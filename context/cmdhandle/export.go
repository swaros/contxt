// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Licensed under the MIT License
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package cmdhandle

import (
	"errors"
	"strings"
)

func ExportTask(target string) (string, error) {
	template, _, exists, terr := GetTemplate()
	if terr != nil {
		return "", terr
	}
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
