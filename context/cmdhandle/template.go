package cmdhandle

import (
	"io/ioutil"

	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/dirhandle"
	"gopkg.in/yaml.v2"
)

// GetTemplate return current template
func GetTemplate() (configure.RunConfig, bool) {
	var template configure.RunConfig
	dir, error := dirhandle.Current()
	if error != nil {
		return template, false
	}
	var path = dir + DefaultExecYaml
	already, errEx := dirhandle.Exists(path)
	if errEx != nil {
		return template, false
	}
	if already {

		ctemplate, err := GetPwdTemplate(path)
		if err == nil {
			return ctemplate, true
		}
	}
	return template, false
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
