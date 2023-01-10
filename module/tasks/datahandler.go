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

	"github.com/swaros/contxt/module/yamc"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// DefaultDataHandler is the default implementation of the DataHandler interface
// It uses the yamc package to store and retrieve data
type DefaultDataHandler struct {
	yamcHndl *yamc.Yamc
}

func NewDefaultDataHandler() *DefaultDataHandler {
	return &DefaultDataHandler{
		yamcHndl: yamc.New(),
	}
}

func (d *DefaultDataHandler) GetJSONPathResult(key, path string) (gjson.Result, bool) {
	if data, err := d.yamcHndl.Gjson(path); err == nil {
		return data, true
	}
	return gjson.Result{}, false
}

func (d *DefaultDataHandler) GetDataAsJson(key string) (string, bool) {
	if data, err := d.yamcHndl.ToString(yamc.NewJsonReader()); err == nil {
		return data, true
	}
	return "", false
}

func (d *DefaultDataHandler) GetDataAsYaml(key string) (string, bool) {
	if data, err := d.yamcHndl.ToString(yamc.NewYamlReader()); err == nil {
		return data, true
	}
	return "", false
}

func (d *DefaultDataHandler) AddJSON(key, jsonString string) error {
	rdr := yamc.New()
	if err := rdr.Parse(yamc.NewJsonReader(), []byte(jsonString)); err != nil {
		return err
	}
	m := rdr.GetData()
	d.updateData(key, m)
	return nil
}

func (d *DefaultDataHandler) updateData(key string, data interface{}) {
	currentData := d.yamcHndl.GetData()
	currentData[key] = data
	d.yamcHndl.SetData(currentData)
}

func (d *DefaultDataHandler) SetJSONValueByPath(key, path, value string) error {
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

func (d *DefaultDataHandler) ImportDataFromYAMLFile(key string, filename string) error {
	ymlYmc, err := yamc.NewByYaml(filename)
	if err != nil {
		return err
	}
	data := ymlYmc.GetData()
	d.updateData(key, data)
	return nil
}

func (d *DefaultDataHandler) AddData(key string, data interface{}) {
	d.updateData(key, data)
}

func (d *DefaultDataHandler) GetData(key string) (interface{}, bool) {
	data := d.yamcHndl.GetData()
	if data != nil {
		if value, ok := data[key]; ok {
			return value, true
		}
	}
	return nil, false
}

func (d *DefaultDataHandler) GetYamc() *yamc.Yamc {
	return d.yamcHndl
}
