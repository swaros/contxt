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
	includeFile   string                           // the include file name what is usual .inc.contxt.yml in the current directory
	user          *user.User                       // the current user (set in Init)
	path          string                           // the current path (set in Init)
	includeConfig configure.IncludePaths           // the include config contains all the files to include
	dataMap       sync.Map                         // the data map contains all key values they are used for parsing go/template files
	tplParser     CtxTemplate                      // the template parser that is parsing any text file with go/template placeholders
	linter        *yaclint.Linter                  // the linter is used to lint the template files
	linting       bool                             // if linting is enabled we need to track the files
	logger        mimiclog.Logger                  // the logger interface
	onLoadFn      func(*configure.RunConfig) error // if this callback is set it will be called after the template file is loaded
}

// create a new Template struct
func New() *Template {
	return &Template{
		includeFile: DefaultIncludeFile,
		logger:      mimiclog.NewNullLogger(),
		tplParser:   CtxTemplate{},
	}
}

// set the callback function that is called after the template file is loaded
func (t *Template) SetOnLoad(fn func(*configure.RunConfig) error) {
	t.onLoadFn = fn
}

// SetLogger sets the logger interface
func (t *Template) SetLogger(logger mimiclog.Logger) {
	t.logger = logger
}

// SetLinting enables or disables the linting process
func (t *Template) SetLinting(linting bool) {
	t.linting = linting
}

// GetLinter returns the linter interface
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

// Init initializes the Template struct
// this is ment to be called before any other function
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

// FindTemplateFileName searchs for Template files in the current directory
// the bool value indicates if the file exists or not
// the default template file name is .contxt.yml in the current directory
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

// FindIncludeFileName returns the full path to the include file if it exists
// the bool value indicates if the file exists or not
// the default include file name is .inc.contxt.yml in the current directory
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

// the customTemplateLoader is used to load the template file,
// and make sure that any template placeholders are processed.
// this function is injected in the yacl. load process.
func (t *Template) customTemplateLoader(path string) ([]byte, error) {
	parsedContent, err := t.readAsTemplate()
	if err != nil {
		return nil, err
	}
	return []byte(parsedContent), nil
}

// LoadV2ByAbsolutePath loads the template file and returns the parsed content.
// this is ment for loading a template file from a different location.
func (t *Template) LoadV2ByAbsolutePath(absolutePath string) (configure.RunConfig, error) {
	var template configure.RunConfig
	confLoader := yacl.New(&template, yamc.NewYamlReader()).
		SetFileAndPathsByFullFilePath(absolutePath).
		SetCustomFileLoader(t.customTemplateLoader)

	if err := confLoader.Load(); err != nil {
		return template, err
	}
	return template, nil
}

// LoadV2 loads the template file and returns the parsed content.
// this is ment for loading the default template file.
// if you like to load a template file from a different location
// use LoadV2ByAbsolutePath
func (t *Template) LoadV2() (configure.RunConfig, error) {
	var template configure.RunConfig
	confLoader := yacl.New(&template, yamc.NewYamlReader()).
		SetSingleFile(DefaultTemplateFile).
		SetCustomFileLoader(t.customTemplateLoader)

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

// readAsTemplate reads the template file and returns the parsed content
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

// Load loads the template file and returns the parsed content.
// this is ment for loading the default template file.
// so anything here is depending the default template file name,
// and the current directory.
func (t *Template) Load() (configure.RunConfig, bool, error) {
	if _, ok := t.FindTemplateFileName(); !ok {
		return configure.RunConfig{}, false, nil
	}

	if Template, err := t.LoadV2(); err != nil {
		return configure.RunConfig{}, false, err
	} else {
		// if we have a callback function we call it here to let the user do some stuff
		// and maybe change the template file
		if t.onLoadFn != nil {
			if err := t.onLoadFn(&Template); err != nil {
				return Template, false, err
			}
		}
		return Template, true, nil
	}
}

// LoadInclude loads the include files and returns the parsed content.
// these files are defined in the default include file named .inc.contxt.yml
// any of these files can have template placeholders, they will processed
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
