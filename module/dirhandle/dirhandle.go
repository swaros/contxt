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

package dirhandle

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/swaros/contxt/configure"
)

// Current returns the current path
func Current() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir, err
}

// PrintDir prints the all the paths
func PrintDir(number int) {
	for index, path := range configure.UsedConfig.Paths {
		if number == index {
			fmt.Println(path)
			return
		}
	}
	fmt.Println(".")
}

// GetDir returns the path by index
func GetDir(number int) string {
	for index, path := range configure.UsedConfig.Paths {
		if number == index {

			return path
		}
	}
	return "."
}

// Exists checks if a path exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// FileTypeHandler calls function depending on file ending
// and if this fil exists
func FileTypeHandler(path string, jsonHandle func(string), yamlHandle func(string), notExists func(string, error)) {
	fileInfo, err := os.Stat(path)
	if err == nil && !fileInfo.IsDir() {
		var extension = filepath.Ext(path)
		var basename = filepath.Base(path)
		switch extension {
		case ".yaml", ".yml":
			yamlHandle(basename)
		case ".json":
			jsonHandle(basename)
		}
	} else {
		notExists(path, err)
	}
}