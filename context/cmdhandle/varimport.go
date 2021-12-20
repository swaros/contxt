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

	"github.com/imdario/mergo"
	"github.com/tidwall/gjson"

	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/output"

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
	equalsMark      = "#@if-equals"
	osCheck         = "#@if-os"
	codeLinePH      = "__LINE__"
	codeKeyPH       = "__KEY__"
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
			parts := strings.Split(line, inlineCmdSep)
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
						output.Error("invalid usage", equalsMark, "need: str1 str2 ")
					}
				} else {
					output.Error("invalid usage", equalsMark, " can not be used in another if")
				}

			case equalsMark:
				if !inIfState {
					if len(parts) == 3 {
						leftEq := parts[1]
						rightEq := parts[2]
						inIfState = true
						ifState = leftEq == rightEq
					} else {
						output.Error("invalid usage", equalsMark, "need: str1 str2 ")
					}
				} else {
					output.Error("invalid usage", equalsMark, " can not be used in another if")
				}

			case inlineMark:
				if inIteration {
					iterationLines = append(iterationLines, strings.Replace(line, inlineMark+" ", "", 4))
					GetLogger().WithField("code", iterationLines).Debug("append to subscript")
				} else {
					output.Error("invalid usage", inlineMark, " only valid while in iteration")
				}

			case fromJSONMark:
				if len(parts) == 3 {
					err := AddJSON(parts[1], parts[2])
					if err != nil {
						output.Error("import from json string failed", parts[2], err)
					}
				} else {
					output.Error("invalid usage", fromJSONMark, " needs 2 arguments. <keyname> <json-source>")
				}

			case fromJSONCmdMark:
				if len(parts) >= 3 {
					returnValue := ""
					restSlice := parts[2:]
					keyname := parts[1]
					cmd := strings.Join(restSlice, " ")
					GetLogger().WithFields(logrus.Fields{"key": keyname, "cmd": restSlice}).Info("execute for import-json-exec")
					//GetLogger().WithField("slice", restSlice).Info("execute for import-json-exec")
					execCode, realExitCode, execErr := ExecuteScriptLine("bash", []string{"-c"}, cmd, func(output string) bool {
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
							output.Error("import from json string failed", err, ' ', returnValue)
						}
					}
				} else {
					output.Error("invalid usage", fromJSONCmdMark, " needs 2 arguments at least. <keyname> <bash-command>")
				}

			case parseVarsMark:
				if len(parts) >= 2 {
					var returnValues []string
					restSlice := parts[2:]
					cmd := strings.Join(restSlice, " ")
					internalCode, cmdCode, errorFromCm := ExecuteScriptLine("bash", []string{"-c"}, cmd, func(output string) bool {
						returnValues = append(returnValues, output)
						return true
					}, func(proc *os.Process) {
						GetLogger().WithField(parseVarsMark, proc).Trace("sub process")
					})

					if internalCode == ExitOk && errorFromCm == nil && cmdCode == 0 {
						GetLogger().WithField("values", returnValues).Trace("got values")
						SetPH(parts[1], HandlePlaceHolder(strings.Join(returnValues, "\n")))
					} else {
						GetLogger().WithFields(logrus.Fields{
							"returnCode": cmdCode,
							"error":      errorFromCm.Error,
						}).Error("subcommand failed.")
						output.Error("Subcommand failed", cmd, " ... was used to get json context.")
					}

				} else {
					output.Error("invalid usage", parseVarsMark, " needs 2 arguments at least. <varibale-name> <bash-command>")
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
					returnCode := ExitOk

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
						output.Error("undefined data from path", parts[1], parts[2])
					} else {
						inIteration = true
						iterationCollect = impMap
						GetLogger().WithField("data", impMap).Debug("ITERATION: START")
					}
				} else {
					output.Error("invalid arguments", "#@iterate needs <name-of-import> <path-to-data>")
				}
			default:
				GetLogger().WithField("unknown", parts[0]).Error("there is no command exists")
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
	return false, ExitOk, parsedScript
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
	if err := mergo.Merge(&maporigin, mapin, mergo.WithOverride); err != nil {
		output.Error("FATAL", "error while trying merge map")
		os.Exit(10)
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

func UpdateOriginMap(mapData map[string]interface{}) {
	GetLogger().WithField("DATA", mapData).Trace("update variables map")
	AddData("CTX_VAR_MAP", mapData)
}

// ImportFolder reads folder recursiv and reads all .json, .yml and .yaml files
func ImportFolder(path string, templatePath string) (map[string]interface{}, error) {

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
