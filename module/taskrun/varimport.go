// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
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
package taskrun

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/imdario/mergo"
	"github.com/tidwall/gjson"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/manout"

	"github.com/sirupsen/logrus"

	"github.com/Masterminds/sprig/v3"
	"github.com/ghodss/yaml"
)

const (
	inlineCmdSep    = " "
	startMark       = "#@"
	inlineMark      = "#@-"
	iterateMark     = "#@foreach"
	endMark         = "#@end"
	fromJSONMark    = "#@import-json"
	fromJSONCmdMark = "#@import-json-exec"
	parseVarsMark   = "#@var"
	setvarMark      = "#@set"
	setvarInMap     = "#@set-in-map"
	exportToYaml    = "#@export-to-yaml"
	exportToJson    = "#@export-to-json"
	addvarMark      = "#@add"
	equalsMark      = "#@if-equals"
	notEqualsMark   = "#@if-not-equals"
	osCheck         = "#@if-os"
	codeLinePH      = "__LINE__"
	codeKeyPH       = "__KEY__"
	writeVarToFile  = "#@var-to-file"
)

// TryParse to parse a line and set a value depending on the line command
func TryParse(script []string, regularScript func(string) (bool, int)) (bool, int, []string) {
	inIteration := false
	inIfState := false
	ifState := true
	var iterationLines []string
	var parsedScript []string
	var iterationCollect gjson.Result
	for _, line := range script {
		line = HandlePlaceHolder(line)
		if len(line) > len(startMark) && line[0:len(startMark)] == startMark {
			//parts := strings.Split(line, inlineCmdSep)
			parts := SplitQuoted(line, inlineCmdSep)

			GetLogger().WithField("keywords", parts).Debug("try to parse parts")
			if len(parts) < 1 {
				continue
			}
			switch parts[0] {

			case osCheck:
				if !inIfState {
					if len(parts) == 2 {
						leftEq := parts[1]
						rightEq := configure.GetOs()
						inIfState = true
						ifState = leftEq == rightEq
					} else {
						manout.Error("invalid usage ", equalsMark, " need: str1 str2 ")
					}
				} else {
					manout.Error("invalid usage ", equalsMark, " can not be used in another if")
				}

			case equalsMark:
				if !inIfState {
					if len(parts) == 3 {
						leftEq := parts[1]
						rightEq := parts[2]
						inIfState = true
						ifState = leftEq == rightEq
						GetLogger().WithFields(logrus.Fields{"condition": ifState, "left": leftEq, "right": rightEq}).Debug(equalsMark)
					} else {
						manout.Error("invalid usage ", equalsMark, " need: str1 str2 (got:", len(parts), ")")
					}
				} else {
					manout.Error("invalid usage ", equalsMark, " can not be used in another if")
				}

			case notEqualsMark:
				if !inIfState {
					if len(parts) == 3 {
						leftEq := parts[1]
						rightEq := parts[2]
						inIfState = true
						ifState = leftEq != rightEq
						GetLogger().WithFields(logrus.Fields{"condition": ifState, "left": leftEq, "right": rightEq}).Debug(notEqualsMark)
					} else {
						manout.Error("invalid usage ", notEqualsMark, " need: str1 str2 (got:", len(parts), ")")
					}
				} else {
					manout.Error("invalid usage ", equalsMark, " can not be used in another if")
				}

			case inlineMark:
				if inIteration {
					iterationLines = append(iterationLines, strings.Replace(line, inlineMark+" ", "", 4))
					GetLogger().WithField("code", iterationLines).Debug("append to subscript")
				} else {
					manout.Error("invalid usage", inlineMark, " only valid while in iteration")
				}

			case fromJSONMark:
				if len(parts) == 3 {
					err := AddJSON(parts[1], parts[2])
					if err != nil {
						manout.Error("import from json string failed", parts[2], err)
					}
				} else {
					manout.Error("invalid usage", fromJSONMark, " needs 2 arguments. <keyname> <json-source-string>")
				}

			case fromJSONCmdMark:
				if len(parts) >= 3 {
					returnValue := ""
					restSlice := parts[2:]
					keyname := parts[1]
					cmd := strings.Join(restSlice, " ")
					GetLogger().WithFields(logrus.Fields{"key": keyname, "cmd": restSlice}).Info("execute for import-json-exec")
					//GetLogger().WithField("slice", restSlice).Info("execute for import-json-exec")
					exec, args := GetExecDefaults()
					execCode, realExitCode, execErr := ExecuteScriptLine(exec, args, cmd, func(output string, e error) bool {
						returnValue = returnValue + output
						GetLogger().WithField("cmd-output", output).Info("result of command")
						return true
					}, func(proc *os.Process) {
						GetLogger().WithField("import-json-proc", proc).Trace("import-json-process")
					})

					if execErr != nil {
						GetLogger().WithFields(logrus.Fields{
							"intern":       execCode,
							"process-exit": realExitCode,
							"key":          keyname,
							"cmd":          restSlice}).Error("execute for import-json-exec failed")
					} else {

						err := AddJSON(keyname, returnValue)
						if err != nil {
							GetLogger().WithField("error-on-parsing-string", returnValue).Debug("result of command")
							manout.Error("import from json string failed", err, ' ', systools.StringSubLeft(returnValue, 75))
						}
					}
				} else {
					manout.Error("invalid usage", fromJSONCmdMark, " needs 2 arguments at least. <keyname> <bash-command>")
				}
			case addvarMark:
				if len(parts) >= 2 {
					setKeyname := parts[1]
					setValue := strings.Join(parts[2:], " ")
					if ok := AppendToPH(setKeyname, HandlePlaceHolder(setValue)); !ok {
						manout.Error("variable must exists for add ", addvarMark, " ", setKeyname)
					}
				} else {
					manout.Error("invalid usage", setvarMark, " needs 2 arguments at least. <keyname> <value>")
				}
			case setvarMark:
				if len(parts) >= 2 {
					setKeyname := parts[1]
					setValue := strings.Join(parts[2:], " ")
					SetPH(setKeyname, HandlePlaceHolder(setValue))
				} else {
					manout.Error("invalid usage", setvarMark, " needs 2 arguments at least. <keyname> <value>")
				}
			case setvarInMap:
				if len(parts) >= 3 {
					mapName := parts[1]
					path := parts[2]
					setValue := strings.Join(parts[3:], " ")
					if err := SetJSONValueByPath(mapName, path, setValue); err != nil {
						manout.Error(err.Error())
					}
				} else {
					manout.Error("invalid usage", setvarInMap, " needs 3 arguments at least. <mapName> <json.path> <value>")
				}
			case writeVarToFile:
				if len(parts) == 3 {
					varName := parts[1]
					fileName := parts[2]
					ExportVarToFile(varName, fileName)
				} else {
					manout.Error("invalid usage", writeVarToFile, " needs 2 arguments at least. <variable> <filename>")
				}
			case exportToJson:
				if len(parts) == 3 {
					mapKey := parts[1]
					varName := parts[2]
					if exists, newStr := GetDataAsJson(mapKey); exists {
						SetPH(varName, HandlePlaceHolder(newStr))
					} else {
						manout.Error("map with key ", mapKey, " not exists")
					}
				} else {
					manout.Error("invalid usage", exportToJson, " needs 2 arguments at least. <map-key> <variable>")
				}
			case exportToYaml:
				if len(parts) == 3 {
					mapKey := parts[1]
					varName := parts[2]
					if exists, newStr := GetDataAsYaml(mapKey); exists {
						SetPH(varName, HandlePlaceHolder(newStr))
					} else {
						manout.Error("map with key ", mapKey, " not exists")
					}
				} else {
					manout.Error("invalid usage", exportToYaml, " needs 2 arguments at least. <map-key> <variable>")
				}
			case parseVarsMark:
				if len(parts) >= 2 {
					var returnValues []string
					restSlice := parts[2:]
					cmd := strings.Join(restSlice, " ")
					exec, args := GetExecDefaults()
					internalCode, cmdCode, errorFromCm := ExecuteScriptLine(exec, args, cmd, func(output string, e error) bool {
						if e == nil {
							returnValues = append(returnValues, output)
						}
						return true

					}, func(proc *os.Process) {
						GetLogger().WithField(parseVarsMark, proc).Trace("sub process")
					})

					if internalCode == systools.ExitOk && errorFromCm == nil && cmdCode == 0 {
						GetLogger().WithField("values", returnValues).Trace("got values")
						SetPH(parts[1], HandlePlaceHolder(strings.Join(returnValues, "\n")))
					} else {
						GetLogger().WithFields(logrus.Fields{
							"returnCode": cmdCode,
							"error":      errorFromCm.Error,
						}).Error("subcommand failed.")
						manout.Error("Subcommand failed", cmd, " ... was used to get json context. ", errorFromCm.Error())
						manout.Error("cmd:", exec, "  ", args, " ", cmd)
					}

				} else {
					manout.Error("invalid usage", parseVarsMark, " needs 2 arguments at least. <varibale-name> <bash-command>")
				}

			case endMark:

				if inIfState {
					GetLogger().Debug("IF: DONE")
					inIfState = false
					ifState = true
				}
				if inIteration {
					GetLogger().Debug("ITERATION: DONE")
					inIteration = false
					abortFound := false
					returnCode := systools.ExitOk

					iterationCollect.ForEach(func(key gjson.Result, value gjson.Result) bool {
						var parsedExecLines []string
						for _, iLine := range iterationLines {
							iLine = strings.Replace(iLine, codeLinePH, value.String(), 1)
							iLine = strings.Replace(iLine, codeKeyPH, key.String(), 1)
							parsedExecLines = append(parsedExecLines, iLine)
						}
						GetLogger().WithFields(logrus.Fields{
							"key":       key,
							"value":     value,
							"subscript": parsedExecLines,
						}).Debug("... delegate script")
						abort, rCode, subs := TryParse(parsedExecLines, regularScript)
						returnCode = rCode
						parsedScript = append(parsedScript, subs...)

						if abort {
							abortFound = true
							return false
						}
						return true
					})

					if abortFound {
						return true, returnCode, parsedScript
					}
				}

			case iterateMark:
				if len(parts) == 3 {
					impMap, found := GetJSONPathResult(parts[1], parts[2])
					if !found {
						manout.Error("undefined data from path", parts[1], parts[2])
					} else {
						inIteration = true
						iterationCollect = impMap
						GetLogger().WithField("data", impMap).Debug("ITERATION: START")
					}
				} else {
					manout.Error("invalid arguments", "#@iterate needs <name-of-import> <path-to-data>")
				}
			default:
				GetLogger().WithField("unknown", parts[0]).Error("there is no command exists")
				manout.Error("ERROR depending inline macros annotated with "+startMark, " there is no macro defined named ", parts[0])
			}
		} else {
			parsedScript = append(parsedScript, line)
			// execute the *real* script lines
			if ifState {
				abort, returnCode := regularScript(line)
				if abort {
					return true, returnCode, parsedScript
				}
			} else {
				GetLogger().WithField("code", line).Debug("ignored because of if state")
			}
		}
	}
	GetLogger().WithFields(logrus.Fields{
		"parsed": parsedScript,
	}).Debug("... parsed result")
	return false, systools.ExitOk, parsedScript
}

func GetArgQuotedEntries(oristr string) ([]string, bool) {
	var result []string
	found := false
	re := regexp.MustCompile(`'[^']+'`)
	newStrs := re.FindAllString(oristr, -1)
	for _, s := range newStrs {
		found = true
		result = append(result, s)

	}
	return result, found
}

func SplitQuoted(oristr string, sep string) []string {
	var result []string
	var placeHolder map[string]string = make(map[string]string)

	found := false
	re := regexp.MustCompile(`'[^']+'`)
	newStrs := re.FindAllString(oristr, -1)
	i := 0
	for _, s := range newStrs {
		pl := "[$" + strconv.Itoa(i) + "]"
		placeHolder[pl] = strings.ReplaceAll(s, `'`, "")
		oristr = strings.Replace(oristr, s, pl, 1)
		found = true
		i++
	}
	result = strings.Split(oristr, sep)
	if !found {
		return result
	}

	for index, val := range result {
		if orgStr, fnd := placeHolder[val]; fnd {
			result[index] = orgStr
		}
	}

	return result
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

func ExportVarToFile(variable string, filename string) error {
	strData := GetPH(variable)
	if strData == "" {
		return errors.New("variable " + variable + " can not be used for export to file. not exists or empty")
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	var scopeVars map[string]string // empty but required
	if _, err2 := f.WriteString(handlePlaceHolder(strData, scopeVars)); err2 != nil {
		return err2
	}

	return nil
}

// ImportYAMLFile imports a yaml file as used for json map
func ImportYAMLFile(filename string) (map[string]interface{}, error) {
	if data, err := parseFileAsTemplateToByte(filename); err == nil {
		jsond, jerr := yaml.YAMLToJSON(data)
		if jerr != nil {
			return nil, jerr
		}
		m := make(map[string]interface{})
		if err := json.Unmarshal([]byte(jsond), &m); err != nil {
			return nil, err
		}
		if GetLogger().IsLevelEnabled(logrus.TraceLevel) {
			traceMap(m, filename)
		}
		return m, nil
	} else {
		return nil, err
	}
}

// ImportJSONFile imports a json file for reading
func ImportJSONFile(fileName string) (map[string]interface{}, error) {

	if data, err := parseFileAsTemplateToByte(fileName); err == nil {
		m := make(map[string]interface{})
		err = json.Unmarshal([]byte(data), &m)
		if err != nil {
			return testAndConvertJsonType(data)
		}
		if GetLogger().IsLevelEnabled(logrus.TraceLevel) {
			traceMap(m, fileName)
		}
		return m, nil
	} else {
		return nil, err
	}

}

// testAndConvertJsonType try to read a json string that might be an []interface{}
// if this succeeds then we convert it to an map[string]interface{}
// or return the UNmarschal error if this is failing too
func testAndConvertJsonType(data []byte) (map[string]interface{}, error) {
	var m []interface{}
	convert := make(map[string]interface{})
	if err := json.Unmarshal([]byte(data), &m); err == nil {
		for key, val := range m {
			keyStr := fmt.Sprintf("%d", key)
			switch val.(type) {
			case string, interface{}:
				convert[keyStr] = val
			default:
				return nil, errors.New("unsupported json structure")

			}

		}
		return convert, err
	} else {
		return convert, err
	}
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
	if err := mergo.Merge(&maporigin, mapin, mergo.WithOverride); err != nil {
		manout.Error("FATAL", "error while trying merge map")
		systools.Exit(systools.ErrorTemplate)
	}
	return maporigin
}

// ImportFolders import a list of folders recusiv
func ImportFolders(templatePath string, paths ...string) (string, error) {
	mapOrigin := GetOriginMap()

	template, terr := ImportFileContent(templatePath)
	if terr != nil {
		return "", terr
	}

	for _, path := range paths {
		path = HandlePlaceHolder(path)
		GetLogger().WithField("folder", path).Debug("process path")
		pathMap, parseErr := ImportFolder(path, templatePath)
		if parseErr != nil {
			return "", parseErr
		}
		mapOrigin = MergeVariableMap(pathMap, mapOrigin)
		UpdateOriginMap(mapOrigin)
	}
	result, herr := HandleJSONMap(template, mapOrigin)
	if herr != nil {
		return "", herr
	}
	template = result

	return template, nil
}

func parseFileAsTemplateToByte(path string) ([]byte, error) {
	var data []byte
	if parsedCnt, err := ParseFileAsTemplate(path); err != nil {
		return nil, err
	} else {
		data = []byte(parsedCnt)
		return data, nil
	}

}

func ParseFileAsTemplate(path string) (string, error) {
	path = HandlePlaceHolder(path)               // take care about placeholders
	mapOrigin := GetOriginMap()                  // load the current maps
	fileContent, terr := ImportFileContent(path) // load file content as string
	if terr != nil {
		return "", terr
	}

	// parsing as template
	if result, herr := HandleJSONMap(fileContent, mapOrigin); herr != nil {
		return "", herr
	} else {
		return result, nil
	}
}

func GetOriginMap() map[string]interface{} {
	exists, storedData := GetData("CTX_VAR_MAP")
	if exists {
		GetLogger().WithField("DATA", storedData).Trace("returning existing Variables map")
		return storedData
	}
	mapOrigin := make(map[string]interface{})
	GetLogger().Trace("returning NEW Variables map")
	return mapOrigin
}

// UpdateOriginMap updates the main templating map with an new one
func UpdateOriginMap(mapData map[string]interface{}) {
	GetLogger().WithField("DATA", mapData).Trace("update variables map")
	AddData("CTX_VAR_MAP", mapData)
}

// copies the placeholder to the origin map
// so there can be used in templates to
// this should be done after initilize the application.
// but not while runtime
func CopyPlaceHolder2Origin() {
	origin := GetOriginMap()
	GetPlaceHoldersFnc(func(phKey, phValue string) {
		origin[phKey] = phValue
	})
	UpdateOriginMap(origin)
}

// ImportFolder reads folder recursiv and reads all .json, .yml and .yaml files
func ImportFolder(path string, _ string) (map[string]interface{}, error) {

	//var mapOrigin map[string]interface{}
	//mapOrigin = make(map[string]interface{})
	mapOrigin := GetOriginMap()

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
			case ".yaml", ".yml":
				GetLogger().WithField("file", path).Debug("parsing included file (YAML)")
				jsonMap, loaderr = ImportYAMLFile(path)
				hit = true
			}
			if loaderr != nil {
				return loaderr
			}
			if hit {
				GetLogger().WithFields(logrus.Fields{
					"origin":   mapOrigin,
					"imported": jsonMap,
				}).Trace("merged Variable map")
				mapOrigin = MergeVariableMap(jsonMap, mapOrigin)
				GetLogger().WithField("result", mapOrigin).Trace("result of merge")
			}
		}

		return nil
	})
	return mapOrigin, err
}

// ImportFileContent imports a file and returns content as string
func ImportFileContent(filename string) (string, error) {
	GetLogger().WithField("file", filename).Debug("import file content")
	data, err := os.ReadFile(filename)
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
	eErr := tt.Execute(out, &m)
	if eErr != nil {
		return "", eErr
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
