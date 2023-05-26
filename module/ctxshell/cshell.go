package ctxshell

import (
	"log"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Cshell struct {
	CobraRootCmd      *cobra.Command     // the root command of the cobra command tree
	navtiveCmds       []*NativeCmd       // commands that are not part of the cobra command tree
	getPrompt         func() string      // function that returns the prompt string
	exitCmdStr        string             // the command string that exits the shell
	rlInstance        *readline.Instance // the readline instance
	asyncCobraExec    bool               // if true, cobra commands are executed in a separate goroutine. a general rule.
	asyncNativeCmd    bool               // if true, native commands are executed in a separate goroutine. a general rule.
	tickTimerDuration time.Duration      // the duration of the tick timer for print the buffered messages
	messages          *CshellMsgFifo     // the message buffer
	neverAsncCmds     []string           // commands that are never executed in a separate goroutine
	ignoreCobraCmds   []string           // commands that are ignored by the cobra command tree

}

func NewCshell() *Cshell {
	return &Cshell{
		exitCmdStr:        "exit",
		tickTimerDuration: 100 * time.Millisecond,
		messages:          NewCshellMsgScope(100),
		neverAsncCmds:     []string{},
	}
}

func (t *Cshell) SetNeverAsyncCmds(cmds []string) *Cshell {
	t.neverAsncCmds = cmds
	return t
}

func (t *Cshell) SetNeverAsyncCmd(cmd ...string) *Cshell {
	t.neverAsncCmds = append(t.neverAsncCmds, cmd...)
	return t
}

func (t *Cshell) isNeverAsyncCmd(cmd string) bool {
	for _, c := range t.neverAsncCmds {
		if c == cmd {
			return true
		}
	}
	return false
}

func (t *Cshell) SetIgnoreCobraCmds(cmds []string) *Cshell {
	t.ignoreCobraCmds = cmds
	return t
}

func (t *Cshell) SetIgnoreCobraCmd(cmd ...string) *Cshell {
	t.ignoreCobraCmds = append(t.ignoreCobraCmds, cmd...)
	return t
}

func (t *Cshell) isIgnoreCobraCmd(cmd string) bool {
	for _, c := range t.ignoreCobraCmds {
		if c == cmd {
			return true
		}
	}
	return false
}

func (t *Cshell) SetTickTimerDuration(d time.Duration) *Cshell {
	t.tickTimerDuration = d
	return t
}

func (t *Cshell) ResizeMessageProvider(size int) *Cshell {
	t.messages = NewCshellMsgScope(size)
	return t
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

func (t *Cshell) SetExitCmdStr(s string) *Cshell {
	t.exitCmdStr = s
	return t
}

func (t *Cshell) SetAsyncCobraExec(b bool) *Cshell {
	t.asyncCobraExec = b
	return t
}

func (t *Cshell) SetAsyncNativeCmd(b bool) *Cshell {
	t.asyncNativeCmd = b
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
			if t.isIgnoreCobraCmd(c.Name()) {
				continue
			}
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

func (t *Cshell) init() error {
	completer := t.createCompleter()
	var err error
	t.rlInstance, err = readline.NewEx(&readline.Config{
		Prompt:              " > ",
		HistoryFile:         "/tmp/readline.tmp",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
		UniqueEditLine:      true,
	})
	return err
}

func (t *Cshell) RunOnceWithCmd(cmd func()) error {
	// skip if no cmd is given
	if cmd == nil {
		return nil
	}
	if err := t.init(); err != nil {
		return err
	}
	defer t.rlInstance.Close()
	defer t.messages.FlushAndClose()
	cmd()
	return nil

}

// Run starts the shell
func (t *Cshell) Run() error {
	completer := t.createCompleter()
	var err error
	t.rlInstance, err = readline.NewEx(&readline.Config{
		Prompt:              " > ",
		HistoryFile:         "/tmp/readline.tmp",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
		UniqueEditLine:      true,
	})
	if err != nil {
		return err
	}
	defer t.rlInstance.Close()
	t.rlInstance.CaptureExitSignal()
	log.SetOutput(t.rlInstance.Stderr())
	if t.getPrompt != nil {
		t.rlInstance.SetPrompt(t.getPrompt())
	}
	// start the message provider they prints the messages
	// any time defined by tickTimerDuration
	t.StartMessageProvider()

	// the main loop
	for {

		ln := t.rlInstance.Line()
		if ln.CanContinue() {
			continue
		} else if ln.CanBreak() {
			break
		}
		line := ln.Line
		// skip empty lines
		if line == "" {
			continue
		}

		// get out by typing exit
		if line == t.exitCmdStr {
			break
		}

		line = strings.TrimSpace(line)
		lineCmd := strings.Split(line, " ")[0]
		fullArgs := strings.Split(line, " ")
		weDidSomething := false
		// native commands just a wrapper for the completion
		// together with the exec function
		if c := t.getNativeCmd(lineCmd); c != nil {
			weDidSomething = true
			if t.asyncNativeCmd && !t.isNeverAsyncCmd(lineCmd) {
				go func() {
					if err := c.Exec(fullArgs); err != nil {
						log.Printf("error executing command: %s", err)
					}
					t.rlInstance.Write([]byte(lineCmd + "..done\n"))
				}()
			} else {
				if err := c.Exec(fullArgs); err != nil {
					log.Printf("error executing command: %s", err)
				}
				t.rlInstance.Write([]byte(lineCmd + "..done\n"))
			}
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
					weDidSomething = true
					t.CobraRootCmd.SetArgs(strings.Split(line, " "))
					if t.asyncCobraExec && !t.isNeverAsyncCmd(c.Name()) {
						go func() {
							if err := t.CobraRootCmd.Execute(); err != nil {
								log.Printf("error executing command: %s", err)
							}
							t.rlInstance.Write([]byte(lineCmd + "..done\n"))
						}()
					} else {
						if err := t.CobraRootCmd.Execute(); err != nil {
							log.Printf("error executing command: %s", err)
						}
						t.rlInstance.Write([]byte(lineCmd + "..done\n"))
					}
					continue
				}
			}
		}
		// if we are here, we have no idea what to do
		if !weDidSomething {
			log.Printf("unknown command: %s", lineCmd)
		}
		// move to the next line
		t.rlInstance.Write([]byte("\n"))
		if t.getPrompt != nil {
			t.rlInstance.SetPrompt(t.getPrompt())
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
