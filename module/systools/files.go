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

package systools

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/swaros/contxt/module/dirhandle"
)

func PathCompare(left, right string) bool {
	l := filepath.FromSlash(left)
	r := filepath.FromSlash(right)

	return l == r
}

func CopyFile(source, target string) error {
	r, err := os.Open(filepath.Clean(source))
	if err != nil {
		return err
	}
	defer r.Close()
	w, err := os.Create(filepath.Clean(target))
	if err != nil {
		return err
	}
	defer w.Close()
	if _, err := w.ReadFrom(r); err != nil {
		return err
	}
	return nil
}

// ReadFileAsStrings reads a file and returns its content as a slice of strings
// each line is a string
func ReadFileAsStrings(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}

func ReadFileAsString(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFileIfNotExists writes a file if it does not exist
// reports 0 if file was written, 1 if file exists, 2 if error
// on error, error is returned
func WriteFileIfNotExists(filename, content string) (int, error) {
	fileExists, existsCheckErr := dirhandle.Exists(filename)
	if existsCheckErr == nil && !fileExists {
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return 2, err
		}
		return 0, nil
	} else if fileExists {
		return 1, nil
	}
	return 2, existsCheckErr

}

func WriteFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

// Exists reports whether the named file or directory exists.
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

// fileExists reports whether the named file exists.
// it checks for file and not for directory
func FileExists(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}

// UpdateExistingFileIfNotContains updates a file if it does not contain a string
// this is made to avoid multiple updates of the same file
func UpdateExistingFileIfNotContains(filename, content, doNotContain string) (bool, error) {
	ok, errDh := Exists(filename)
	errmsg := ""
	if errDh == nil && ok {
		byteCnt, err := os.ReadFile(filename)
		if err != nil {
			return false, errors.New("file not readable " + filename + " " + err.Error())
		}
		strContent := string(byteCnt)
		if strings.Contains(strContent, doNotContain) {
			return false, errors.New("it seems file is already updated. it contains: " + doNotContain)
		} else {
			file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {

				return false, errors.New("error while opening file " + filename)
			}
			defer file.Close()
			if _, err := file.WriteString(content); err != nil {

				return false, errors.New("error adding content to file " + filename)
			}
			return true, nil
		}

	} else {
		errmsg = "file update error: file not exists " + filename
	}
	return false, errors.New(errmsg)
}

func IsDirWriteable(path string) bool {
	path, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		tempfile := uuid.New().String() + "_testfilecreate.tmp"
		file, err := os.CreateTemp(path, tempfile)
		if err != nil {
			return false
		}

		defer os.Remove(file.Name())
		defer file.Close()

		return true
	}
	return false
}
