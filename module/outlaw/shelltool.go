package outlaw

import (
	"github.com/swaros/contxt/configure"
	"github.com/swaros/contxt/dirhandle"
	"github.com/swaros/contxt/taskrun"
)

func autoRecoverWs() {
	if !inWs() {
		configure.WorkSpaces(func(ws string) {
			if configure.UsedConfig.CurrentSet == ws {
				configure.ChangeWorkspace(ws, taskrun.CallBackOldWs, taskrun.CallBackNewWs)
			}
		})
	}
}

func inWs() bool {
	dir, err := dirhandle.Current()
	if err != nil {
		panic(err)
	}
	return configure.PathMeightPartOfWs(dir)
}

func resetShell() {
	taskrun.ClearAll()
	taskrun.MainInit()
}
