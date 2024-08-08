package tasks

import (
	"errors"
	"fmt"
	"os"

	"github.com/swaros/contxt/module/configure"
)

func (t *targetExecuter) GetFnAsDefaults(anko *AnkoRunner) []AnkoDefiner {
	// define the default functions
	cmdList := []AnkoDefiner{
		{"exit",
			func() {
				anko.cancelationFn()
			},
			RISK_LEVEL_LOW,
		},
		{"ifos",
			func(s string) bool {
				return configure.GetOs() == s
			}, RISK_LEVEL_LOW,
		},
		{"getos",
			func() string {
				return configure.GetOs()
			},
			RISK_LEVEL_LOW,
		},
		{"importJson",
			func(key, json string) error {
				err := t.dataHandler.AddJSON(key, json)
				if err != nil {
					anko.ThrowException(err, fmt.Sprintf("importJson('%s','%s')", key, json))
					t.out(MsgError(MsgError{Err: errors.New("error while parsing json: " + err.Error()), Reference: "importJson(key,json)", Target: t.target}))
				}
				return err
			}, RISK_LEVEL_LOW,
		},
		{"varAsJson",
			func(key string) string {
				data, _ := t.dataHandler.GetDataAsJson(key)
				return data
			}, RISK_LEVEL_LOW,
		},
		{"varAsYaml",
			func(key string) string {
				data, _ := t.dataHandler.GetDataAsYaml(key)
				return data
			}, RISK_LEVEL_LOW,
		},
		{"exec",
			func(cmd string) (string, int, error) {
				returnValue := ""
				add := ""
				runCmd, runArgs := t.commandFallback.GetMainCmd(configure.Options{})
				_, cmdExit, execErr := t.ExecuteScriptLine(runCmd, runArgs, cmd, func(output string, e error) bool {
					returnValue = returnValue + add + output
					add = "\n"
					if e != nil {
						anko.ThrowException(e, fmt.Sprintf("exec('%s')", cmd))
						return false
					}
					return true
				}, func(proc *os.Process) {
					// nothing to do here
				})
				if execErr != nil {
					anko.ThrowException(execErr, fmt.Sprintf("exec('%s')", cmd))
					t.out(MsgError(MsgError{Err: execErr, Reference: "exec(cmd)", Target: t.target}))
					return returnValue, cmdExit, execErr
				} else {
					return returnValue, cmdExit, nil
				}
			}, RISK_LEVEL_HIGH,
		},
		{"varSet",
			func(key, value string) {
				value = t.phHandler.HandlePlaceHolder(value)
				t.phHandler.SetPH(key, value)
			},
			RISK_LEVEL_LOW,
		},
		{"varAppend",
			func(key, value string) {
				value = t.phHandler.HandlePlaceHolder(value)
				t.phHandler.AppendToPH(key, value)
			},
			RISK_LEVEL_LOW,
		},
		{"varGet",
			func(key string) string {
				return t.phHandler.GetPH(key)
			},
			RISK_LEVEL_LOW,
		},
		{"varMapSet",
			func(key, path, value string) {
				value = t.phHandler.HandlePlaceHolder(value)
				t.dataHandler.SetJSONValueByPath(key, path, value)
			},
			RISK_LEVEL_LOW,
		},
		{"varMapToJson",
			func(mapKey, varKey string) string {
				if data, ok := t.dataHandler.GetDataAsJson(mapKey); ok {
					return data
				}
				return ""
			},
			RISK_LEVEL_LOW,
		},
		{"varMapToYaml",
			func(mapKey, varKey string) string {
				if data, ok := t.dataHandler.GetDataAsYaml(mapKey); ok {
					return data
				}
				return ""
			},
			RISK_LEVEL_LOW,
		},
		{"varWrite",
			func(varName, fileName string) error {
				return t.phHandler.ExportVarToFile(varName, fileName)
			},
			RISK_LEVEL_HIGH,
		},
		{"writeFile",
			func(fileName, content string) error {
				f, err := os.Create(fileName)
				if err != nil {
					anko.ThrowException(err, fmt.Sprintf("writeFile.create('%s','%s')", fileName, content))
					return err
				}
				defer f.Close()
				if _, err2 := f.WriteString(t.phHandler.HandlePlaceHolder(content)); err2 != nil {
					anko.ThrowException(err2, fmt.Sprintf("writeFile.write('%s','%s')", fileName, content))
					return err2
				}
				return nil
			},
			RISK_LEVEL_HIGH,
		},
	}
	return cmdList
}

func (t *targetExecuter) SetFunctions(anko *AnkoRunner) {
	anko.AddDefaultDefines(t.GetFnAsDefaults(anko))
}
