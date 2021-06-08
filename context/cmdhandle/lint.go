package cmdhandle

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/swaros/contxt/context/output"

	"github.com/kylelemons/godebug/pretty"
	"gopkg.in/yaml.v2"
)

func compareContent(a, b interface{}, showBooth bool, size int, right int) {
	diffOut := pretty.Compare(a, b)
	diffParts := strings.Split(diffOut, "\n")
	var errors []string
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
			fmt.Println(output.MessageCln(backColor, output.ForeYellow, output.Dim, lft, line))
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
			fmt.Println(output.MessageCln(backColor, output.BoldTag, output.ForeDarkGrey, line, output.ForeRed, output.BoldTag, rgt))
		}
		if !leftDiff && !rightDiff {
			line = getMaxLineString(line, size+right)
			i++
			fmt.Println(output.MessageCln(backColor, output.ForeBlue, line))
		}

	}

	if len(errors) > 0 {
		fmt.Println(output.MessageCln(output.ForeWhite, "found unsupported elements. you can add --full for showing supported elements"))
	}

	for _, errMsg := range errors {
		fmt.Println(output.MessageCln(output.ForeYellow, errMsg))
	}
}

// ShowAsYaml prints the generated source of the task file
func ShowAsYaml() {
	_, path, exists := GetTemplate()
	if exists {
		data, err := GetParsedTemplateSource(path)
		if err != nil {
			output.Error("template loading", err)
			return
		}
		fmt.Println(data)
	}
}

// LintOut prints the source code and the parsed content
// in a table view, and marks configured and not configured entries
// with dfferent colors
func LintOut(leftcnt, rightcnt int, all bool) {
	template, path, exists := GetTemplate()
	if exists && rightcnt >= 0 && leftcnt >= 0 {
		data, err := GetParsedTemplateSource(path)
		if err != nil {
			output.Error("template loading", err)
			return
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

				compareContent(origMap, m, all, leftcnt, rightcnt)
			}

		} else {
			prinfFile(path, leftcnt+rightcnt)
			output.Error("parsing error", yerr)
		}

	} else {
		output.Error("template not found ", path)
	}
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
