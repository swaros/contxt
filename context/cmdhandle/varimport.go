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

	"github.com/tidwall/gjson"

	"github.com/swaros/contxt/context/output"

	"github.com/sirupsen/logrus"

	"github.com/Masterminds/sprig/v3"
	"github.com/ghodss/yaml"
)

const (
	inlineCmdSep = " "
	startMark    = "#@"
	inlineMark   = "#@-"
	iterateMark  = "#@foreach"
	endMark      = "#@end"
)

// TryParse to parse a line and set a value depending on the line command
func TryParse(script []string) (bool, []string) {
	inIteration := false
	var iterationLines []string
	var parsedScript []string
	var iterationCollect gjson.Result
	for _, line := range script {
		if len(line) > len(startMark) && line[0:len(startMark)] == startMark {
			parts := strings.Split(line, inlineCmdSep)
			GetLogger().WithField("keywords", parts).Debug("try to parse parts")
			if len(parts) < 1 {
				parsedScript = append(parsedScript, line)
				continue
			}
			switch parts[0] {

			case inlineMark:
				if inIteration {
					iterationLines = append(iterationLines, strings.Replace(line, inlineMark+" ", "", 4))
					GetLogger().WithField("code", iterationLines).Debug("append to subscript")
				} else {
					output.Error("invalid usage", inlineMark, " only valid while in iteration")
				}
				break
			case endMark:
				GetLogger().Debug("ITERATION: DONE")
				if inIteration {
					inIteration = false
					iterationCollect.ForEach(func(key gjson.Result, value gjson.Result) bool {
						var parsedExecLines []string
						for _, iLine := range iterationLines {
							iLine = strings.Replace(iLine, "__LINE__", value.String(), 1)
							parsedExecLines = append(parsedExecLines, iLine)
						}
						GetLogger().WithFields(logrus.Fields{
							"key":       key,
							"value":     value,
							"subscript": parsedExecLines,
						}).Debug("... delegate script")
						_, subs := TryParse(parsedExecLines)
						for _, subLine := range subs {
							parsedScript = append(parsedScript, subLine)
						}
						return true
					})
				}

			case iterateMark:
				if len(parts) == 3 {
					impMap, found := GetJSONPathResult(parts[1], parts[2])
					if !found {
						output.Error("undefined data from path", parts[1], parts[2])
					} else {
						inIteration = true
						iterationCollect = impMap
						GetLogger().WithField("data", impMap).Debug("ITERATION: START")
					}
				} else {
					output.Error("invalid arguments", "#@iterate needs <name-of-import> <path-to-data>")
				}
				break
			default:
				GetLogger().WithField("unknown", parts[0]).Error("there is no command exists")
			}
		} else {
			parsedScript = append(parsedScript, line)
		}
	}
	GetLogger().WithFields(logrus.Fields{
		"parsed": parsedScript,
	}).Debug("... parsed result")
	return false, parsedScript
}

func handleImport(filename, path string) {

}

// YAMLToMap Convert yaml source string into map
func YAMLToMap(source string) (map[string]interface{}, error) {
	jsond, jerr := yaml.YAMLToJSON([]byte(source))
	if jerr != nil {
		return nil, jerr
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsond), &m); err != nil {
		return nil, err
	}
	return m, nil
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
	if GetLogger().IsLevelEnabled(logrus.TraceLevel) {
		traceMap(m, filename)
	}
	return m, nil

}

// ImportJSONFile imports a json file for reading
func ImportJSONFile(fileName string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		GetLogger().Error(err)
		return nil, err
	}
	m := make(map[string]interface{})
	err = json.Unmarshal([]byte(data), &m)
	if err != nil {
		GetLogger().Error("ImportJSONFile : Unmarshal :", fileName, " : ", err)
		return nil, err
	}
	if GetLogger().IsLevelEnabled(logrus.TraceLevel) {
		traceMap(m, fileName)
	}
	return m, nil

}

func traceMap(mapShow map[string]interface{}, add string) {
	for k, v := range mapShow {
		//mapShow[k] = v
		//GetLogger().WithField("VAR", v).Trace("imported placeholder from " + add + " " + k)
		GetLogger().WithFields(logrus.Fields{
			"source":  add,
			"key":     k,
			"content": v,
		}).Trace("imported content")
	}
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
		GetLogger().WithField("folder", path).Debug("process path")
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
				GetLogger().WithField("file", path).Debug("parsing included file (JSON)")
				jsonMap, loaderr = ImportJSONFile(path)
				hit = true
				break
			case ".yaml", ".yml":
				GetLogger().WithField("file", path).Debug("parsing included file (YAML)")
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
	GetLogger().WithField("file", filename).Debug("import file template")
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
	funcMap := MergeVariableMap(tf, sprig.FuncMap())
	tpl := template.New("contxt-map-string-func").Funcs(funcMap)
	tt, err := tpl.Parse(tmpl)
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
