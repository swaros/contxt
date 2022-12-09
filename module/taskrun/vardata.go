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

package taskrun

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var dataStorage sync.Map

// AddData adds a Data Map to the storage
func AddData(key string, data map[string]interface{}) {
	dataStorage.Store(key, data)
}

// ImportDataFromJSONFile imports a map from a json file and assign it to a key
func ImportDataFromJSONFile(key string, filename string) error {
	data, err := ImportJSONFile(filename)
	if err != nil {
		return err
	}
	GetLogger().WithFields(logrus.Fields{"key": key, "file": filename, "value": data}).Trace("variables import")
	AddData(key, data)
	return nil
}

// GetJSONPathValueString returns the value depending key and path as string
func GetJSONPathValueString(key, path string) string {
	ok, data := GetData(key)
	if ok && data != nil {
		jsonData, err := json.Marshal(data)
		if err == nil {
			value := gjson.Get(string(jsonData), path)
			GetLogger().WithFields(logrus.Fields{"key": key, "path": path, "value": value}).Debug("placeholder: found map entrie")
			return value.String()
		}
		GetLogger().WithField("key", key).Error("placeholder: error while marshal data")
	} else {
		GetLogger().WithField("key", key).Error("placeholder: error by getting data from named map")
	}
	return ""
}

func SetJSONValueByPath(key, path, value string) error {
	ok, data := GetData(key)
	if ok && data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if newContent, err := sjson.Set(string(jsonData), path, value); err != nil {
			return err
		} else {
			if err2 := AddJSON(key, newContent); err2 != nil {
				return err2
			}
			return nil
		}

	}
	return errors.New("error by getting data from " + key)

}

// GetJSONPathResult returns the value depending key and path as string
func GetJSONPathResult(key, path string) (gjson.Result, bool) {
	ok, data := GetData(key)
	if ok && data != nil {
		//mapdata := make(map[string]interface{})
		jsonData, err := json.Marshal(data)
		if err == nil {
			value := gjson.Get(string(jsonData), path)
			GetLogger().WithFields(logrus.Fields{
				"key":   key,
				"path":  path,
				"value": value.Value()}).Debug("GetJSONPathResult: found map entrie")
			return value, true
		}
		GetLogger().WithField("key", key).Error("GetJSONPathResult: error while marshal data")
	} else {
		GetLogger().WithField("key", key).Error("GetJSONPathResult: error by getting data from named map")
	}
	return gjson.Result{
		Index: 0,
	}, false
}

// ImportDataFromYAMLFile imports a map from a json file and assign it to a key
func ImportDataFromYAMLFile(key string, filename string) error {
	data, err := ImportYAMLFile(filename)
	if err != nil {
		return err
	}

	AddData(key, data)
	return nil
}

// AddJSON imports data by a json String
func AddJSON(key, jsonString string) error {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonString), &m)
	if err != nil {
		return err
	}
	AddData(key, m)
	return nil
}

// GetData returns a data Map by the key.
// or nil if nothing is stored
func GetData(key string) (bool, map[string]interface{}) {
	result, ok := dataStorage.Load(key)
	if ok {
		return ok, result.(map[string]interface{})
	}
	return false, nil
}

// GetDataAsYaml converts the map given by key infto a yaml string
func GetDataAsYaml(key string) (bool, string) {
	if found, data := GetData(key); found {
		if outData, err := yaml.Marshal(data); err == nil {
			return true, string(outData)
		}
	}
	return false, ""
}

// GetDataAsYaml converts the map given by key infto a yaml string
func GetDataAsJson(key string) (bool, string) {
	if found, data := GetData(key); found {
		if outData, err := json.MarshalIndent(data, "", "  "); err == nil {
			return true, string(outData)
		}
	}
	return false, ""
}

// ClearAllData removes all entries
func ClearAllData() {
	dataStorage.Range(func(key, _ interface{}) bool {
		dataStorage.Delete(key)
		return true
	})
}

// GetDataKeys returns all current keys
func GetDataKeys() []string {
	var keys []string
	dataStorage.Range(func(key, _ interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}
