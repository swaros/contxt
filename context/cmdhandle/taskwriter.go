package cmdhandle

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// CreateImport creates import settings
func CreateImport(path string, pathToAdd string) error {
	importTemplate, pathToFile, fullPath, existing := getIncludeConfig(path)

	_, ferr := os.Stat(pathToAdd)
	if ferr != nil {
		return ferr
	}

	GetLogger().WithFields(logrus.Fields{
		"exists": existing,
		"path":   pathToFile,
		"folder": fullPath,
		"add":    pathToAdd,
	}).Debug("read imports")

	for _, existingPath := range importTemplate.Include.Folders {
		if existingPath == pathToAdd {
			err1 := errors.New("path already exists")
			return err1
		}
	}

	importTemplate.Include.Folders = append(importTemplate.Include.Folders, pathToAdd)
	res, err := yaml.Marshal(importTemplate)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(pathToFile, []byte(res), 0644)
	if err != nil {
		return err
	}

	return nil
}
