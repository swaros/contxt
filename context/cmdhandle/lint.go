package cmdhandle

import (
	"fmt"
	"strings"

	"github.com/swaros/contxt/context/output"

	"github.com/kylelemons/godebug/pretty"
	"gopkg.in/yaml.v2"
)

func compareContent(a, b interface{}, showBooth bool, size int, right int) {
	diffOut := pretty.Compare(a, b)
	diffParts := strings.Split(diffOut, "\n")
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
			rgt := getMaxLineString("  unsupported element", right)
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
}

// LintOut prints the source code and the parsed content
// in a table view, and marks configured and not configured entries
// with dfferent colors
func LintOut(leftcnt, rightcnt int, all bool) {
	template, path, exists := GetTemplate()

	if exists && rightcnt >= 0 && leftcnt >= 0 {

		origMap, yerr := ImportYAMLFile(path)
		if yerr == nil {
			conversionres, conerr := yaml.Marshal(template)
			if conerr == nil {
				m := make(map[string]interface{})
				yaml.Unmarshal(conversionres, &m)

				compareContent(origMap, m, all, leftcnt, rightcnt)
			}

		}

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
