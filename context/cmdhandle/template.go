package cmdhandle

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"

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
		log.Fatal(err)
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

// LoadIncTempalte check if .inc.contxt.yml files exists
// and if this is the case the content will be loaded and all defined paths
// used to get values for parsing the template file
func LoadIncTempalte(path string) (string, bool) {

	fullPath := filepath.Dir(path)
	checkIncPath := fullPath + "/.inc.contxt.yml"
	fmt.Println(checkIncPath)
	existing, fileerror := dirhandle.Exists(checkIncPath)
	if fileerror != nil || !existing {
		return "", false
	}

	var importTemplate configure.IncludePaths
	file, ferr := ioutil.ReadFile(checkIncPath)
	if ferr != nil {
		return "", false
	}
	err := yaml.Unmarshal(file, &importTemplate)
	if err != nil {
		fmt.Println(err)
		os.Exit(incFileParseError)
	}
	// imports by Include.Folders
	if len(importTemplate.Include.Folders) > 0 || importTemplate.Include.Basedir {
		var dirs []string = importTemplate.Include.Folders
		if importTemplate.Include.Basedir {
			dirs = append(dirs, fullPath)
		}

		parsedTemplate, perr := ImportFolders(path, dirs...)
		if perr != nil {
			fmt.Println(perr)
			os.Exit(incFileParseError)
		}
		return parsedTemplate, true
	}
	return "", false
}

// GetPwdTemplate returns the template path if exists.
// it also parses the content of the template
// against imports and handles them
func GetPwdTemplate(path string) (configure.RunConfig, error) {
	var template configure.RunConfig
	existing, fileerror := dirhandle.Exists(path)
	if fileerror != nil {
		return template, fileerror
	}

	if existing {

		// first check if includes exists
		templateSource, inExists := LoadIncTempalte(path)
		if inExists {
			err2 := yaml.Unmarshal([]byte(templateSource), &template)

			if err2 != nil {
				fmt.Println("error parsing ", path, "after resolving imports. check result", err2)
				fmt.Println(templateSource)
				return template, err2
			}
		} else {
			// no imports .... load template file
			file, ferr := ioutil.ReadFile(path)
			if ferr != nil {
				return template, ferr
			}
			err := yaml.Unmarshal(file, &template)

			if err != nil {
				fmt.Println("error parsing template", path, "check result", err)
				fmt.Println(string(file))
				return template, err
			}
		}

	}
	return template, fileerror
}
