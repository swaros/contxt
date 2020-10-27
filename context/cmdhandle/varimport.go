package cmdhandle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
)

const (
	inlineCmdSep = "::"
	startMark    = "@"
)

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

// MergeVariableMap merges two maps
func MergeVariableMap(mapin map[string]interface{}, maporigin map[string]interface{}) map[string]interface{} {
	for k, v := range mapin {
		maporigin[k] = v
	}
	return maporigin
}

// ImportFolders import a list of folders recusiv
func ImportFolders(templatePath string, paths ...string) (string, error) {
	var mapOrigin map[string]interface{}
	mapOrigin = make(map[string]interface{})

	template, terr := ImportFileContent(templatePath)
	if terr != nil {
		return "", terr
	}

	for _, path := range paths {
		pathMap, parseErr := ImportFolder(path, templatePath)
		if parseErr != nil {
			return "", parseErr
		}
		mapOrigin = MergeVariableMap(pathMap, mapOrigin)
	}
	result, herr := HandleJSONMap(template, mapOrigin)
	if herr != nil {
		return "", herr
	}
	template = result

	return template, nil
}

// ImportFolder reads folder recusiv and reads all .json, .yml and .yaml files
func ImportFolder(path string, templatePath string) (map[string]interface{}, error) {

	var mapOrigin map[string]interface{}
	mapOrigin = make(map[string]interface{})

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var jsonMap map[string]interface{}
		var loaderr error
		hit := false
		if !info.IsDir() {
			var extension = filepath.Ext(path)
			var basename = filepath.Base(path)
			if basename == ".contxt.yml" {
				return nil
			}
			switch extension {
			case ".json":
				jsonMap, loaderr = ImportJSONFile(path)
				hit = true
				break
			case ".yaml", ".yml":
				jsonMap, loaderr = ImportYAMLFile(path)
				hit = true
				break
			}
			if loaderr != nil {
				return loaderr
			}
			if hit {
				mapOrigin = MergeVariableMap(jsonMap, mapOrigin)
			}
		}

		return nil
	})

	return mapOrigin, err
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
		return "", err
	}
	out := new(bytes.Buffer)
	tt.Execute(out, &m)
	if err != nil {
		return "", err
	}
	return out.String(), nil

}

/*
func IsList(i interface{}) bool {
	v := reflect.ValueOf(i).Kind()
	return v == reflect.Array || v == reflect.Slice
}

func IsNumber(i interface{}) bool {
	v := reflect.ValueOf(i).Kind()
	switch v {
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func IsInt(i interface{}) bool {
	v := reflect.ValueOf(i).Kind()
	switch v {
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func IsFloat(i interface{}) bool {
	v := reflect.ValueOf(i).Kind()
	return v == reflect.Float32 || v == reflect.Float64
}
*/
