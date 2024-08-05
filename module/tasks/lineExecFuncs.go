package tasks

import (
	"errors"
	"fmt"
	"os"

	"github.com/swaros/contxt/module/configure"
)

func (t *targetExecuter) SetFunctions(anko *AnkoRunner) {
	// make it possible, to stop the execution of the script, by its own
	anko.Define("exit", func() {
		anko.cancelationFn()
	})

	// ifos(string) implementation
	anko.Define("ifos", func(s string) bool {
		return configure.GetOs() == s
	})

	anko.Define("getos", func() string {
		return configure.GetOs()
	})

	// importJson(string,string) implementation
	anko.Define("importJson", func(key, json string) error {
		err := t.dataHandler.AddJSON(key, json)
		if err != nil {
			anko.ThrowException(err, fmt.Sprintf("importJson('%s','%s')", key, json))
			t.out(MsgError(MsgError{Err: errors.New("error while parsing json: " + err.Error()), Reference: "importJson(key,json)", Target: t.target}))
		}
		return err
	})
	// varAsJson(string) string implementation
	anko.Define("varAsJson", func(key string) string {
		data, _ := t.dataHandler.GetDataAsJson(key)
		return data
	})
	// varAsYaml(string) string implementation
	anko.Define("varAsYaml", func(key string) string {
		data, _ := t.dataHandler.GetDataAsYaml(key)
		return data
	})

	// exec(string) (string, int, error) implementation
	anko.Define("exec", func(cmd string) (string, int, error) {
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
	})

}
