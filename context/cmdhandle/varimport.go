package cmdhandle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
)

const (
	inlineCmdSep = "::"
	startMark    = "@"
)

// Todo just for example
type Todo struct {
	Name        string
	Description string
}

// Demo is just to try templates
func Demo() string {
	td := Todo{"test1", "test1 is replaced"}
	t, err := template.New("todos").Parse("name: \"{{ .Name}}\" with description: \"{{ .Description}}\"")
	if err != nil {
		panic(err)
	}
	out := new(bytes.Buffer)
	err = t.Execute(out, td)
	if err != nil {
		panic(err)
	}
	return out.String()
}

// TryParse to parse a line and set a value depending on the line command
func TryParse(line string) bool {
	if line[0:1] == startMark {
		parts := strings.Split(line, inlineCmdSep)
		if len(parts) < 3 {
			return false
		}

		switch parts[0] {
		case "@import":
			handleImport(parts[1], parts[2])
			return true

		}

	}
	return false
}

func handleImport(filename, path string) {

}

// ImportYAMLFile imports a yaml file as used for json map
func ImportYAMLFile(filename string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	jsond, jerr := yaml.YAMLToJSON(data)
	if jerr != nil {
		return nil, jerr
	}
	m := make(map[string]interface{})
	if err = json.Unmarshal([]byte(jsond), &m); err != nil {
		return nil, err
	}
	return m, nil

}

// ImportJSONFile imports a json file for reading
func ImportJSONFile(fileName string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	if err = json.Unmarshal([]byte(data), &m); err != nil {
		return nil, err
	}
	return m, nil

}

// ImportFileContent imports a file and returns content as string
func ImportFileContent(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("File reading error", err)
		return "", err
	}
	return string(data), nil
}

// HandleJSONMap parsing json content for text/template
func HandleJSONMap(tmpl string, m map[string]interface{}) (string, error) {
	tf := template.FuncMap{
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
	t := template.New("contxt-vars").Funcs(tf)
	tt, err := t.Parse(tmpl)
	if err != nil {
		panic(err)
	}
	out := new(bytes.Buffer)
	tt.Execute(out, &m)
	if err != nil {
		return "", err
	}
	return out.String(), nil

}
