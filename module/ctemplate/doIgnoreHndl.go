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
	"strings"

	"github.com/swaros/contxt/module/systools"
)

// these chars are not allowed to be ignored, because they are used as delimiters
// and would break the whole file structure if we would not ignore them.
var notToIgnore = []string{" ", "\n", "\t", "\r"}

type IgnorreHndl struct {
	origin    string
	masked    string
	ignoreSet IgnoreSetMap
}

func NewIgnorreHndl(origin string) *IgnorreHndl {
	return &IgnorreHndl{
		origin:    origin,
		ignoreSet: NewIgnoreSetMap(),
	}
}

func (i *IgnorreHndl) AddIgnores(stringToIgnore ...string) {
	for _, toIgnore := range stringToIgnore {
		if toIgnore == "" {
			continue
		}
		if systools.SliceContains(notToIgnore, toIgnore) {
			continue
		}

		i.ignoreSet.Add(strings.Trim(toIgnore, " "))
	}
}

func (i *IgnorreHndl) CreateMaskedString() string {
	i.masked = i.origin
	replaces := make([]string, 0)
	for _, ign := range i.ignoreSet {
		key := ign.key
		searchText := ign.origin
		replaces = append(replaces, searchText)
		replaces = append(replaces, key)
	}
	replacer := strings.NewReplacer(replaces...)
	i.masked = replacer.Replace(i.origin)
	return i.masked
}

func (i *IgnorreHndl) RestoreOriginalString(useThisString string) string {
	for _, ign := range i.ignoreSet {
		key := ign.key
		searchText := ign.origin
		useThisString = strings.ReplaceAll(useThisString, key, searchText)
	}
	return useThisString
}
