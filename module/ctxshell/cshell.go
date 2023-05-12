package ctxshell

import (
	"io"
	"log"
	"strings"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

type Cshell struct {
	CobraRootCmd *cobra.Command
	cobraCmdList []string // just to remember if we deal with an cobra command
}

func NewCshell() *Cshell {
	return &Cshell{}
}

func (t *Cshell) SetCobraRootCommand(cmd *cobra.Command) *Cshell {
	t.CobraRootCmd = cmd
	return t
}

func (t *Cshell) createCompleter() *readline.PrefixCompleter {
	completer := readline.NewPrefixCompleter(
		readline.PcItem("run"),
		readline.PcItem("help"),
		readline.PcItem("exit"),
	)

	if t.CobraRootCmd != nil {
		for _, c := range t.CobraRootCmd.Commands() {
			t.cobraCmdList = append(t.cobraCmdList, c.Name())
			newCmd := readline.PcItem(c.Name())
			if c.HasSubCommands() {
				newCmd = t.createSubCommandCompleter(newCmd, c)
			}
			completer.Children = append(completer.Children, newCmd)
		}
	}

	return completer
}

func (t *Cshell) createSubCommandCompleter(compl *readline.PrefixCompleter, cmd *cobra.Command) *readline.PrefixCompleter {
	for _, c := range cmd.Commands() {
		newCmd := readline.PcItem(c.Name())
		if c.HasSubCommands() {
			newCmd = t.createSubCommandCompleter(newCmd, c)
		}
		compl.Children = append(compl.Children, newCmd)
	}
	return compl
}

func (t *Cshell) Run() {
	completer := t.createCompleter()
	l, err := readline.NewEx(&readline.Config{
		Prompt:          " CTX \033[31m»\033[0m ",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	l.CaptureExitSignal()
	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		switch {
		case line == "exit":
			return
		case line == "help":
			l.Write([]byte("help\n"))
		default:
			// check if we deal with an cobra command
			if t.CobraRootCmd != nil {
				for _, c := range t.CobraRootCmd.Commands() {
					// the name is in the list of cobra commands
					// rest is the args

					lineCmd := strings.Split(line, " ")[0]
					if c.Name() == lineCmd {
						t.CobraRootCmd.SetArgs(strings.Split(line, " "))
						t.CobraRootCmd.Execute()
						continue
					}
				}
			} else {
				l.Write([]byte("unknown: " + line + "\n"))
			}
			l.Write([]byte("\n\n"))
		}
	}

}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}
