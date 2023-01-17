package ctemplate

import (
	"bytes"
	"reflect"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/imdario/mergo"
)

type CtxTemplate struct {
	data map[string]interface{}
}

func NewCtxTemplate() *CtxTemplate {
	return &CtxTemplate{
		data: make(map[string]interface{}),
	}
}

func (c *CtxTemplate) SetData(data map[string]interface{}) {
	c.data = data
}

func (c *CtxTemplate) AddDataValue(key string, value interface{}) {
	if c.data == nil {
		c.data = make(map[string]interface{})
	}
	c.data[key] = value
}

func (c *CtxTemplate) AddDataMap(m map[string]interface{}) {
	if c.data == nil {
		c.data = make(map[string]interface{})
	}
	for k, v := range m {
		c.data[k] = v
	}
}

func (c *CtxTemplate) ParseTemplate(tmpl *template.Template) (string, error) {
	if c.data == nil {
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
	tt.Execute(out, &c.data)
	if err != nil {
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
