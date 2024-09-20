package runner

import (
	"errors"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/systools"
	"gopkg.in/yaml.v2"
)

func CreateContxtFile() error {
	// Define the content of the file
	ctxContent := `
task:
  - id: "my_task"
    script:
      - echo "Hello World"
`
	_, err := systools.WriteFileIfNotExists(".contxt.yml", ctxContent)
	return err
}

// AddPathToIncludeImports adds a path to the include section of the .inc.contxt.yml file
// so it will be read as an value file
func AddPathToIncludeImports(incConfig *configure.IncludePaths, pathToAdd string) (string, error) {

	ok, er := systools.Exists(pathToAdd)
	if er != nil {
		return "", er
	}
	if !ok {
		return "", errors.New("path " + pathToAdd + " does not exist")
	}

	incConfig.Include.Folders = append(incConfig.Include.Folders, pathToAdd)
	// Marshal the template
	res, err := yaml.Marshal(incConfig)
	if err != nil {
		return "", err
	}

	return string(res), nil
}
