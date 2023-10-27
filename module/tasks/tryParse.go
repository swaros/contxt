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

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/mimiclog"
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
	t.getLogger().Debug("TPARSE: entered")
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

			t.getLogger().Debug("TPARSE: try to parse parts", parts)
			if len(parts) < 1 {
				continue
			}
			switch parts[0] {

			case osCheck:
				t.getLogger().Debug("TPARSE: check os")
				if !inIfState {
					if len(parts) == 2 {
						leftEq := parts[1]
						rightEq := configure.GetOs()
						inIfState = true
						ifState = leftEq == rightEq
					} else {
						t.out(MsgError(MsgError{Err: errors.New("invalid usage " + equalsMark + " need: str1 str2"), Reference: line, Target: t.target}))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					t.out(MsgError(MsgError{Err: errors.New("invalid usage " + equalsMark + " can not be used in another if"), Reference: line, Target: t.target}))
					return true, systools.ErrorCheatMacros, parsedScript
				}

			case equalsMark:
				t.getLogger().Debug("TPARSE: equals")
				if !inIfState {
					if len(parts) == 3 {
						leftEq := parts[1]
						rightEq := parts[2]
						inIfState = true
						ifState = leftEq == rightEq
						logFields := mimiclog.Fields{"condition": ifState, "left": leftEq, "right": rightEq}
						t.getLogger().Debug(equalsMark, logFields)
					} else {
						t.out(MsgError(MsgError{Err: errors.New("invalid usage " + equalsMark + " need: str1 str2"), Reference: line, Target: t.target}))
						return true, systools.ErrorCheatMacros, parsedScript

					}
				} else {
					t.out(MsgError(MsgError{Err: errors.New("invalid usage " + equalsMark + " can not be used in another if"), Reference: line, Target: t.target}))
					return true, systools.ErrorCheatMacros, parsedScript
				}

			case notEqualsMark:
				t.getLogger().Debug("TPARSE: not equals")
				if !inIfState {
					if len(parts) == 3 {
						leftEq := parts[1]
						rightEq := parts[2]
						inIfState = true
						ifState = leftEq != rightEq
						logFields := mimiclog.Fields{"condition": ifState, "left": leftEq, "right": rightEq}
						t.getLogger().Debug(notEqualsMark, logFields)
					} else {
						t.out(MsgError(MsgError{Err: errors.New("invalid usage " + notEqualsMark + " need: str1 str2"), Reference: line, Target: t.target}))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					t.out(MsgError(MsgError{Err: errors.New("invalid usage " + notEqualsMark + " can not be used in another if"), Reference: line, Target: t.target}))
					return true, systools.ErrorCheatMacros, parsedScript
				}

			case inlineMark:
				t.getLogger().Debug("TPARSE: inline")
				if inIteration {
					iterationLines = append(iterationLines, strings.Replace(line, inlineMark+" ", "", 4))
					logFields := mimiclog.Fields{"iterationLines": iterationLines}
					t.getLogger().Debug("TPARSE: append to subscript", logFields)
				} else {
					t.out(MsgError(MsgError{Err: errors.New("invalid usage " + inlineMark + " only valid while in iteration"), Reference: line, Target: t.target}))
					return true, systools.ErrorCheatMacros, parsedScript
				}

			case fromJSONMark:
				t.getLogger().Debug("TPARSE: from json")
				if len(parts) >= 3 {
					err := t.dataHandler.AddJSON(parts[1], strings.Join(parts[2:], ""))
					if err != nil {
						t.out(MsgError(MsgError{Err: errors.New("error while parsing json: " + err.Error()), Reference: line, Target: t.target}))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					t.out(MsgError(MsgError{Err: errors.New("invalid usage " + fromJSONMark + " needs 2 arguments. <keyname> <json-source-string>"), Reference: line, Target: t.target}))
					return true, systools.ErrorCheatMacros, parsedScript
				}

			case fromJSONCmdMark:
				t.getLogger().Debug("TPARSE: from json exec")
				if len(parts) >= 3 {
					returnValue := ""
					restSlice := parts[2:]
					keyname := parts[1]
					cmd := strings.Join(restSlice, " ")
					runCmd, runArgs := t.commandFallback.GetMainCmd(configure.Options{})
					logFields := mimiclog.Fields{"key": keyname, "cmd": restSlice}
					t.getLogger().Info("execute for import-json-exec", logFields)
					execCode, realExitCode, execErr := t.ExecuteScriptLine(runCmd, runArgs, cmd, func(output string, e error) bool {
						returnValue = returnValue + output
						logFields := mimiclog.Fields{"key": keyname, "cmd": restSlice, "output": output}
						t.getLogger().Info("result of command", logFields)
						return true
					}, func(proc *os.Process) {
						logFields := mimiclog.Fields{"key": keyname, "cmd": restSlice, "pid": proc.Pid}
						t.getLogger().Trace("import-json-process", logFields)
					})

					if execErr != nil {
						logFields := mimiclog.Fields{"key": keyname, "cmd": restSlice, "error": execErr, "exit-code": realExitCode, "intern": execCode}
						t.getLogger().Error("execute for import-json-exec failed", logFields)
						t.out(MsgError(MsgError{Err: errors.New("error while executing command: " + execErr.Error()), Reference: line, Target: t.target}))
						return true, systools.ErrorCheatMacros, parsedScript
					} else {

						err := t.dataHandler.AddJSON(keyname, returnValue)
						if err != nil {
							t.getLogger().Debug("TPARSE: result of command", returnValue)
							t.out(MsgError(MsgError{Err: errors.New("error while parsing json: " + err.Error()), Reference: line, Target: t.target}))
							return true, systools.ErrorCheatMacros, parsedScript
						}
					}
				} else {
					t.out(MsgError(MsgError{Err: errors.New("invalid usage " + fromJSONCmdMark + " needs 2 arguments at least. <keyname> <bash-command>"), Reference: line, Target: t.target}))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			case addvarMark:
				if len(parts) >= 2 {
					setKeyname := parts[1]
					setValue := strings.Join(parts[2:], " ")
					if ok := t.phHandler.AppendToPH(setKeyname, t.phHandler.HandlePlaceHolder(setValue)); !ok {
						t.out(MsgError(MsgError{Err: errors.New("variable must exists for add " + addvarMark + " " + setKeyname), Reference: line, Target: t.target}))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					t.out(MsgError(MsgError{Err: errors.New("invalid usage " + addvarMark + " needs 2 arguments at least. <keyname> <value>"), Reference: line, Target: t.target}))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			case setvarMark:
				if len(parts) >= 2 {
					setKeyname := parts[1]
					setValue := strings.Join(parts[2:], " ")
					t.phHandler.SetPH(setKeyname, t.phHandler.HandlePlaceHolder(setValue))
				} else {
					t.out(MsgError(MsgError{Err: errors.New("invalid usage " + setvarMark + " needs 2 arguments at least. <keyname> <value>"), Reference: line, Target: t.target}))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			case setvarInMap:
				if len(parts) >= 3 {
					mapName := parts[1]
					path := parts[2]
					setValue := strings.Join(parts[3:], " ")
					if err := t.dataHandler.SetJSONValueByPath(mapName, path, setValue); err != nil {
						t.out(MsgError(MsgError{Err: errors.New("error while setting value in map: " + err.Error()), Reference: line, Target: t.target}))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					t.out(MsgError(MsgError{Err: errors.New("invalid usage " + setvarInMap + " needs 3 arguments at least. <mapName> <json.path> <value>"), Reference: line, Target: t.target}))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			case writeVarToFile:
				if len(parts) == 3 {
					varName := parts[1]
					fileName := parts[2]
					if err := t.phHandler.ExportVarToFile(varName, fileName); err != nil {
						t.out(MsgError(MsgError{Err: errors.New("error while writing variable to file: " + err.Error()), Reference: line, Target: t.target}))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					t.out(MsgError(MsgError{Err: errors.New("invalid usage " + writeVarToFile + " needs 2 arguments at least. <variable> <filename>"), Reference: line, Target: t.target}))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			case exportToJson:
				if len(parts) == 3 {
					mapKey := parts[1]
					varName := parts[2]
					if newStr, exists := t.dataHandler.GetDataAsJson(mapKey); exists {
						t.phHandler.SetPH(varName, t.phHandler.HandlePlaceHolder(newStr))
					} else {
						t.out(MsgError(MsgError{Err: errors.New("map with key " + mapKey + " not exists"), Reference: line, Target: t.target}))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					//t.out(MsgError(errors.New("invalid usage " + exportToJson + " needs 2 arguments at least. <map-key> <variable>")))
					t.out(MsgError(MsgError{Err: errors.New("invalid usage " + exportToJson + " needs 2 arguments at least. <map-key> <variable>"), Reference: line, Target: t.target}))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			case exportToYaml:
				if len(parts) == 3 {
					mapKey := parts[1]
					varName := parts[2]
					if newStr, exists := t.dataHandler.GetDataAsYaml(mapKey); exists {
						t.phHandler.SetPH(varName, t.phHandler.HandlePlaceHolder(newStr))
					} else {
						//t.out(MsgError(errors.New("map with key " + mapKey + " not exists")))
						t.out(MsgError(MsgError{Err: errors.New("map with key " + mapKey + " not exists"), Reference: line, Target: t.target}))
						return true, systools.ErrorCheatMacros, parsedScript
					}
				} else {
					//t.out(MsgError(errors.New("invalid usage " + exportToYaml + " needs 2 arguments at least. <map-key> <variable>")))
					t.out(MsgError(MsgError{Err: errors.New("invalid usage " + exportToYaml + " needs 2 arguments at least. <map-key> <variable>"), Reference: line, Target: t.target}))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			case parseVarsMark:
				if len(parts) >= 2 {
					var returnValues []string
					restSlice := parts[2:]
					cmd := strings.Join(restSlice, " ")
					runCmd, runArgs := t.commandFallback.GetMainCmd(configure.Options{})
					internalCode, cmdCode, errorFromCm := t.ExecuteScriptLine(runCmd, runArgs, cmd, func(output string, e error) bool {
						if e == nil {
							returnValues = append(returnValues, output)
						}
						return true

					}, func(proc *os.Process) {
						logField := mimiclog.Fields{"pid": proc.Pid}
						t.getLogger().Trace("sub process", logField)
					})

					if internalCode == systools.ExitOk && errorFromCm == nil && cmdCode == 0 {
						t.getLogger().Trace("got values", returnValues)
						t.phHandler.SetPH(parts[1], t.phHandler.HandlePlaceHolder(strings.Join(returnValues, "\n")))
					} else {
						logFields := mimiclog.Fields{"returnCode": cmdCode, "error": errorFromCm}
						t.getLogger().Error("subcommand failed.", logFields)
						t.out(MsgError(MsgError{Err: errors.New("error while executing command: " + errorFromCm.Error()), Reference: line, Target: t.target}))
						return true, systools.ErrorCheatMacros, parsedScript
					}

				} else {
					t.out(MsgError{Err: errors.New("invalid usage " + parseVarsMark + " needs 2 arguments at least. <varibale-name> <bash-command>"), Reference: line, Target: t.target})
					return true, systools.ErrorCheatMacros, parsedScript
				}

			case endMark:

				if inIfState {
					t.getLogger().Debug("TPARSE: IF: DONE")
					inIfState = false
					ifState = true
				}
				if inIteration {
					t.getLogger().Debug("TPARSE: ITERATION: DONE")
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
						logFields := mimiclog.Fields{"key": key, "value": value, "subscript": parsedExecLines}
						t.getLogger().Debug("TPARSE: ... delegate script", logFields)
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
						t.out(MsgError(MsgError{Err: errors.New("undefined data from path " + parts[1] + " " + parts[2]), Reference: line, Target: t.target}))
					} else {
						inIteration = true
						iterationCollect = impMap
						t.getLogger().Debug("TPARSE: ITERATION: START", impMap)
					}
				} else {
					t.out(MsgError(MsgError{Err: errors.New("invalid arguments #@iterate needs <name-of-import> <path-to-data>"), Reference: line, Target: t.target}))
					return true, systools.ErrorCheatMacros, parsedScript
				}
			default:
				logFields := mimiclog.Fields{"code": line, "unknown": parts[0]}
				t.getLogger().Error("there is no command exists", logFields)
				t.out(MsgError(MsgError{Err: errors.New("unknown macro " + parts[0]), Reference: line, Target: t.target}))
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
				t.getLogger().Debug("TPARSE: ignored because of if state", line)
			}
		}
	}
	t.getLogger().Debug("TPARSE: ... parsed result", parsedScript)
	return false, systools.ExitOk, parsedScript
}
