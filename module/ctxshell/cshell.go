package ctxshell

import (
	"io"
	"log"
	"strings"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Cshell struct {
	CobraRootCmd *cobra.Command
	cobraCmdList []string // just to remember if we deal with an cobra command
	navtiveCmds  []*NativeCmd
	getPrompt    func() string
}

func NewCshell() *Cshell {
	return &Cshell{}
}

func (t *Cshell) SetCobraRootCommand(cmd *cobra.Command) *Cshell {
	t.CobraRootCmd = cmd
	return t
}

func (t *Cshell) AddNativeCmd(cmd *NativeCmd) *Cshell {
	t.navtiveCmds = append(t.navtiveCmds, cmd)
	return t
}

func (t *Cshell) SetPromptFunc(f func() string) *Cshell {
	t.getPrompt = f
	return t
}

func (t *Cshell) createCompleter() *readline.PrefixCompleter {
	completer := readline.NewPrefixCompleter()
	for _, c := range t.navtiveCmds {
		if c.CompleterFunc != nil {
			nativeCmd := readline.PcItem(c.Name)
			nativeCmd.Callback = c.CompleterFunc
			completer.Children = append(completer.Children, nativeCmd)
		}

	}

	if t.CobraRootCmd != nil {
		for _, c := range t.CobraRootCmd.Commands() {
			t.cobraCmdList = append(t.cobraCmdList, c.Name())
			newCmd := readline.PcItem(c.Name())
			if c.HasSubCommands() {
				newCmd = t.createSubCommandCompleter(newCmd, c)
			}
			c.Flags().VisitAll(func(f *pflag.Flag) {
				if f.Shorthand != "" {
					newCmd.Children = append(newCmd.Children, readline.PcItem("-"+f.Shorthand))
				}
				if f.Name != "" {
					newCmd.Children = append(newCmd.Children, readline.PcItem("--"+f.Name))
				}
			})

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
		c.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Shorthand != "" {
				newCmd.Children = append(newCmd.Children, readline.PcItem("-"+f.Shorthand))
			}
			if f.Name != "" {
				newCmd.Children = append(newCmd.Children, readline.PcItem("--"+f.Name))
			}
		})
		compl.Children = append(compl.Children, newCmd)
	}
	return compl
}

// cmd have to be taken from the first word of the comand line.
// e.g. "cmd arg1 arg2" -> "cmd"
func (t *Cshell) getNativeCmd(cmd string) *NativeCmd {
	for _, c := range t.navtiveCmds {
		if c.Name == cmd {
			return c
		}
	}
	return nil
}

// Run starts the shell
func (t *Cshell) Run() error {
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
		return err
	}
	defer l.Close()
	l.CaptureExitSignal()
	log.SetOutput(l.Stderr())
	if t.getPrompt != nil {
		l.SetPrompt(t.getPrompt())
	}
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
		lineCmd := strings.Split(line, " ")[0]
		fullArgs := strings.Split(line, " ")
		switch {
		case line == "exit":
			return nil
		case line == "help":
			l.Write([]byte("help\n"))
		default:
			if c := t.getNativeCmd(lineCmd); c != nil {
				c.Exec(fullArgs)
				continue
			}
			// check if we deal with an cobra command
			// so we do not execute the root command, because we would not
			// know if this is an valid command or not
			if t.CobraRootCmd != nil {
				for _, c := range t.CobraRootCmd.Commands() {
					// the name is in the list of cobra commands
					// rest is the args

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
	return nil
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}
