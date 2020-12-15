package cmdhandle

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/swaros/contxt/context/output"

	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/dirhandle"
	"gopkg.in/yaml.v2"
)

const (
	incFileParseError  = 105
	mainFileParseError = 104
)

// FindTemplate searchs for Template files in different spaces
func FindTemplate() (string, bool) {
	// 1. looking in user mirror path
	usr, err := user.Current()
	if err != nil {
		GetLogger().Fatal(err)
	}

	homeDirYml := usr.HomeDir + configure.DefaultPath + configure.MirrorPath + DefaultExecYaml
	exists, exerr := dirhandle.Exists(homeDirYml)
	if exerr == nil && exists {
		return homeDirYml, true
	}

	// 2. looking in current path with user name as prefix
	dir, curerr := dirhandle.Current()
	if curerr == nil {
		userYml := dir + "/" + usr.Username + defaultExecYamlName
		exists, exerr = dirhandle.Exists(userYml)
		if exerr == nil && exists {
			return userYml, true
		}

		// 3. plain template in current dir
		regularPath := dir + DefaultExecYaml
		exists, exerr = dirhandle.Exists(regularPath)
		if exerr == nil && exists {
			return regularPath, true
		}
	}

	return "", false
}

// GetTemplate return current template
func GetTemplate() (configure.RunConfig, string, bool) {

	foundPath, success := FindTemplate()
	var template configure.RunConfig
	if !success {
		return template, "", false
	}
	ctemplate, err := GetPwdTemplate(foundPath)
	if err == nil {
		return ctemplate, foundPath, true
	}

	return template, "", false
}

func getIncludeConfigPath(path string) (string, string, bool) {
	fullPath := filepath.Dir(path)
	checkIncPath := fullPath + "/.inc.contxt.yml"
	existing, fileerror := dirhandle.Exists(checkIncPath)
	if fileerror != nil || !existing {
		return checkIncPath, fullPath, false
	}
	GetLogger().WithField("include-config", checkIncPath).Debug("found include setting")
	return checkIncPath, fullPath, true
}

func getIncludeConfig(path string) (configure.IncludePaths, string, string, bool) {
	var importTemplate configure.IncludePaths
	checkIncPath, fullPath, existing := getIncludeConfigPath(path)
	if !existing {
		return importTemplate, checkIncPath, fullPath, false
	}

	file, ferr := ioutil.ReadFile(checkIncPath)
	if ferr != nil {
		return importTemplate, checkIncPath, fullPath, false
	}
	err := yaml.Unmarshal(file, &importTemplate)
	if err != nil {
		fmt.Println(output.MessageCln(output.ForeRed, "error reading include config file: ", output.ForeWhite, checkIncPath), err)
		os.Exit(incFileParseError)
	}
	return importTemplate, checkIncPath, fullPath, true
}

// LoadIncTempalte check if .inc.contxt.yml files exists
// and if this is the case the content will be loaded and all defined paths
// used to get values for parsing the template file
func LoadIncTempalte(path string) (string, bool) {
	importTemplate, _, fullPath, existing := getIncludeConfig(path)
	if !existing {
		return "", false
	}
	// imports by Include.Folders
	if len(importTemplate.Include.Folders) > 0 || importTemplate.Include.Basedir {
		GetLogger().WithField("file", path).Info("parsing task-file")
		var dirs []string = importTemplate.Include.Folders
		if importTemplate.Include.Basedir {
			GetLogger().WithField("dir", fullPath).Debug("add parsing source dir")
			dirs = append(dirs, fullPath)
		}

		parsedTemplate, perr := ImportFolders(path, dirs...)
		if perr != nil {
			fmt.Println(perr)
			fmt.Println(output.MessageCln(output.ForeRed, "error parsing files from path: ", output.ForeWhite, path), perr)
			os.Exit(incFileParseError)
		}
		return parsedTemplate, true
	}
	return "", false
}

// GetParsedTemplateSource Returns the soucecode of the template
// including parsing placeholders
func GetParsedTemplateSource(path string) (string, error) {
	existing, fileerror := dirhandle.Exists(path)
	if fileerror != nil {
		return "", fileerror
	}
	if existing {
		// first check if includes exists
		templateSource, inExists := LoadIncTempalte(path)
		if inExists {
			return templateSource, nil
		}
		// no imports .... load template file
		file, ferr := ioutil.ReadFile(path)
		if ferr != nil {
			return "", ferr
		}
		return string(file), nil
	}
	notExistsErr := errors.New("file not exists")
	return "", notExistsErr
}

// GetPwdTemplate returns the template path if exists.
// it also parses the content of the template
// against imports and handles them
func GetPwdTemplate(path string) (configure.RunConfig, error) {
	var template configure.RunConfig
	source, err := GetParsedTemplateSource(path)
	if err != nil {
		return template, err
	}

	err2 := yaml.Unmarshal([]byte(source), &template)

	if err2 != nil {
		fmt.Println("error parsing ", path, "after resolving imports. check result", err2)
		fmt.Println(source)
		return template, err2
	}
	return template, nil
}
