package cmdhandle

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/swaros/contxt/context/output"

	"github.com/swaros/contxt/context/configure"
	"gopkg.in/yaml.v2"
)

// LintOut is for finding errors in the yaml file
func LintOut(leftcnt, rightcnt int) {
	template, path, exists := GetTemplate()
	if exists {
		file, ferr := ioutil.ReadFile(path)
		yamlSource, err := getTemplateAsYAMLString(template)
		if ferr == nil && err == nil {

			yamlSource := strings.Split(yamlSource, "\n")
			fileSource := strings.Split(string(file), "\n")
			max := len(fileSource)
			if len(yamlSource) > max {
				max = len(yamlSource)
			}
			lineStr := "--------------------------------------------------------------------------------------------"
			fmt.Println(output.MessageCln(output.BackDarkGrey, output.ForeWhite, getMaxLineString("source", leftcnt), "|", getMaxLineString("current state", rightcnt)))
			fmt.Println(output.MessageCln(output.BackDarkGrey, output.ForeWhite, getMaxLineString(lineStr, leftcnt), "+", getMaxLineString(lineStr, rightcnt)))
			for i := 0; i < max; i++ {

				left := ""
				right := ""
				if i < len(yamlSource) {
					left = yamlSource[i]
				}
				if i < len(fileSource) {
					right = fileSource[i]
				}
				backColor := output.BackWhite
				if i%2 == 0 {
					backColor = output.BackLightGrey
				}
				fmt.Println(output.MessageCln(backColor, output.ForeBlue, getMaxLineString(right, leftcnt), output.ForeDarkGrey, "|", output.ForeGreen, getMaxLineString(left, rightcnt)))
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

func getTemplateAsYAMLString(template configure.RunConfig) (string, error) {
	res, err := yaml.Marshal(template)
	if err == nil {
		return string(res), nil
	}
	return "", err

}
