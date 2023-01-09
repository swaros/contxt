// Copyright (c) 2022 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
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

	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/yamc"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// CombinedDh is a data handler that combines the functionality of the Placeholder Handler and the Datahandler
// It uses the Yamc library to store and retrieve data
type CombinedDh struct {
	yamcHndl map[string]*yamc.Yamc // these are the data handlers for the different data set by key
	yamcRoot *yamc.Yamc            // these is used for plain string based placeholders
	brackets string                // the brackets used for the placeholders
}

func NewCombinedDataHandler() *CombinedDh {
	dh := &CombinedDh{
		yamcHndl: make(map[string]*yamc.Yamc),
		yamcRoot: yamc.New(),
		brackets: "{}",
	}
	return dh
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
func (d *CombinedDh) GetDataAsJson(key string) (bool, string) {
	if !d.ifKeyExists(key) {
		return false, ""
	}
	if data, err := d.getYamcByKey(key).ToString(yamc.NewJsonReader()); err == nil {
		return true, data
	}
	return false, ""
}

// GetDataAsYaml returns the data as yaml string
// if the key is not present it returns an empty string
func (d *CombinedDh) GetDataAsYaml(key string) (bool, string) {
	if data, err := d.getYamcByKey(key).ToString(yamc.NewYamlReader()); err == nil {
		return true, data
	}
	return false, ""
}

// AddJSON adds data by parsing a json string
// and store them with the given key
func (d *CombinedDh) AddJSON(key, jsonString string) error {
	ymc := d.getYamcByKey(key)
	if err := ymc.Parse(yamc.NewJsonReader(), []byte(jsonString)); err != nil {
		return err
	}
	return nil
}

// SetJSONValueByPath sets a value by a json path using
// the sjson library
func (d *CombinedDh) SetJSONValueByPath(key, path, value string) error {
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

func (d *CombinedDh) AddData(key string, subkey string, data interface{}) {
	ymc := d.getYamcByKey(key)
	dataStored := ymc.GetData()
	if dataStored == nil {
		dataStored = make(map[string]interface{})
	}
	dataStored[subkey] = data
	ymc.SetData(dataStored)
}

func (d *CombinedDh) GetData(key string) (interface{}, bool) {
	data := d.getYamcByKey(key).GetData()
	if data != nil {
		return data, true
	}
	return nil, false
}

func (d *CombinedDh) GetDataSub(key, subKey string) (interface{}, bool) {
	data := d.getYamcByKey(key).GetData()
	if data != nil {
		if subData, ok := data[subKey]; ok {
			return subData, true
		}
	}
	return nil, false
}

func (d *CombinedDh) GetYamc() *yamc.Yamc {
	return d.yamcRoot
}

func (d *CombinedDh) SetPH(key, value string) {
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
		d.yamcRoot.Store(key, value)
	}
}

func (d *CombinedDh) GetPH(key string) string {
	if val, found := d.yamcRoot.Get(key); !found {
		return ""
	} else {
		return val.(string)
	}

}

func (d *CombinedDh) GetPHExists(key string) (string, bool) {
	if val, found := d.yamcRoot.Get(key); !found {
		return "", false
	} else {
		return val.(string), true
	}
}

func (d *CombinedDh) GetPlaceHoldersFnc(inspectFunc func(phKey string, phValue string)) {
	d.yamcRoot.Range(func(key interface{}, value interface{}) bool {
		if strValue, ok := value.(string); ok {
			if keyValue, ok := key.(string); ok {
				inspectFunc(keyValue, strValue)
			}
		}
		return true
	})
}

func (d *CombinedDh) replaceAllByBrackets(line string) string {
	d.GetPlaceHoldersFnc(func(phKey string, phValue string) {
		line = d.replaceByVar(line, phKey, phValue)
	})
	return line
}

func (d *CombinedDh) replaceByVar(line, phKey, replace string) string {
	if systools.StrLen(d.brackets) != 2 {
		panic("brackets must be 2 chars")
	}
	return strings.ReplaceAll(line, string(d.brackets[0])+phKey+string(d.brackets[1]), replace)
}

func (d *CombinedDh) HandlePlaceHolder(line string) string {
	return d.HandlePlaceHolderWithScope(line, make(map[string]string))
}

func (d *CombinedDh) HandlePlaceHolderWithScope(line string, scopeVars map[string]string) string {
	if systools.StrLen(d.brackets) != 2 {
		return line
	}
	line = d.replaceAllByBrackets(line)
	for key, value := range scopeVars {
		line = d.replaceByVar(line, key, value)
	}
	return line
}

func (d *CombinedDh) ClearAll() {
	d.yamcRoot.Reset()
}

func (d *CombinedDh) ExportVarToFile(variable string, filename string) error {
	strData := d.GetPH(variable)
	if strData == "" {
		return errors.New("variable " + variable + " can not be used for export to file. not exists or empty")
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
