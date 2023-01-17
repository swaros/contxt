package ctemplate

import (
	"os"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yamc"
	"gopkg.in/yaml.v3"
)

const (
	DefaultTemplateFile = ".contxt.yml"
	DefaultIncludeFile  = ".inc.contxt.yml"
)

type Template struct {
	includeFile   string
	user          *user.User
	path          string
	includeConfig configure.IncludePaths
	dataMap       sync.Map
	tplParser     CtxTemplate
}

func New() *Template {
	return &Template{
		includeFile: DefaultIncludeFile,
		tplParser:   CtxTemplate{},
	}
}

func (t *Template) SetIncludeFile(file string) {
	t.includeFile = file
}

func (t *Template) GetIncludeFile() string {
	return t.includeFile
}

func (t *Template) Init() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	t.user = usr
	if dir, err := os.Getwd(); err == nil {
		t.path = dir
	} else {
		return err
	}
	return nil
}

func (t *Template) FindTemplateFileName() (string, bool) {
	if err := t.Init(); err != nil {
		panic(err)
	}
	fileName := t.path + string(os.PathSeparator) + DefaultTemplateFile
	if _, err := os.Stat(fileName); err == nil {
		return fileName, true
	} else {
		return "", false
	}
}

func (t *Template) FindIncludeFileName() (string, bool) {
	if err := t.Init(); err != nil {
		panic(err)
	}
	fileName := t.path + string(os.PathSeparator) + DefaultIncludeFile
	if _, err := os.Stat(fileName); err == nil {
		return fileName, true
	} else {
		return "", false
	}
}

func (t *Template) LoadTemplatePlain(path string) (configure.RunConfig, error) {
	var template configure.RunConfig
	if err := yacl.New(&template, yamc.NewYamlReader()).SetSingleFile(DefaultTemplateFile).Load(); err != nil {
		return template, err
	}
	return template, nil
}

func (t *Template) Load() (configure.RunConfig, bool, error) {
	// just check if we have a template file
	path, ok := t.FindTemplateFileName()
	if !ok {
		return configure.RunConfig{}, false, nil // no template file is also fine. so no error
	}
	templateData, ferr := os.ReadFile(path) // read the content of the file for later use
	if ferr != nil {
		return configure.RunConfig{}, false, ferr // this should not happen because we got already the file exists. so that might be a permission issue
	}
	if _, _, err := t.LoadInclude(); err != nil { // load the include files
		return configure.RunConfig{}, false, err // if we have an error here we can not continue
	}

	// now use the template parser to parse the template file
	t.tplParser.SetData(t.GetOriginMap())
	if templateParsed, err := t.tplParser.ParseTemplateString(string(templateData)); err != nil {
		return configure.RunConfig{}, false, err
	} else {
		var template configure.RunConfig
		// now we just use the plain yaml unmarshal to parse the template
		if err := yaml.Unmarshal([]byte(templateParsed), &template); err != nil {
			return configure.RunConfig{}, false, err
		} else {
			return template, true, nil
		}
	}
}

func (t *Template) LoadInclude() (configure.IncludePaths, bool, error) {
	// just check if we have a include file
	_, ok := t.FindIncludeFileName()
	if !ok {
		return configure.IncludePaths{}, false, nil // no include file is also fine. so no error
	}
	var include configure.IncludePaths
	// try to load the included files. can be json or yaml
	if err := yacl.New(&include, yamc.NewYamlReader(), yamc.NewJsonReader()).SetSingleFile(DefaultIncludeFile).Load(); err != nil {
		return include, false, err
	}
	t.includeConfig = include
	if err := t.parseIncludes(); err != nil { // parse the include files
		return include, false, err
	}
	return include, true, nil
}

func (t *Template) parseIncludes() error {
	if len(t.includeConfig.Include.Folders) > 0 {
		for _, include := range t.includeConfig.Include.Folders {

			if mapData, err := t.ImportFolder(include); err != nil {
				return err
			} else {
				t.UpdateOriginMap(mapData)
			}
		}
	}
	return nil
}

func (t *Template) ImportFolder(startPath string) (map[string]interface{}, error) {
	mapOrigin := t.GetOriginMap()

	err := filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var jsonMap map[string]interface{}
		var loaderr error
		hit := false
		if !info.IsDir() {
			var extension = filepath.Ext(path)
			var basename = filepath.Base(path)
			if basename == DefaultTemplateFile || basename == DefaultIncludeFile {
				return nil
			}
			switch extension {
			case ".json":
				rdr := yamc.NewJsonReader()
				loaderr = rdr.FileDecode(path, &jsonMap)
				hit = true
			case ".yaml", ".yml":
				rdr := yamc.NewYamlReader()
				loaderr = rdr.FileDecode(path, &jsonMap)
				hit = true
			}
			if loaderr != nil {
				return loaderr
			}
			if hit {
				mapOrigin, loaderr = MergeVariableMap(jsonMap, mapOrigin)
				if loaderr != nil {
					return loaderr
				}
			}
		}

		return nil
	})
	return mapOrigin, err
}

func (t *Template) GetOriginMap() map[string]interface{} {
	data := make(map[string]interface{})
	t.dataMap.Range(func(key, value interface{}) bool {
		data[key.(string)] = value
		return true
	})
	return data

}

func (t *Template) UpdateOriginMap(mapData map[string]interface{}) {
	for key, value := range mapData {
		t.dataMap.Store(key, value)
	}
}
