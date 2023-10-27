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
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/yamc"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// CombinedDh is a data handler that combines the functionality of the Placeholder Handler and the Datahandler
// It uses the Yamc library to store and retrieve data
type CombinedDh struct {
	yamcHndl           map[string]*yamc.Yamc // these are the data handlers for the different data set by key
	yamcRoot           *yamc.Yamc            // these is used for plain string based placeholders
	openBracket        string                // the brackets used for the placeholders as opening bracket
	closeBracket       string                // the brackets used for the placeholders as closing bracket
	inBracketSeperator string                // the seperator used to get the key for map placeholders
	logger             mimiclog.Logger       // the logger
}

func NewCombinedDataHandler() *CombinedDh {
	dh := &CombinedDh{
		yamcHndl:           make(map[string]*yamc.Yamc), // we use a map for the different data sets
		yamcRoot:           yamc.New(),                  // we use the root yamc for plain string based placeholders
		openBracket:        "${",                        // this is the opening bracket for the placeholders
		closeBracket:       "}",                         // this is the closing bracket for the placeholders
		inBracketSeperator: ":",                         // this is the seperator for the key in the brackets
	}
	return dh
}

func (d *CombinedDh) SetLogger(logger mimiclog.Logger) {
	d.logger = logger
}

func (d *CombinedDh) getLogger() mimiclog.Logger {
	if d.logger == nil {
		return mimiclog.NewNullLogger()
	}
	return d.logger
}

// we need to get a yamc handler by key
// if the key is not present we create a new one
func (d *CombinedDh) getYamcByKey(key string) *yamc.Yamc {
	if _, ok := d.yamcHndl[key]; !ok {
		d.yamcHndl[key] = yamc.New()
	}
	return d.yamcHndl[key]
}

// testing if the key is present
func (d *CombinedDh) ifKeyExists(key string) bool {
	if _, ok := d.yamcHndl[key]; ok {
		return true
	}
	return false
}

// GetJSONPathResult returns the result of a json path query
// if the key is not present it returns an empty result
// the returned value is a gjson.Result
func (d *CombinedDh) GetJSONPathResult(key, path string) (gjson.Result, bool) {
	if !d.ifKeyExists(key) {
		return gjson.Result{}, false
	}
	ymc := d.getYamcByKey(key)
	if data, err := ymc.Gjson(path); err == nil && data.Exists() {
		return data, true
	}
	return gjson.Result{}, false
}

// GetDataAsJson returns the data as json string
// if the key is not present it returns an empty string
func (d *CombinedDh) GetDataAsJson(key string) (string, bool) {
	if !d.ifKeyExists(key) {
		d.getLogger().Debug("GetDataAsJson: key [" + key + "] not found")
		return "", false
	}
	data, err := d.getYamcByKey(key).ToString(yamc.NewJsonReader())
	return data, (err == nil)
}

// GetDataAsYaml returns the data as yaml string
// if the key is not present it returns an empty string
func (d *CombinedDh) GetDataAsYaml(key string) (string, bool) {
	if !d.ifKeyExists(key) {
		d.getLogger().Debug("GetDataAsYaml: key [" + key + "] not found")
		return "", false
	}
	data, err := d.getYamcByKey(key).ToString(yamc.NewYamlReader())
	return data, (err == nil)

}

// AddJSON adds data by parsing a json string
// and store them with the given key
func (d *CombinedDh) AddJSON(key, jsonString string) error {
	if d.getLogger().IsTraceEnabled() {
		d.getLogger().Trace("AddJSON: key [" + key + "] jsonString [" + systools.StringSubLeft(jsonString, 40) + "]")
	}
	ymc := d.getYamcByKey(key)
	return ymc.Parse(yamc.NewJsonReader(), []byte(jsonString))

}

// AddJSON adds data by parsing a json string
// and store them with the given key
func (d *CombinedDh) AddYaml(key, yamlString string) error {
	if d.getLogger().IsTraceEnabled() {
		d.getLogger().Trace("AddYaml: key [" + key + "] yamlString [" + systools.StringSubLeft(yamlString, 40) + "]")
	}
	ymc := d.getYamcByKey(key)
	return ymc.Parse(yamc.NewYamlReader(), []byte(yamlString))
}

// SetJSONValueByPath sets a value by a json path using
// the sjson library
func (d *CombinedDh) SetJSONValueByPath(key, path, value string) error {
	if !d.ifKeyExists(key) {
		d.getLogger().Error("SetJSONValueByPath: key [" + key + "] not found")
		return errors.New("the key [" + key + "] does not exists")
	}
	ymc := d.getYamcByKey(key)
	data := ymc.GetData()
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if newContentString, err := sjson.Set(string(jsonData), path, value); err != nil {
			return err
		} else {
			d.AddJSON(key, newContentString)
			return nil
		}

	}
	return errors.New("error by getting data from " + key)
}

func (d *CombinedDh) ImportDataFromYAMLFile(key string, filename string) error {
	ymlYmc, err := yamc.NewByYaml(filename)
	if err != nil {
		return err
	}
	data := ymlYmc.GetData()
	d.getYamcByKey(key).SetData(data)
	return nil
}

func (d *CombinedDh) AddData(key string, data map[string]interface{}) {
	d.getYamcByKey(key).SetData(data)
}

func (d *CombinedDh) GetData(key string) (map[string]interface{}, bool) {
	if !d.ifKeyExists(key) {
		return nil, false
	}
	return d.getYamcByKey(key).GetData(), true
}

func (d *CombinedDh) GetYamc() *yamc.Yamc {
	return d.yamcRoot
}

func (d *CombinedDh) SetPH(key, value string) {
	if d.getLogger().IsTraceEnabled() {
		d.getLogger().Trace("SetPH: key [" + key + "] value [" + systools.StringSubLeft(value, 40) + " ...]")
	}
	d.yamcRoot.Store(key, value)
}

func (d *CombinedDh) AppendToPH(key, value string) bool {
	return d.yamcRoot.Update(key, func(val interface{}) interface{} {
		if val == nil {
			return value
		}
		if strValue, ok := val.(string); ok {
			return strValue + value
		}
		return value

	})
}

func (d *CombinedDh) SetIfNotExists(key, value string) {
	if _, found := d.yamcRoot.Get(key); !found {
		if d.getLogger().IsTraceEnabled() {
			d.getLogger().Trace("SetIfNotExists: key [" + key + "] value [" + systools.StringSubLeft(value, 40) + " ...]")
		}
		d.yamcRoot.Store(key, value)
	}
}

func (d *CombinedDh) GetPH(key string) string {
	if val, found := d.yamcRoot.Get(key); !found {
		d.getLogger().Warn("GetPH: key [" + key + "] not found")
		return ""
	} else {
		return val.(string)
	}

}

// GetPHExists returns the value of the placeholder
// and a boolean if the placeholder exists
func (d *CombinedDh) GetPHExists(key string) (string, bool) {
	if val, found := d.yamcRoot.Get(key); !found {
		return "", false
	} else {
		return val.(string), true
	}
}

// GetPlaceHoldersFnc iterates over all placeholders and calls the inspectFunc
func (d *CombinedDh) GetPlaceHoldersFnc(inspectFunc func(phKey string, phValue string)) {
	d.yamcRoot.Range(func(key interface{}, value interface{}) bool {
		if strValue, ok := value.(string); ok {
			if keyValue, ok := key.(string); ok {
				if d.logger.IsTraceEnabled() {
					d.getLogger().Trace("PARSE PLACEHOLDERs: key [" + keyValue + "] = [" + systools.StringSubLeft(strValue, 40) + " ...]")
				}
				inspectFunc(keyValue, strValue)
			}
		}
		return true
	})
	// parsing environment variables
	for _, env := range os.Environ() {
		envParts := strings.Split(env, "=")
		if len(envParts) == 2 {
			envKey := envParts[0]
			envValue := envParts[1]
			if d.logger.IsTraceEnabled() {
				d.getLogger().Trace("PARSE ENVIRONMENT VARIABLEs: key [" + envKey + "] = [" + systools.StringSubLeft(envValue, 40) + " ...]")
			}
			inspectFunc(envKey, envValue)
		}
	}
}

// findSeparatorBetweenBrackets finds the separator between brackets
// it returns the left and right part of the separator
// and the end point of the separator
// this is used to find the separator between brackets, what means that this placeholder
// is used to get a value from a map
func (d *CombinedDh) findSeparatorBetweenBrackets(line string, start int) (leftFromSep, rightFromSep string, endPoint int, found bool) {
	line = line[start:]
	leftFromSep = ""
	rightFromSep = ""
	endPoint = -1
	found = false
	preventLongLIne := strings.Contains(line, string(d.openBracket))
	preventLongLIne = preventLongLIne && strings.Contains(line, string(d.closeBracket))
	preventLongLIne = preventLongLIne && strings.Contains(line, string(d.inBracketSeperator))
	if preventLongLIne {
		startMarker := strings.Index(line, string(d.openBracket))
		endMarker := strings.Index(line, string(d.closeBracket))
		if startMarker < endMarker {
			sepMarker := strings.Index(line[startMarker:], string(d.inBracketSeperator))
			if sepMarker > 0 {
				sepMarker += startMarker
			}
			if sepMarker > startMarker && sepMarker < endMarker {
				leftFromSep = line[startMarker+len(d.openBracket) : sepMarker]
				rightFromSep = line[sepMarker+1 : endMarker]
				found = true
				endPoint = endMarker + len(d.closeBracket)
			}
		}
	}
	return
}

// handleMapPlaceHolder handles the map place holder
// e.g. ${map:myMapKey:myMapPath}
// it will replace the placeholder with the value of the map
// if the map does not exist, it will return the original line
func (d *CombinedDh) handleMapPlaceHolder(line string) string {
	start := 0
	maxTry := 100 // prevent endless loop. also defines the max amount of nested brackets
	for {
		maxTry--
		if maxTry < 0 {
			return line
		}
		leftFromSep, rightFromSep, rstart, found := d.findSeparatorBetweenBrackets(line, start)
		start = rstart
		if found {
			if d.ifKeyExists(leftFromSep) {
				ymc := d.getYamcByKey(leftFromSep)
				if newVal, err := ymc.GetGjsonString(rightFromSep); err == nil && newVal != "" {
					line = d.replaceByMapVar(line, leftFromSep, rightFromSep, newVal)
					start = 0 // reset start offset after we found something that changes the line
				}
			}
		} else {
			return line
		}
	}
}

// replaceAllByBrackets replaces all placeholders in the line
// e.g. ${myKey} will be replaced with the value of myKey
// if the key does not exist, it will return the original line
func (d *CombinedDh) replaceAllByBrackets(line string) string {
	d.GetPlaceHoldersFnc(func(phKey string, phValue string) {
		line = d.replaceByVar(line, phKey, phValue)
	})

	line = d.handleMapPlaceHolder(line)

	return line
}

// replaceByVar replaces the placeholder with the value
// e.g. ${myKey} will be replaced with the value of myKey
// if the key does not exist, it will return the original line
func (d *CombinedDh) replaceByVar(line, phKey, replace string) string {
	return strings.ReplaceAll(line, string(d.openBracket)+phKey+string(d.closeBracket), replace)
}

// replaceByMapVar replaces the placeholder with the value
// e.g. ${map:myMapKey:myMapPath} will be replaced with the value of myMapPath
// if the key does not exist, it will return the original line
func (d *CombinedDh) replaceByMapVar(line, id, phKey, replace string) string {
	keyStr := string(d.openBracket) + id + string(d.inBracketSeperator) + phKey
	return strings.ReplaceAll(line, keyStr+string(d.closeBracket), replace)
}

// HandlePlaceHolder handles the placeholders in the line
// e.g. ${myKey} will be replaced with the value of myKey
// if the key does not exist, it will return the original line
// it will also handle the map placeholders
// e.g. ${map:myMapKey:myMapPath}
// it will replace the placeholder with the value of the map
// if the map does not exist, it will return the original line
func (d *CombinedDh) HandlePlaceHolder(line string) string {
	return d.HandlePlaceHolderWithScope(line, make(map[string]string))
}

// HandlePlaceHolderWithScope handles the placeholders in the line
// e.g. ${myKey} will be replaced with the value of myKey
// if the key does not exist, it will return the original line
// it will also handle the map placeholders
// e.g. ${map:myMapKey:myMapPath}
// in addition it will replace the placeholders with the values from the scopeVars
func (d *CombinedDh) HandlePlaceHolderWithScope(line string, scopeVars map[string]string) string {
	line = d.replaceAllByBrackets(line)
	for key, value := range scopeVars {
		line = d.replaceByVar(line, key, value)
	}
	return line
}

// ClearAll clears all the data
func (d *CombinedDh) ClearAll() {
	d.yamcRoot.Reset()
}

func (d *CombinedDh) ExportVarToFile(variable string, filename string) error {
	strData, exists := d.GetPHExists(variable)
	if strData == "" || !exists {
		if !exists {
			return errors.New("variable '" + variable + "' can not be used for export to file. this variable not exists")
		}
		return errors.New("variable '" + variable + "' can not be used for export to file. variable is empty")
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err2 := f.WriteString(d.HandlePlaceHolder(strData)); err != nil {
		return err2
	}

	return nil
}

func (d *CombinedDh) GetDataKeys() []string {
	data := d.yamcRoot.GetData()
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}
