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

package tasks

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cp "github.com/otiai10/copy"
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
			"exit the script with errror",
		},
		{"ifos",
			func(s string) bool {
				return configure.GetOs() == s
			},
			RISK_LEVEL_LOW,
			"check if the current os is the same as the given string. e.g. if ifos('linux') { ... }",
		},
		{"getos",
			func() string {
				return configure.GetOs()
			},
			RISK_LEVEL_LOW,
			"get the current os. e.g. println('you are working on: ',getos())",
		},
		{"importJson",
			func(key, json string) error {
				json = t.phHandler.HandlePlaceHolder(json)
				err := t.dataHandler.AddJSON(key, json)
				if err != nil {
					anko.ThrowException(err, fmt.Sprintf("importJson('%s','%s')", key, json))
					t.out(MsgError(MsgError{Err: errors.New("error while parsing json: " + err.Error()), Reference: "importJson(key,json)", Target: t.target}))
				}
				return err
			},
			RISK_LEVEL_LOW,
			"import a json string into the data store. e.g. importJson('key','{\"key\":\"value\"}')",
		},
		{"varAsJson",
			func(key string) string {
				data, _ := t.dataHandler.GetDataAsJson(key)
				return data
			},
			RISK_LEVEL_LOW,
			"get the data as json string. e.g. data = varAsJson('key')",
		},
		{"varAsYaml",
			func(key string) string {
				data, _ := t.dataHandler.GetDataAsYaml(key)
				return data
			},
			RISK_LEVEL_LOW,
			"get the data as yaml string. e.g. data = varAsYaml('key')",
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
			},
			RISK_LEVEL_HIGH,
			`execute a command and return the output. e.g. path,exitCode,err = exec('ls -l')
it returns the output of the command, the exit code and an error if any
this can be also accessed as map:
result = exec('ls -l')
path = result[0]
exitCode = result[1]
err = result[2]`,
		},
		{"varSet",
			func(key, value string) {
				value = t.phHandler.HandlePlaceHolder(value)
				t.phHandler.SetPH(key, value)
			},
			RISK_LEVEL_LOW,
			`set a variable. e.g. varSet('key','value')`,
		},
		{"varAppend",
			func(key, value string) {
				value = t.phHandler.HandlePlaceHolder(value)
				t.phHandler.AppendToPH(key, value)
			},
			RISK_LEVEL_LOW,
			`append a value to a variable. e.g. varAppend('key','value')`,
		},
		{"varGet",
			func(key string) string {
				return t.phHandler.GetPH(key)
			},
			RISK_LEVEL_LOW,
			`get a variable. e.g. value = varGet('key')`,
		},
		{"varParse",
			func(line string) string {
				return t.phHandler.HandlePlaceHolder(line)
			},
			RISK_LEVEL_LOW,
			`parses the given line with existing values = varParse('hello ${key}')`,
		},
		{"varExists",
			func(key string) bool {
				_, ok := t.phHandler.GetPHExists(key)
				return ok
			},
			RISK_LEVEL_LOW,
			`check if a variable exists. e.g. if varExists('key') { ... }`,
		},
		{"varMapSet",
			func(key, path, value string) error {
				value = t.phHandler.HandlePlaceHolder(value)
				err := t.dataHandler.SetJSONValueByPath(key, path, value)
				if err != nil {
					anko.ThrowException(err, fmt.Sprintf("varMapSet('%s','%s','%s')", key, path, value))
					t.out(MsgError(MsgError{Err: err, Reference: "varMapSet(key,path,value)", Target: t.target}))
				}
				return err
			},
			RISK_LEVEL_LOW,
			`set a value in a map. e.g. varMapSet('key','path','value')`,
		},
		{"varMapToJson",
			func(mapKey, varKey string) string {
				if data, ok := t.dataHandler.GetDataAsJson(mapKey); ok {
					return data
				}
				return ""
			},
			RISK_LEVEL_LOW,
			`get a map as json string. e.g. data = varMapToJson('key','path')`,
		},
		{"varMapToYaml",
			func(mapKey, varKey string) string {
				if data, ok := t.dataHandler.GetDataAsYaml(mapKey); ok {
					return data
				}
				return ""
			},
			RISK_LEVEL_LOW,
			`get a map as yaml string. e.g. data = varMapToYaml('key','path')`,
		},
		{"varWrite",
			func(varName, fileName string) error {
				return t.phHandler.ExportVarToFile(varName, fileName)
			},
			RISK_LEVEL_HIGH,
			`write a variable to a file. e.g. varWrite('key','output.txt')`,
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
			`write a content to a file. e.g. writeFile('output.txt','hello world')`,
		},
		{"ReadFile",
			func(fileName string) (string, error) {
				f, err := os.Open(fileName)
				if err != nil {
					anko.ThrowException(err, fmt.Sprintf("ReadFile('%s')", fileName))
					return "", err
				}
				defer f.Close()
				fi, err := f.Stat()
				if err != nil {
					anko.ThrowException(err, fmt.Sprintf("ReadFile('%s')", fileName))
					return "", err
				}
				data := make([]byte, fi.Size())
				if _, err := f.Read(data); err != nil {
					anko.ThrowException(err, fmt.Sprintf("ReadFile('%s')", fileName))
					return "", err
				}
				return string(data), nil
			},
			RISK_LEVEL_HIGH,
			`read a file. e.g. content,err = ReadFile('input.txt')`,
		},
		{"copy",
			func(src, dst string, include ...string) error {
				opt := cp.Options{
					Skip: func(info os.FileInfo, src, dest string) (bool, error) {
						if len(include) == 0 {
							return false, nil
						}
						for _, s := range include {
							if strings.HasSuffix(src, s) {
								return true, nil
							}
						}
						return false, nil
					},
				}
				return cp.Copy(src, dst, opt)
			},
			RISK_LEVEL_HIGH,
			`copy a file. e.g. copy('input.txt','output.txt') or copy a directory copy('input','output')
you can also limit the copy by suffix. e.g. copy('input','output','.yml','.json')
this will only copy files with the given suffixes`,
		},
		{"copyButSkip",
			func(src, dst string, skip ...string) error {
				opt := cp.Options{
					Skip: func(info os.FileInfo, src, dest string) (bool, error) {
						if len(skip) == 0 {
							return false, nil
						}
						for _, s := range skip {
							if strings.HasSuffix(src, s) {
								return true, nil
							}
						}
						return false, nil
					},
				}
				return cp.Copy(src, dst, opt)
			},
			RISK_LEVEL_HIGH,
			`copy a file but skip some files depending suffix. e.g. copyButSkip('input','output','.git','.tmp')
this will copy any files except the ones with the given suffixes`,
		},
		{"remove",
			func(path string) error {
				path = t.phHandler.HandlePlaceHolder(path)
				path = filepath.FromSlash(path)
				return os.RemoveAll(path)
			},
			RISK_LEVEL_HIGH,
			`remove a file or directory. e.g. remove('output.txt') or remove('output')`,
		},
		{"mkdir",
			func(path string) error {
				path = t.phHandler.HandlePlaceHolder(path)
				path = filepath.FromSlash(path)
				return os.MkdirAll(path, os.ModePerm)
			},
			RISK_LEVEL_HIGH,
			`create a directory. e.g. mkdir('output'). it creates all directories in the path`,
		},
		{"base64Encode",
			func(data string) string {
				str := base64.StdEncoding.EncodeToString([]byte(data))
				return str

			},
			RISK_LEVEL_LOW,
			`encode a string to base64. e.g. base64Encode('hello world')`,
		},
		{"base64Decode",
			func(data string) (string, error) {
				str, err := base64.StdEncoding.DecodeString(data)
				if err != nil {
					anko.ThrowException(err, fmt.Sprintf("base64Decode('%s')", data))
					return "", err
				}
				return string(str), nil
			},
			RISK_LEVEL_LOW,
			`decode a base64 string. e.g. base64Decode('aGVsbG8gd29ybGQ=')`,
		},
	}
	return cmdList
}

func (t *targetExecuter) SetFunctions(anko *AnkoRunner) {
	if err := anko.AddDefaultDefines(t.GetFnAsDefaults(anko)); err != nil {
		t.out(MsgError(MsgError{Err: err, Reference: "SetFunctions()", Target: t.target}))
	}
}
