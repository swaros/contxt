package cmdhandle

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/swaros/contxt/context/output"

	"github.com/kylelemons/godebug/pretty"
	"gopkg.in/yaml.v3"
)

func compareContent(a, b interface{}, showBooth bool, size int, right int, noOut bool) bool {
	diffOut := pretty.Compare(a, b)
	diffParts := strings.Split(diffOut, "\n")
	var errors []string
	noDiff := true
	i := 0
	for _, line := range diffParts {
		backColor := output.BackWhite
		if i%2 == 0 {
			backColor = output.BackLightGrey
		}
		leftDiff := strings.HasPrefix(line, "+")
		rightDiff := strings.HasPrefix(line, "-")

		if leftDiff && showBooth {
			lft := getMaxLineString("", size)
			line = getMaxLineString(line, right)
			if !noOut {
				fmt.Println(output.MessageCln(backColor, output.ForeYellow, output.Dim, lft, line))
			}
			i++
		}
		if rightDiff {
			errors = append(errors, "unsupported: "+line)
			rgt := getMaxLineString("  unsupported element ", right)
			line = getMaxLineString(line, size)
			backColor := output.BackYellow
			if i%2 == 0 {
				backColor = output.BackLightYellow
			}
			i++
			if !noOut {
				fmt.Println(output.MessageCln(backColor, output.BoldTag, output.ForeDarkGrey, line, output.ForeRed, output.BoldTag, rgt))
			}
		}
		if !leftDiff && !rightDiff {
			line = getMaxLineString(line, size+right)
			i++
			if !noOut {
				fmt.Println(output.MessageCln(backColor, output.ForeBlue, line))
			}
		}

	}

	if len(errors) > 0 {
		noDiff = false
		output.Error("found unsupported elements.", "count of errors:", len(errors))
	}

	for _, errMsg := range errors {
		fmt.Println(output.MessageCln(output.ForeYellow, errMsg))
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
	template, path, exists := GetTemplate()
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
				output.Error("error parsing template", conerr)
				os.Exit(1)
			}

		} else {
			data, err := GetParsedTemplateSource(path)
			if err != nil {
				output.Error("template loading", err)
				os.Exit(1)
			}
			fmt.Println(data)
		}
	}
}

// LintOut prints the source code and the parsed content
// in a table view, and marks configured and not configured entries
// with dfferent colors
func LintOut(leftcnt, rightcnt int, all bool, noOut bool) bool {
	template, path, exists := GetTemplate()
	if exists && rightcnt >= 0 && leftcnt >= 0 {
		data, err := GetParsedTemplateSource(path)
		if err != nil {
			output.Error("template loading", err)
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
			output.Error("parsing error", yerr)
		}

	} else {
		output.Error("template not found ", path)
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

	backColor := output.BackWhite
	lines := strings.Split(data, "\n")
	i := 0
	for _, line := range lines {
		i++
		prefix := getMaxLineString(strconv.Itoa(i), 5)
		line = getMaxLineString(line, size)
		fmt.Println(output.MessageCln(output.BackCyan, output.ForeWhite, prefix, backColor, output.ForeBlue, line))
	}
	return nil
}
