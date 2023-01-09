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

type CombinedDh struct {
	yamcHndl *yamc.Yamc
	brackets string
}

func NewCombinedDataHandler() *CombinedDh {
	dh := &CombinedDh{
		yamcHndl: yamc.New(),
		brackets: "{}",
	}
	return dh
}

func (d *CombinedDh) GetJSONPathResult(key, path string) (gjson.Result, bool) {

	if data, err := d.yamcHndl.Gjson(path); err == nil {
		return data, true
	}
	return gjson.Result{}, false
}

func (d *CombinedDh) GetDataAsJson(key string) (bool, string) {
	if data, err := d.yamcHndl.ToString(yamc.NewJsonReader()); err == nil {
		return true, data
	}
	return false, ""
}

func (d *CombinedDh) GetDataAsYaml(key string) (bool, string) {
	if data, err := d.yamcHndl.ToString(yamc.NewYamlReader()); err == nil {
		return true, data
	}
	return false, ""
}

func (d *CombinedDh) AddJSON(key, jsonString string) error {
	rdr := yamc.New()
	if err := rdr.Parse(yamc.NewJsonReader(), []byte(jsonString)); err != nil {
		return err
	}
	m := rdr.GetData()
	d.updateData(key, m)
	return nil
}

func (d *CombinedDh) updateData(key string, data interface{}) {
	currentData := d.yamcHndl.GetData()
	currentData[key] = data
	d.yamcHndl.SetData(currentData)
}

func (d *CombinedDh) SetJSONValueByPath(key, path, value string) error {
	data := d.yamcHndl.GetData()
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if newContent, err := sjson.Set(string(jsonData), path, value); err != nil {
			return err
		} else {
			d.updateData(key, newContent)
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
	d.updateData(key, data)
	return nil
}

func (d *CombinedDh) AddData(key string, data interface{}) {
	d.updateData(key, data)
}

func (d *CombinedDh) GetData(key string) (interface{}, bool) {
	data := d.yamcHndl.GetData()
	if data != nil {
		if value, ok := data[key]; ok {
			return value, true
		}
	}
	return nil, false
}

func (d *CombinedDh) GetYamc() *yamc.Yamc {
	return d.yamcHndl
}

func (d *CombinedDh) SetPH(key, value string) {
	d.yamcHndl.Store(key, value)
}

func (d *CombinedDh) AppendToPH(key, value string) bool {
	return d.yamcHndl.Update(key, func(val interface{}) interface{} {
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
	if _, found := d.yamcHndl.Get(key); !found {
		d.yamcHndl.Store(key, value)
	}
}

func (d *CombinedDh) GetPH(key string) string {
	if val, found := d.yamcHndl.Get(key); !found {
		return ""
	} else {
		return val.(string)
	}

}

func (d *CombinedDh) GetPHExists(key string) (string, bool) {
	if val, found := d.yamcHndl.Get(key); !found {
		return "", false
	} else {
		return val.(string), true
	}
}

func (d *CombinedDh) GetPlaceHoldersFnc(inspectFunc func(phKey string, phValue string)) {
	d.yamcHndl.Range(func(key interface{}, value interface{}) bool {
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
	d.yamcHndl.Reset()
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
