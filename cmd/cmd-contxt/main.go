package main

import (
	"os"

	"github.com/swaros/contxt/outlaw"
	"github.com/swaros/contxt/taskrun"
)

func main() {
	if len(os.Args) > 1 {
		taskrun.MainExecute()
	} else {
		outlaw.RunIShell()
	}

}
