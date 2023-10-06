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

// defines commands that are never executed in a separate goroutine
func (t *Cshell) SetNeverAsyncCmds(cmds []string) *Cshell {
	t.neverAsncCmds = cmds
	return t
}

// defines a command that are never executed in a separate goroutine
func (t *Cshell) SetNeverAsyncCmd(cmd ...string) *Cshell {
	t.neverAsncCmds = append(t.neverAsncCmds, cmd...)
	return t
}

// checks if a command is never executed in a separate goroutine
func (t *Cshell) isNeverAsyncCmd(cmd string) bool {
	for _, c := range t.neverAsncCmds {
		if c == cmd {
			return true
		}
	}
	return false
}

// defines commands that are ignored by the cobra command tree
func (t *Cshell) SetIgnoreCobraCmds(cmds []string) *Cshell {
	t.ignoreCobraCmds = cmds
	return t
}

// defines commands they are ignored by the cobra command tree
func (t *Cshell) SetIgnoreCobraCmd(cmd ...string) *Cshell {
	t.ignoreCobraCmds = append(t.ignoreCobraCmds, cmd...)
	return t
}

// checks if a command is ignored by the cobra command tree
func (t *Cshell) isIgnoreCobraCmd(cmd string) bool {
	for _, c := range t.ignoreCobraCmds {
		if c == cmd {
			return true
		}
	}
	return false
}

// this sets the duration of the tick timer for print the buffered messages
// this is required because the readline instance is not thread safe.
// so we have to collect any output from any threads and print it in the main thread synchronously
func (t *Cshell) SetTickTimerDuration(d time.Duration) *Cshell {
	t.tickTimerDuration = d
	return t
}

// this resizes the message buffer
func (t *Cshell) ResizeMessageProvider(size int) *Cshell {
	t.messages = NewCshellMsgScope(size)
	return t
}

// set the cobra root command. this is the root of the cobra command tree
// it is used to parse the command line and execute the commands, and
// to get the flags and arguments for the completer
func (t *Cshell) SetCobraRootCommand(cmd *cobra.Command) *Cshell {
	t.CobraRootCmd = cmd
	return t
}

// add a native command to the shell. native commands are just a wrapper
func (t *Cshell) AddNativeCmd(cmd *NativeCmd) *Cshell {
	t.navtiveCmds = append(t.navtiveCmds, cmd)
	return t
}

// set the prompt function. this function is called every time the prompt
// is printed. so you can change the prompt string dynamically
func (t *Cshell) SetPromptFunc(f func() string) *Cshell {
	t.getPrompt = f
	return t
}

// set the exit command string. this is the command string that exits the shell
func (t *Cshell) SetExitCmdStr(s string) *Cshell {
	t.exitCmdStr = s
	return t
}

// enable or disable the async execution of cobra commands
func (t *Cshell) SetAsyncCobraExec(b bool) *Cshell {
	t.asyncCobraExec = b
	return t
}

// enable or disable the async execution of native commands
func (t *Cshell) SetAsyncNativeCmd(b bool) *Cshell {
	t.asyncNativeCmd = b
	return t
}

// we initialize the completer with the native commands and the cobra command tree
// the native commands are just a wrapper for the completion
// together with the exec function
// the cobra command tree is parsed and the commands are added to the completer
// the flags are added to the completer too
// the completer function of the cobra command is called only once and the result
// is added to the completer
func (t *Cshell) createCompleter() *readline.PrefixCompleter {
	completer := readline.NewPrefixCompleter()
	for _, c := range t.navtiveCmds {
		if c.CompleterFunc != nil {
			nativeCmd := readline.PcItem(c.Name)
			nativeCmd.Callback = c.CompleterFunc
			completer.Children = append(completer.Children, nativeCmd)
		}

	}

	// parsing the cobra command tree
	// start with the root command
	if t.CobraRootCmd != nil {
		// get all commands from the root command and iterate over them
		for _, c := range t.CobraRootCmd.Commands() {
			// ignore commands that are in the ignore list
			if t.isIgnoreCobraCmd(c.Name()) {
				continue
			}
			// create a new completer item
			newCmd := readline.PcItem(c.Name())
			// if the command has subcommands, we have to create a completer for them
			if c.HasSubCommands() {
				newCmd = t.createSubCommandCompleter(newCmd, c)
			}
			// add the flags from the cobra command to the completer
			c.Flags().VisitAll(func(f *pflag.Flag) {
				if f.Shorthand != "" {
					newCmd.Children = append(newCmd.Children, readline.PcItem("-"+f.Shorthand))
				}
				if f.Name != "" {
					newCmd.Children = append(newCmd.Children, readline.PcItem("--"+f.Name))
				}
			})
			t.ApplyCobraCompletionOnce(newCmd, c)
			completer.Children = append(completer.Children, newCmd)
		}
	}

	return completer
}

// createSubCommandCompleter creates a completer for the subcommands of the cobra command
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
		// check if the command has a completer function
		t.ApplyCobraCompletionOnce(newCmd, c)
		compl.Children = append(compl.Children, newCmd)
	}
	return compl
}

// ApplyCobraCompletionOnce applies the completer function of the cobra command
// only once. this is necessary because the completer function is called
// every time the user hits the tab key. so just read the valid args once
// and add them to the completer item.
// that means, if somehow the valid args are changing, the user have to restart
// the shell.
func (t *Cshell) ApplyCobraCompletionOnce(newCmd *readline.PrefixCompleter, c *cobra.Command) {
	if c.ValidArgsFunction != nil {
		// if yes, we have to add the completer function to the completer item.
		// we execute the completer function with an empty slice of strings
		// and get a slice of strings back. we add them to the completer item
		sliceRes, _ := c.ValidArgsFunction(t.CobraRootCmd, []string{}, "")

		for _, s := range sliceRes {
			// replace tab with space
			s = strings.ReplaceAll(s, "\t", " ")
			// we want only the first word of the result
			s = strings.Split(s, " ")[0]
			newCmd.Children = append(newCmd.Children, readline.PcItem(s))
		}
	}
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
	// create the completer
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
