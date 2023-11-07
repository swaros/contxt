package terminal

import (
	"errors"
	"runtime"
)

type TermFind struct {
	matchingOs     []string // the os to match
	cmd            string   // the command to run
	args           []string // the arguments to pass to the command and execute the commands. like the -c on bash
	argsToKeepOpen []string // the arguments to pass to the command to keep it open. like without -c for bash
}

var (
	termFindMap = map[string]*TermFind{
		"windows": {
			matchingOs:     []string{"windows"},
			cmd:            "powershell",
			args:           []string{"-nologo", "-noprofile"},
			argsToKeepOpen: []string{"-nologo", "-noprofile", "-noexit", "-NonInteractive"},
		},
		"linux": {
			matchingOs:     []string{"linux", "darwin"},
			cmd:            "bash",
			args:           []string{"-c"},
			argsToKeepOpen: []string{},
		},
	}
	ErrNoTerminalFound = errors.New("no terminal found")
)

// GetTerminal returns the terminal finder for the current os
// if the terminal is not found by the os keyword
// it will try to find it by the os name
func GetTerminal() (*TermFind, error) {
	if termFind, ok := termFindMap[runtime.GOOS]; ok {
		return termFind, nil
	} else {
		// did not found the terminal by the keyword
		// so we will try to find it by the os
		for _, termFind := range termFindMap {
			for _, os := range termFind.matchingOs {
				if os == runtime.GOOS {
					return termFind, nil
				}
			}
		}
	}
	return nil, ErrNoTerminalFound
}

func (t *TermFind) GetCmd() string {
	return t.cmd
}

func (t *TermFind) GetArgs() []string {
	return t.args
}

func (t *TermFind) GetArgsToKeepOpen() []string {
	return t.argsToKeepOpen
}

func (t *TermFind) CombineArgs(args ...string) []string {
	cmds := []string{}

	cmds = append(cmds, t.args...)
	cmds = append(cmds, args...)
	return cmds
}
