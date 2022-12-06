// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Licensed under the MIT License
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package taskrun

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/imdario/mergo"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/manout"
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
		userYml := dir + string(os.PathSeparator) + usr.Username + defaultExecYamlName
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

// GetTemplate return current template, the absolute path,if it exists, any error
func GetTemplate() (configure.RunConfig, string, bool, error) {

	foundPath, success := FindTemplate()
	var template configure.RunConfig
	if !success {
		return template, "", false, errors.New("template not found or have failures")
	}
	ctemplate, err := GetPwdTemplate(foundPath)
	if err == nil {
		// checking required shared templates
		// and merge them into the current template
		if len(ctemplate.Config.Require) > 0 {
			for _, reqSource := range ctemplate.Config.Require {
				GetLogger().WithField("path", reqSource).Debug("handle required ")
				fullPath, pathError := CheckOrCreateUseConfig(reqSource)
				if pathError == nil {
					GetLogger().WithField("path", fullPath).Info("require path resolved")
					subTemplate, tError := GetPwdTemplate(fullPath + string(os.PathSeparator) + DefaultExecYaml)
					if tError == nil {
						mergo.Merge(&ctemplate, subTemplate, mergo.WithOverride, mergo.WithAppendSlice)
						GetLogger().WithField("template", ctemplate).Debug("merged")
					} else {
						return template, "", false, tError
					}
				} else {
					return template, "", false, pathError
				}
			}
		} else {
			GetLogger().Debug("no required files configured")
		}

		return ctemplate, foundPath, true, nil
	}

	return template, "", false, err
}

func getIncludeConfigPath(path string) (string, string, bool) {
	fullPath := filepath.Dir(path)
	checkIncPath := fullPath + string(os.PathSeparator) + ".inc.contxt.yml"
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
		fmt.Println(manout.MessageCln(manout.ForeRed, "error reading include config file: ", manout.ForeWhite, checkIncPath), err)
		systools.Exit(incFileParseError)
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
			fmt.Println(manout.MessageCln(manout.ForeRed, "error parsing files from path: ", manout.ForeWhite, path), perr)
			systools.Exit(incFileParseError)
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
		printErrSource(err2, source)
		return template, err2
	}
	return template, nil
}

func getLineNr(str string) (int, error) {
	re := regexp.MustCompile("[0-9]+")
	found := re.FindAllString(str, -1)
	if len(found) > 0 {
		return strconv.Atoi(found[0])
	}
	return -1, errors.New("no line number found in message " + str)
}

func printErrSource(err error, source string) {
	errPlain := err.Error()
	errParts := strings.Split(errPlain, ":")

	if len(errParts) == 3 { // this is depending an regular error message from yaml. like: yaml: line 3: mapping values are not allowed in this context
		if lineNr, lErr := getLineNr(errParts[1]); lErr == nil {
			sourceParts := strings.Split(source, "\n")
			if len(sourceParts) >= lineNr && lineNr >= 0 {
				min := lineNr - 3
				max := lineNr + 3
				if min < 0 {
					min = 0
				}
				if max > len(sourceParts) {
					max = len(sourceParts)
				}
				for i := min; i < max; i++ {
					nrback := manout.BackWhite
					nrFore := manout.ForeBlue
					msgFore := manout.ForeCyan
					msgBack := ""
					msg := ""
					if i == lineNr {
						nrback = manout.BackLightRed
						nrFore = manout.ForeRed
						msgFore = manout.ForeWhite
						msgBack = manout.BackRed
						msg = errParts[2]
					}

					padLineNr := fmt.Sprintf("%4d |", i)
					outstr := manout.MessageCln(
						nrback,
						nrFore,
						" ",
						padLineNr,
						manout.CleanTag,
						msgFore,
						msgBack,
						sourceParts[i],
						manout.CleanTag,
						manout.ForeLightYellow,
						" ",
						msg)

					fmt.Println(outstr)
				}
			} else {
				fmt.Println("source parsing faliure", sourceParts)
			}
		} else {
			fmt.Println(lErr)
		}
	} else {
		fmt.Println("unexpected message format ", len(errParts), " ", errParts)
	}
}
