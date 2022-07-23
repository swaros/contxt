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
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kylelemons/godebug/pretty"
	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/manout"
	"gopkg.in/yaml.v3"
)

func compareContent(a, b interface{}, showBooth bool, size int, right int, noOut bool) bool {
	diffOut := pretty.Compare(a, b)
	diffParts := strings.Split(diffOut, "\n")
	var errors []string
	noDiff := true
	i := 0
	for _, line := range diffParts {
		backColor := manout.BackWhite
		if i%2 == 0 {
			backColor = manout.BackLightGrey
		}
		leftDiff := strings.HasPrefix(line, "+")
		rightDiff := strings.HasPrefix(line, "-")

		if leftDiff && showBooth {
			lft := getMaxLineString("", size)
			line = getMaxLineString(line, right)
			if !noOut {
				fmt.Println(manout.MessageCln(backColor, manout.ForeYellow, manout.Dim, lft, line))
			}
			i++
		}
		if rightDiff {
			errors = append(errors, "unsupported: "+line)
			rgt := getMaxLineString("  unsupported element ", right)
			line = getMaxLineString(line, size)
			backColor := manout.BackYellow
			if i%2 == 0 {
				backColor = manout.BackLightYellow
			}
			i++
			if !noOut {
				fmt.Println(manout.MessageCln(backColor, manout.BoldTag, manout.ForeDarkGrey, line, manout.ForeRed, manout.BoldTag, rgt))
			}
		}
		if !leftDiff && !rightDiff {
			line = getMaxLineString(line, size+right)
			i++
			if !noOut {
				fmt.Println(manout.MessageCln(backColor, manout.ForeBlue, line))
			}
		}

	}

	if len(errors) > 0 {
		noDiff = false
		manout.Error("found unsupported elements.", "count of errors:", len(errors))
	}

	for _, errMsg := range errors {
		fmt.Println(manout.MessageCln(manout.ForeYellow, errMsg))
	}
	return noDiff
}

func trySupressDefaults(yamlString string) string {
	ln := "\n"
	outStr := ""
	// first find all defauls values
	lines := strings.Split(yamlString, ln)
	for _, line := range lines {
		checks := strings.Split(line, ": ")
		if len(checks) == 2 {
			// these should have all possible defaults as values.
			if checks[1] != "[]" && checks[1] != "" && checks[1] != "\"\"" && checks[1] != "false" && checks[1] != "0" {
				outStr = outStr + line + ln
			}
		} else {
			outStr = outStr + line + ln
		}
	}
	// next find empty nodes
	lines = strings.Split(outStr, ln)
	newOut := ""
	max := len(lines)
	for index, recheck := range lines {
		if index > 0 && recheck != "" {

			last := recheck[len(recheck)-1:]
			if last == ":" && index < max {
				nextStr := lines[index+1]
				lastNext := nextStr[len(nextStr)-1:]
				if lastNext != ":" {
					newOut = newOut + recheck + ln
				}
			} else {
				newOut = newOut + recheck + ln
			}

		} else {
			newOut = newOut + recheck + ln
		}
	}
	return newOut
}

// ShowAsYaml prints the generated source of the task file
func ShowAsYaml(fullparsed bool, trySupress bool, indent int) {
	template, path, exists, terr := GetTemplate()
	if terr != nil {
		fmt.Println(manout.MessageCln(manout.ForeRed, "Error ", manout.CleanTag, terr.Error()))
		os.Exit(33)
		return
	}
	var b bytes.Buffer
	if exists {
		if fullparsed {
			yamlEncoder := yaml.NewEncoder(&b)
			yamlEncoder.SetIndent(indent)
			conerr := yamlEncoder.Encode(&template)
			if conerr == nil {
				if trySupress {
					fmt.Println(trySupressDefaults(b.String()))
				} else {
					fmt.Println(b.String())
				}

			} else {
				manout.Error("error parsing template", conerr)
				os.Exit(1)
			}

		} else {
			data, err := GetParsedTemplateSource(path)
			if err != nil {
				manout.Error("template loading", err)
				os.Exit(1)
			}
			fmt.Println(data)
		}
	}
}

func TestTemplate() error {
	if template, _, exists, terr := GetTemplate(); terr != nil {
		return terr
	} else {
		if !exists {
			GetLogger().Debug("no template exists to check")
		} else {
			// check version
			// if they is not matching, we die with an error
			if !configure.CheckCurrentVersion(template.Version) {
				return errors.New("unsupported version " + template.Version)
			}
		}
	}
	return nil
}

// LintOut prints the source code and the parsed content
// in a table view, and marks configured and not configured entries
// with dfferent colors
func LintOut(leftcnt, rightcnt int, all bool, noOut bool) bool {
	template, path, exists, terr := GetTemplate()
	if terr != nil {
		manout.Error("ERROR", terr.Error())
		return false
	}
	if exists && rightcnt >= 0 && leftcnt >= 0 {
		data, err := GetParsedTemplateSource(path)
		if err != nil {
			manout.Error("template loading", err)
			return false
		}
		origMap, yerr := YAMLToMap(data)
		if yerr == nil {
			conversionres, conerr := yaml.Marshal(template)
			if conerr == nil {
				m := make(map[string]interface{})
				amlerr := yaml.Unmarshal(conversionres, &m)
				if amlerr != nil {
					fmt.Println(amlerr)
					os.Exit(1)
				}

				return compareContent(origMap, m, all, leftcnt, rightcnt, noOut)
			}

		} else {
			prinfFile(path, leftcnt+rightcnt)
			manout.Error("parsing error", yerr)
		}

	} else {
		manout.Error("template not found ", path)
	}
	return false
}

func getMaxLineString(line string, length int) string {
	if len(line) < length {
		for i := len(line); i < length; i++ {
			line = line + " "
		}
	}
	if len(line) > length {
		line = line[0:length]
	}
	return line
}

func prinfFile(filename string, size int) error {
	data, err := GetParsedTemplateSource(filename)
	if err != nil {
		return err
	}

	backColor := manout.BackWhite
	lines := strings.Split(data, "\n")
	i := 0
	for _, line := range lines {
		i++
		prefix := getMaxLineString(strconv.Itoa(i), 5)
		line = getMaxLineString(line, size)
		fmt.Println(manout.MessageCln(manout.BackCyan, manout.ForeWhite, prefix, backColor, manout.ForeBlue, line))
	}
	return nil
}
