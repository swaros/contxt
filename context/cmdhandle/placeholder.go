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
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var keyValue sync.Map

// SetPH add key value pair
func SetPH(key, value string) {
	GetLogger().WithField(key, value).Trace("add/overwrite placeholder")
	keyValue.Store(key, value)
}

func AppendToPH(key, value string) bool {
	if val, ok := keyValue.Load(key); ok {
		value = val.(string) + value
		keyValue.Store(key, value)
		return true
	}
	return false
}

func SetIfNotExists(key, value string) {
	_, ok := keyValue.Load(key)
	if !ok {
		keyValue.Store(key, value)
	}

}

// GetPH the content of the key as string. if exists.
func GetPHExists(key string) (string, bool) {
	result, ok := keyValue.Load(key)
	if ok {
		return result.(string), ok
	}
	return "", ok
}

// GetPH the content of the key. but at least a empty
// string if not exists. so this is not usefull
// to test if the PH was set.
func GetPH(key string) string {
	result, ok := keyValue.Load(key)
	if ok {
		GetLogger().WithField(key, result.(string)).Trace("deliver content from placeholder")
		return result.(string)
	}
	GetLogger().WithField("key", key).Trace("returns empty string because key is not set")
	return ""
}

// HandlePlaceHolder replaces all defined placeholders
func HandlePlaceHolder(line string) string {
	var scopeVars map[string]string = make(map[string]string)
	for {
		return handlePlaceHolder(line, scopeVars)
	}
}

func HandlePlaceHolderWithScope(line string, scopeVars map[string]string) string {
	for {
		return handlePlaceHolder(line, scopeVars)
	}
}

func handlePlaceHolder(line string, scopeVars map[string]string) string {

	// this block is for logging at trace level only
	if GetLogger().IsLevelEnabled(logrus.TraceLevel) {

		for key, value := range scopeVars {
			keyName := "${" + key + "}"
			if strings.Contains(line, keyName) {
				GetLogger().WithField("line", line).Trace("scope replace: source")
				GetLogger().WithField(keyName, value).Trace("scope replace: variables")
			}
			line = strings.ReplaceAll(line, keyName, value)
		}

		keyValue.Range(func(key, value interface{}) bool {
			keyName := "${" + key.(string) + "}"
			if strings.Contains(line, keyName) {
				GetLogger().WithField("line", line).Trace("replace: source")
				GetLogger().WithField(keyName, value.(string)).Trace("replace: variables")
			}
			line = strings.ReplaceAll(line, keyName, value.(string))
			line = handleMapVars(line)
			return true
		})
	}

	for key, value := range scopeVars {
		keyName := "${" + key + "}"
		line = strings.ReplaceAll(line, keyName, value)
	}

	keyValue.Range(func(key, value interface{}) bool {
		keyName := "${" + key.(string) + "}"
		line = strings.ReplaceAll(line, keyName, value.(string))
		return true
	})

	line = handleMapVars(line)
	for _, value := range os.Environ() {
		pair := strings.SplitN(value, "=", 2)
		if len(pair) == 2 {
			key := pair[0]
			val := pair[1]
			keyName := "${" + key + "}"
			line = strings.ReplaceAll(line, keyName, val)
		}
	}
	return line
}

func handleMapVars(line string) string {
	dataKeys := GetDataKeys()
	if len(dataKeys) == 0 {
		return line
	}
	GetLogger().WithField("key-count", len(dataKeys)).Trace("parsing keymap placeholder")
	for _, keyname := range dataKeys {
		lookup := "${" + keyname + ":"
		if strings.Contains(line, lookup) {
			start := strings.Index(line, lookup)
			if start > -1 {
				end := strings.Index(line[start:], "}")
				if end > -1 {
					pathLine := line[start+len(lookup) : start+end]
					if pathLine != "" {
						replace := lookup + pathLine + "}"
						GetLogger().Debug("replace ", replace)
						line = strings.ReplaceAll(line, replace, GetJSONPathValueString(keyname, pathLine))
					}
				} else {
					GetLogger().WithField("key", lookup).Warn("error by getting end position of prefix")
				}
			} else {
				GetLogger().WithField("key", lookup).Warn("error by getting start position of prefix")
			}
		}
	}
	return line
}

// ClearAll removes all entries
func ClearAll() {
	keyValue.Range(func(key, _ interface{}) bool {
		keyValue.Delete(key)
		return true
	})
	taskList.Range(func(key, _ interface{}) bool {
		keyValue.Delete(key)
		return true
	})
}
