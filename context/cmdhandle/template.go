package cmdhandle

import (
	"io/ioutil"
	"log"
	"os/user"

	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/dirhandle"
	"gopkg.in/yaml.v2"
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

// GetPwdTemplate returns the template from current path if exists
func GetPwdTemplate(path string) (configure.RunConfig, error) {
	var template configure.RunConfig
	existing, fileerror := dirhandle.Exists(path)
	if fileerror != nil {
		return template, fileerror
	}

	if existing {

		file, ferr := ioutil.ReadFile(path)
		if ferr != nil {
			return template, ferr
		}
		err := yaml.Unmarshal(file, &template)

		if err != nil {
			return template, err
		}
	}
	return template, fileerror
}
