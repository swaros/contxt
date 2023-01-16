// Copyright (c) 2023 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// # Licensed under the MIT License
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
package tasks

import (
	"github.com/swaros/contxt/module/configure"
	"github.com/tidwall/gjson"
)

// DataMapHandler is the interface for the data handlers
// they are used to store and retrieve data
type DataMapHandler interface {
	GetJSONPathResult(key, path string) (gjson.Result, bool) // returns the result of a json path query
	GetDataAsJson(key string) (string, bool)                 // returns the data as json string
	GetDataAsYaml(key string) (string, bool)                 // returns the data as yaml string
	AddJSON(key, jsonString string) error                    // adds data by parsing a json string
	SetJSONValueByPath(key, path, value string) error        // sets a value by a json path using
}

// PlaceHolder is the interface for the placeholder handler
// they are used to store and retrieve placeholder
type PlaceHolder interface {
	SetPH(key, value string)                                                    // sets a placeholder and its value
	AppendToPH(key, value string) bool                                          // appends a value to a placeholder
	SetIfNotExists(key, value string)                                           // sets a placeholder and its value if it does not exist
	GetPHExists(key string) (string, bool)                                      // returns the value of a placeholder if it exists
	GetPH(key string) string                                                    // returns the value of a placeholder if it exists, otherwise it returns the placeholder itself
	GetPlaceHoldersFnc(inspectFunc func(phKey string, phValue string))          // iterates over all placeholders and calls the inspectFunc for each placeholder
	HandlePlaceHolder(line string) string                                       // replaces all placeholders in a string
	HandlePlaceHolderWithScope(line string, scopeVars map[string]string) string // replaces all placeholders in a string with a scope
	ClearAll()                                                                  // clears all placeholders
	ExportVarToFile(variable string, filename string) error                     // exports a placeholder to a file
}

// MainCmdSetter is the interface for the main command setter
// they are used to set the main command and its arguments

type MainCmdSetter interface {
	GetMainCmd(cfg configure.Options) (string, []string)
}
