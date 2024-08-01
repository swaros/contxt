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
	"bytes"
	"reflect"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/imdario/mergo"
	"github.com/swaros/contxt/module/mimiclog"
)

type CtxTemplate struct {
	data   map[string]interface{}
	logger mimiclog.Logger
}

func NewCtxTemplate() *CtxTemplate {
	tpl := &CtxTemplate{
		data:   make(map[string]interface{}),
		logger: mimiclog.NewNullLogger(),
	}

	return tpl
}

func (c *CtxTemplate) SetLogger(logger mimiclog.Logger) {
	c.logger = logger
}

func (c *CtxTemplate) SetData(data map[string]interface{}) {
	c.data = data
}

func (c *CtxTemplate) AddDataValue(key string, value interface{}) {
	if c.data == nil {
		c.logger.Debug("Creating new data map by adding value")
		c.data = make(map[string]interface{})
	}
	c.data[key] = value
}

func (c *CtxTemplate) AddDataMap(m map[string]interface{}) {
	if c.data == nil {
		c.logger.Debug("Creating new data map by adding map")
		c.data = make(map[string]interface{})
	}
	for k, v := range m {
		c.data[k] = v
	}
}

func (c *CtxTemplate) ParseTemplate(tmpl *template.Template) (string, error) {
	if c.data == nil {
		c.logger.Debug("Creating new data map by parsing template")
		c.data = make(map[string]interface{})
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c.data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (c *CtxTemplate) getCtxFuncMap() template.FuncMap {
	return template.FuncMap{
		"isInt": func(i interface{}) bool {
			v := reflect.ValueOf(i)
			switch v.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
				return true
			default:
				return false
			}
		},
		"isString": func(i interface{}) bool {
			v := reflect.ValueOf(i)
			switch v.Kind() {
			case reflect.String:
				return true
			default:
				return false
			}
		},
		"isSlice": func(i interface{}) bool {
			v := reflect.ValueOf(i)
			switch v.Kind() {
			case reflect.Slice:
				return true
			default:
				return false
			}
		},
		"isArray": func(i interface{}) bool {
			v := reflect.ValueOf(i)
			switch v.Kind() {
			case reflect.Array:
				return true
			default:
				return false
			}
		},
		"isMap": func(i interface{}) bool {
			v := reflect.ValueOf(i)
			switch v.Kind() {
			case reflect.Map:
				return true
			default:
				return false
			}
		},
	}

}

func (c *CtxTemplate) ParseTemplateString(tmpl string) (string, error) {

	tf := c.getCtxFuncMap()
	funcMap, _ := MergeVariableMap(tf, sprig.FuncMap())
	tpl := template.New("contxt-functions").Funcs(funcMap)
	tt, err := tpl.Parse(tmpl)
	if err != nil {
		return "", err
	}
	out := new(bytes.Buffer)
	tErr := tt.Execute(out, &c.data)
	if tErr != nil {
		return "", err
	}
	return out.String(), nil

}

// MergeVariableMap merges two maps
// this is a global function that is used in different places
func MergeVariableMap(mapin map[string]interface{}, maporigin map[string]interface{}) (map[string]interface{}, error) {
	if err := mergo.Merge(&maporigin, mapin, mergo.WithOverride); err != nil {
		return nil, err
	}
	return maporigin, nil
}
