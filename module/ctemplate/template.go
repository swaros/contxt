// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

package ctemplate

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yaclint"
	"github.com/swaros/contxt/module/yamc"
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
	linter        *yaclint.Linter
	linting       bool
	logger        mimiclog.Logger
}

func New() *Template {
	return &Template{
		includeFile: DefaultIncludeFile,
		logger:      mimiclog.NewNullLogger(),
		tplParser:   CtxTemplate{},
	}
}

func (t *Template) SetLogger(logger mimiclog.Logger) {
	t.logger = logger
}

func (t *Template) SetLinting(linting bool) {
	t.linting = linting
}

func (t *Template) GetLinter() (*yaclint.Linter, error) {
	if t.linter == nil {
		return nil, errors.New("linter not initialized. You need to call SetLinting(true) first")
	}
	return t.linter, nil
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

func (t *Template) LoadV2() (configure.RunConfig, error) {
	var template configure.RunConfig
	confLoader := yacl.New(&template, yamc.NewYamlReader()).
		SetSingleFile(DefaultTemplateFile).
		SetCustomFileLoader(func(path string) ([]byte, error) {
			parsedContent, err := t.readAsTemplate()
			if err != nil {
				return nil, err
			}
			return []byte(parsedContent), nil
		})

	// if linting is enabled we need to track the files
	if t.linting {
		confLoader.SetTrackFiles()
	}
	if err := confLoader.Load(); err != nil {
		return template, err
	}

	if t.linting {
		t.linter = yaclint.NewLinter(*confLoader)
		if err := t.linter.Verify(); err != nil {
			return template, err
		}
	}
	return template, nil
}

func (t *Template) readAsTemplate() (string, error) {
	path, ok := t.FindTemplateFileName()
	if !ok {
		return "", errors.New("no template file found")
	}
	return t.GetFileParsed(path)
}

// GetFileParsed parses a template file and returns the parsed content
// by using all the the include files.
// the usage is the same as the go template parser. so you can use {{ .var }} to access the variables
func (t *Template) GetFileParsed(path string) (string, error) {
	templateData, ferr := os.ReadFile(path) // read the content of the file for later use
	if ferr != nil {
		return "", ferr
	}
	if _, _, err := t.LoadInclude(); err != nil { // load the include files
		return "", err // if we have an error here we can not continue
	}

	// now use the template parser to parse the template file
	t.tplParser.SetData(t.GetOriginMap())
	if templateParsed, err := t.tplParser.ParseTemplateString(string(templateData)); err != nil {
		return "", err
	} else {
		return templateParsed, nil
	}
}

func (t *Template) Load() (configure.RunConfig, bool, error) {
	if _, ok := t.FindTemplateFileName(); !ok {
		return configure.RunConfig{}, false, nil
	}

	if Template, err := t.LoadV2(); err != nil {
		return configure.RunConfig{}, false, err
	} else {
		return Template, true, nil
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
