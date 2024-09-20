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

package runner

import (
	"errors"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/systools"
	"gopkg.in/yaml.v2"
)

func CreateContxtFile() error {
	// Define the content of the file
	ctxContent := `
task:
  - id: "my_task"
    script:
      - echo "Hello World"
`
	_, err := systools.WriteFileIfNotExists(".contxt.yml", ctxContent)
	return err
}

// AddPathToIncludeImports adds a path to the include section of the .inc.contxt.yml file
// so it will be read as an value file
func AddPathToIncludeImports(incConfig *configure.IncludePaths, pathToAdd string) (string, error) {

	ok, er := systools.Exists(pathToAdd)
	if er != nil {
		return "", er
	}
	if !ok {
		return "", errors.New("path " + pathToAdd + " does not exist")
	}

	incConfig.Include.Folders = append(incConfig.Include.Folders, pathToAdd)
	// Marshal the template
	res, err := yaml.Marshal(incConfig)
	if err != nil {
		return "", err
	}

	return string(res), nil
}
