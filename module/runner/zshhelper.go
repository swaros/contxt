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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/swaros/contxt/module/tasks"
)

// the zsh helper is used to find the zsh binary and the fpath
// for the zsh completion files.
type ZshHelper struct {
	paths        []string
	usabalePaths []string
	firstUsable  string
	found        bool
	binpath      string
	collected    bool
}

// create a new zsh helper
func NewZshHelper() *ZshHelper {
	return &ZshHelper{
		paths:        []string{},
		usabalePaths: []string{},
		found:        false,
	}
}

// collect the zsh binary and the fpath
// this method is called by the other methods
// if not already called.
func (z *ZshHelper) collect() error {
	if z.collected {
		return nil
	}
	errInst := z.checkZshInstalled()
	if errInst != nil {
		return errInst
	}
	if z.binpath == "" {
		return fmt.Errorf("zsh seems not installed")
	}
	_, err := z.getFpathByExec()
	if err != nil {
		return err
	}
	z.findUsableFpath()
	z.found = z.firstUsable != ""
	z.collected = true
	return nil
}

// get the zsh binary path or an error
func (z *ZshHelper) GetBinPath() (string, error) {
	if z.found {
		return z.binpath, nil
	}
	err := z.collect()
	if err != nil {
		return "", err
	}
	return z.binpath, nil
}

// get all the fpaths they are usable (writeable) or an error
func (z *ZshHelper) GetFPaths() ([]string, error) {
	if z.found {
		return z.usabalePaths, nil
	}
	err := z.collect()
	if err != nil {
		return nil, err
	}
	return z.usabalePaths, nil
}

// get the first usable fpath or an error
func (z *ZshHelper) GetFirstFPath() (string, error) {
	if z.found {
		return z.firstUsable, nil
	}
	err := z.collect()
	if err != nil {
		return "", err
	}
	if z.firstUsable == "" {
		if len(z.paths) == 0 {
			return "", fmt.Errorf("no fpath found")
		}
		return "", fmt.Errorf("no usable fpath found in %d possible paths. you may retry with sudo", len(z.paths))
	}
	return z.firstUsable, nil
}

// check if zsh is installed
func (z *ZshHelper) checkZshInstalled() error {
	_, _, err := tasks.Execute("bash", []string{"-c"}, "which zsh", func(s string, err error) bool {
		if err == nil {
			z.binpath = s
		}
		return true
	}, func(p *os.Process) {

	})
	if err != nil {
		return err
	}
	return nil
}

// get the fpath by executing zsh -c "echo $fpath"
func (z *ZshHelper) getFpathByExec() (string, error) {
	fpaths := ""
	_, _, errorEx := tasks.Execute("zsh", []string{"-c"}, "echo $fpath", func(s string, err error) bool {
		if err == nil {
			fpaths = s
		}

		return true
	}, func(p *os.Process) {

	})
	z.paths = strings.Split(fpaths, " ")
	return fpaths, errorEx
}

// find the usable fpath by checking the permissions
func (z *ZshHelper) findUsableFpath() {

	for _, path := range z.paths {
		fileStats, err := os.Stat(path)
		if err != nil {
			continue
		}
		permissions := fileStats.Mode().Perm()
		if permissions&0b110000000 == 0b110000000 {
			if aPath, err := filepath.Abs(path); err == nil {
				z.usabalePaths = append(z.usabalePaths, aPath)
				if z.firstUsable == "" {
					z.firstUsable = aPath
				}
			}
		}
	}
}
