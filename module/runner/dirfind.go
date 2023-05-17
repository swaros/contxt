// Copyright (c) 2023 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
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

// Package to decide what path is ment by some inputs, and also
// creates a label for them, if not set
package runner

import (
	"errors"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/systools"
)

func (c *CmdExecutorImpl) DirFindApplyAndSave(args []string) (string, error) {
	dir := c.DirFind(args)
	if dir != "" && dir != "." {
		index, ok := c.FindIndexByPath(dir)
		if !ok {
			return dir, errors.New("path not found in configuration")
		}
		if err := configure.GetGlobalConfig().SetCurrentPathIndex(index); err != nil {
			return dir, err
		}
		return dir, configure.GetGlobalConfig().SaveConfiguration()
	}
	return dir, nil // just no or the same path reported
}

func (c *CmdExecutorImpl) FindIndexByPath(path string) (index string, ok bool) {
	indexFound := ""
	configure.GetGlobalConfig().PathWorkerNoCd(func(index string, p string) {
		if p == path {
			ok = true
			indexFound = index
		}
	})
	return indexFound, ok
}

// DirFind returns the best matching part of depending the arguments, what of the stored paths
// would be the expected one
func (c *CmdExecutorImpl) DirFind(args []string) string {
	// no arguments? then we report the current dir
	if len(args) < 1 {
		return "."
	}

	paths := []string{}
	indexMatchMap := map[string]string{}

	configure.GetGlobalConfig().PathWorkerNoCd(func(index string, path string) {
		paths = append(paths, path)
		indexMatchMap[index] = path
	})

	if iPath, ok := indexMatchMap[args[0]]; ok {
		c.GetLogger().WithFields(logrus.Fields{"path": iPath}).Debug("Found match by index")
		return iPath
	}

	if p, ok := c.DecidePath(args, paths); ok {
		c.GetLogger().WithFields(logrus.Fields{"path": p}).Debug("Found match by comparing strings")
		return p
	}
	return "."

}

func (c *CmdExecutorImpl) DecidePath(searchWords, paths []string) (path string, found bool) {
	usePath := "."
	foundSome := false
	// starts iterate paths
	for _, search := range searchWords {
		if possiblePaths, ok := c.filterStringList(paths, search); ok {
			// keep the best match from the current list
			usePath = c.findBestByLast(possiblePaths, search)
			// now we continue with the filtered result
			paths = possiblePaths
			// for sure we found something
			foundSome = true
		} else {
			// nothing found by filtering. we could be already in a followup. so return the current findings
			return usePath, foundSome
		}

	}

	return usePath, foundSome
}

func (c *CmdExecutorImpl) filterStringList(paths []string, byWord string) (result []string, ok bool) {
	possiblePaths := []string{}
	found := false
	for _, path := range paths {
		// first we add any path to possible paths they have  at least one matching part
		if strings.Contains(path, byWord) {
			possiblePaths = append(possiblePaths, path)
			found = true
		}
	}
	sort.Strings(possiblePaths)
	return possiblePaths, found
}

func (c *CmdExecutorImpl) findBestByLast(paths []string, byWord string) string {
	match := ""
	bestIndex := -1
	bestlen := -1 // best length. we always go for the shortest
	for _, path := range paths {
		// first we add any path to possible paths they have  at least one matching part
		if strings.Contains(path, byWord) {
			if i := strings.Index(path, byWord); i > bestIndex {
				match = path
				bestIndex = i
				bestlen = systools.StrLen(path)
			} else if i == bestIndex && systools.StrLen(path) < bestlen { // on same position we use the shorter path
				match = path
				bestIndex = i
				bestlen = systools.StrLen(path)
			}
		}
	}
	return match
}
