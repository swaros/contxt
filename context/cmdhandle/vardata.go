package cmdhandle

import (
	"encoding/json"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/tidwall/gjson"
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

// ClearAllData removes all entries
func ClearAllData() {
	dataStorage.Range(func(key, value interface{}) bool {
		dataStorage.Delete(key)
		return true
	})
}

// GetDataKeys returns all current keys
func GetDataKeys() []string {
	var keys []string
	dataStorage.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}
