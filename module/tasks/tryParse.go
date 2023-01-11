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
	"errors"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/systools"
	"github.com/tidwall/gjson"
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

func (t *targetExecuter) TryParse(script []string, regularScript func(string) (bool, int)) (bool, int, []string) {
	// first check if any required handler is set
	if t.dataHandler == nil {
		panic("dataHandler is not set")
	}

	if t.phHandler == nil {
		panic("placeholderHandler is not set")
	}

	inIteration := false
	inIfState := false
	ifState := true
	var iterationLines []string
	var parsedScript []string
	var iterationCollect gjson.Result
	for _, line := range script {
		if t.phHandler != nil {
			line = t.phHandler.HandlePlaceHolder(line)
		}
		if len(line) > len(startMark) && line[0:len(startMark)] == startMark {
			//parts := strings.Split(line, inlineCmdSep)
			parts := systools.SplitQuoted(line, inlineCmdSep)

			t.getLogger().WithField("keywords", parts).Debug("try to parse parts")
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
						t.out(MsgError(errors.New("invalid usage " + equalsMark + " need: str1 str2")))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					t.out(MsgError(errors.New("invalid usage " + equalsMark + " can not be used in another if")))
					return true, systools.ErrorCheatMacros, parsedScript
				}

			case equalsMark:
				if !inIfState {
					if len(parts) == 3 {
						leftEq := parts[1]
						rightEq := parts[2]
						inIfState = true
						ifState = leftEq == rightEq
						t.getLogger().WithFields(logrus.Fields{"condition": ifState, "left": leftEq, "right": rightEq}).Debug(equalsMark)
					} else {
						t.out(MsgError(errors.New("invalid usage " + equalsMark + " need: str1 str2")))
						return true, systools.ErrorCheatMacros, parsedScript

					}
				} else {
					t.out(MsgError(errors.New("invalid usage " + equalsMark + " can not be used in another if")))
					return true, systools.ErrorCheatMacros, parsedScript
				}

			case notEqualsMark:
				if !inIfState {
					if len(parts) == 3 {
						leftEq := parts[1]
						rightEq := parts[2]
						inIfState = true
						ifState = leftEq != rightEq
						t.getLogger().WithFields(logrus.Fields{"condition": ifState, "left": leftEq, "right": rightEq}).Debug(notEqualsMark)
					} else {
						t.out(MsgError(errors.New("invalid usage " + notEqualsMark + " need: str1 str2")))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					t.out(MsgError(errors.New("invalid usage " + notEqualsMark + " can not be used in another if")))
					return true, systools.ErrorCheatMacros, parsedScript
				}

			case inlineMark:
				if inIteration {
					iterationLines = append(iterationLines, strings.Replace(line, inlineMark+" ", "", 4))
					t.getLogger().WithField("code", iterationLines).Debug("append to subscript")
				} else {
					t.out(MsgError(errors.New("invalid usage " + inlineMark + " only valid while in iteration")))
					return true, systools.ErrorCheatMacros, parsedScript
				}

			case fromJSONMark:

				if len(parts) >= 3 {
					err := t.dataHandler.AddJSON(parts[1], strings.Join(parts[2:], ""))
					if err != nil {
						t.out(MsgError(errors.New("error while parsing json: " + err.Error())))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					t.out(MsgError(errors.New("invalid usage " + fromJSONMark + " needs 2 arguments. <keyname> <json-source-string>")))
					return true, systools.ErrorCheatMacros, parsedScript
				}

			case fromJSONCmdMark:
				if len(parts) >= 3 {
					returnValue := ""
					restSlice := parts[2:]
					keyname := parts[1]
					cmd := strings.Join(restSlice, " ")
					t.getLogger().WithFields(logrus.Fields{"key": keyname, "cmd": restSlice}).Info("execute for import-json-exec")
					execCode, realExitCode, execErr := t.ExecuteScriptLine(cmd, func(output string, e error) bool {
						returnValue = returnValue + output
						t.getLogger().WithField("cmd-output", output).Info("result of command")
						return true
					}, func(proc *os.Process) {
						t.getLogger().WithField("import-json-proc", proc).Trace("import-json-process")
					})

					if execErr != nil {
						t.getLogger().WithFields(logrus.Fields{
							"intern":       execCode,
							"process-exit": realExitCode,
							"key":          keyname,
							"cmd":          restSlice}).Error("execute for import-json-exec failed")
					} else {

						err := t.dataHandler.AddJSON(keyname, returnValue)
						if err != nil {
							t.getLogger().WithField("error-on-parsing-string", returnValue).Debug("result of command")
							t.out(MsgError(errors.New("error while parsing json: " + err.Error())))
							return true, systools.ErrorCheatMacros, parsedScript
						}
					}
				} else {
					t.out(MsgError(errors.New("invalid usage " + fromJSONCmdMark + " needs 2 arguments at least. <keyname> <bash-command>")))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			case addvarMark:
				if len(parts) >= 2 {
					setKeyname := parts[1]
					setValue := strings.Join(parts[2:], " ")
					if ok := t.phHandler.AppendToPH(setKeyname, t.phHandler.HandlePlaceHolder(setValue)); !ok {
						t.out(MsgError(errors.New("variable must exists for add " + addvarMark + " " + setKeyname)))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					t.out(MsgError(errors.New("invalid usage " + addvarMark + " needs 2 arguments at least. <keyname> <value>")))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			case setvarMark:
				if len(parts) >= 2 {
					setKeyname := parts[1]
					setValue := strings.Join(parts[2:], " ")
					t.phHandler.SetPH(setKeyname, t.phHandler.HandlePlaceHolder(setValue))
				} else {
					t.out(MsgError(errors.New("invalid usage " + setvarMark + " needs 2 arguments at least. <keyname> <value>")))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			case setvarInMap:
				if len(parts) >= 3 {
					mapName := parts[1]
					path := parts[2]
					setValue := strings.Join(parts[3:], " ")
					if err := t.dataHandler.SetJSONValueByPath(mapName, path, setValue); err != nil {
						t.out(MsgError(err))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					t.out(MsgError(errors.New("invalid usage " + setvarInMap + " needs 3 arguments at least. <mapName> <json.path> <value>")))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			case writeVarToFile:
				if len(parts) == 3 {
					varName := parts[1]
					fileName := parts[2]
					t.phHandler.ExportVarToFile(varName, fileName)
				} else {
					t.out(MsgError(errors.New("invalid usage " + writeVarToFile + " needs 2 arguments at least. <variable> <filename>")))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			case exportToJson:
				if len(parts) == 3 {
					mapKey := parts[1]
					varName := parts[2]
					if newStr, exists := t.dataHandler.GetDataAsJson(mapKey); exists {
						t.phHandler.SetPH(varName, t.phHandler.HandlePlaceHolder(newStr))
					} else {
						t.out(MsgError(errors.New("map with key " + mapKey + " not exists")))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					t.out(MsgError(errors.New("invalid usage " + exportToJson + " needs 2 arguments at least. <map-key> <variable>")))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			case exportToYaml:
				if len(parts) == 3 {
					mapKey := parts[1]
					varName := parts[2]
					if newStr, exists := t.dataHandler.GetDataAsYaml(mapKey); exists {
						t.phHandler.SetPH(varName, t.phHandler.HandlePlaceHolder(newStr))
					} else {
						t.out(MsgError(errors.New("map with key " + mapKey + " not exists")))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					t.out(MsgError(errors.New("invalid usage " + exportToYaml + " needs 2 arguments at least. <map-key> <variable>")))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			case parseVarsMark:
				if len(parts) >= 2 {
					var returnValues []string
					restSlice := parts[2:]
					cmd := strings.Join(restSlice, " ")

					internalCode, cmdCode, errorFromCm := t.ExecuteScriptLine(cmd, func(output string, e error) bool {
						if e == nil {
							returnValues = append(returnValues, output)
						}
						return true

					}, func(proc *os.Process) {
						t.getLogger().WithField(parseVarsMark, proc).Trace("sub process")
					})

					if internalCode == systools.ExitOk && errorFromCm == nil && cmdCode == 0 {
						t.getLogger().WithField("values", returnValues).Trace("got values")
						t.phHandler.SetPH(parts[1], t.phHandler.HandlePlaceHolder(strings.Join(returnValues, "\n")))
					} else {
						t.getLogger().WithFields(logrus.Fields{
							"returnCode": cmdCode,
							"error":      errorFromCm.Error,
						}).Error("subcommand failed.")
						t.out(MsgError(errorFromCm))
						return true, systools.ErrorCheatMacros, parsedScript
					}

				} else {
					t.out(MsgError(errors.New("invalid usage " + parseVarsMark + " needs 2 arguments at least. <varibale-name> <bash-command>")))
					return true, systools.ErrorCheatMacros, parsedScript
				}

			case endMark:

				if inIfState {
					t.getLogger().Debug("IF: DONE")
					inIfState = false
					ifState = true
				}
				if inIteration {
					t.getLogger().Debug("ITERATION: DONE")
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
						t.getLogger().WithFields(logrus.Fields{
							"key":       key,
							"value":     value,
							"subscript": parsedExecLines,
						}).Debug("... delegate script")
						abort, rCode, subs := t.TryParse(parsedExecLines, regularScript)
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
					impMap, found := t.dataHandler.GetJSONPathResult(parts[1], parts[2])
					if !found {
						t.out(MsgError(errors.New("undefined data from path " + parts[1] + " " + parts[2])))
					} else {
						inIteration = true
						iterationCollect = impMap
						t.getLogger().WithField("data", impMap).Debug("ITERATION: START")
					}
				} else {
					t.out(MsgError(errors.New("invalid arguments #@iterate needs <name-of-import> <path-to-data>")))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			default:
				t.getLogger().WithField("unknown", parts[0]).Error("there is no command exists")
				t.out(MsgError(errors.New("unknown macro " + parts[0])))
				return true, systools.ErrorCheatMacros, parsedScript
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
				t.getLogger().WithField("code", line).Debug("ignored because of if state")
			}
		}
	}
	t.getLogger().WithFields(logrus.Fields{
		"parsed": parsedScript,
	}).Debug("... parsed result")
	return false, systools.ExitOk, parsedScript
}
