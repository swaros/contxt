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
