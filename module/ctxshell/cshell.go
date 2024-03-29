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
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	UpdateByInit   = 1001
	UpdateByPeriod = 1002
	UpdateBySignal = 1003
	UpdateByInput  = 1004
	UpdateByNotify = 1005
)

type Cshell struct {
	cobraRootCmd         *cobra.Command     // the root command of the cobra command tree
	navtiveCmds          []*NativeCmd       // commands that are not part of the cobra command tree
	getPrompt            func(int) string   // function that returns the prompt string. the update type is passed as argument. if return is empty, the prompt is not updated
	exitCmdStr           string             // the command string that exits the shell
	rlInstance           *readline.Instance // the readline instance
	asyncCobraExec       bool               // if true, cobra commands are executed in a separate goroutine. a general rule.
	asyncNativeCmd       bool               // if true, native commands are executed in a separate goroutine. a general rule.
	tickTimerDuration    time.Duration      // the duration of the tick timer for print the buffered messages
	messages             *CshellMsgFifo     // the message buffer
	neverAsncCmds        []string           // commands that are never executed in a separate goroutine
	ignoreCobraCmds      []string           // commands that are ignored by the cobra command tree
	updatePromptDuration time.Duration      // the period for updating the prompt
	messageDisplayTime   time.Duration      // the time a message is displayed
	updatePromptEnabled  bool               // if true, the prompt is updated periodically
	lastInput            string             // the last input
	StopOutput           bool               // stop printing the output to stdout
	captureExitSignal    bool               // if true, the exit signal is captured
	keyBindings          []KeyFunc          // key bindings
	promptMessages       []Msg              // messages that are printed before the prompt
	currentMessage       Msg                // the current message
	currentMsgExpire     time.Time          // the time when the current message expires
	runOnceCmds          []string           // commands that are executed only once
	onShutDown           func()             // function that is called on shutdown
	onErrorFn            func(error)        // function that is called on error
	onUnknownCmd         func(string) error // function that is called on unknown command
	noMessageDuplication bool               // if true, messages are not duplicated for notifications
	hooks                []Hook             // hooks

}

// key binding struct
type KeyFunc struct {
	Key rune        // what key to bind
	Fn  func() bool // what function to call. returning false means do populate the key
}

// NewCshell creates a new cshell instance with default values.
// use the setter functions to change the default values before calling Run()
func NewCshell() *Cshell {
	return &Cshell{
		exitCmdStr:           "exit",
		tickTimerDuration:    100 * time.Millisecond,
		messages:             NewCshellMsgScope(100),
		updatePromptDuration: 5 * time.Second,
		messageDisplayTime:   2 * time.Second, // 2 seconds is default for displaying messages
		neverAsncCmds:        []string{},
	}
}

// add a key binding
// painic if the key is already in the list
func (t *Cshell) AddKeyBinding(key rune, fn func() bool) *Cshell {
	// check if the key is already in the list
	for _, k := range t.keyBindings {
		if k.Key == key {
			panic("AddKeyBinding duplicate. key already in list")
		}
	}
	t.keyBindings = append(t.keyBindings, KeyFunc{Key: key, Fn: fn})
	return t
}

// SetNoMessageDuplication enables or disables the duplication of messages.
// here it is all about internal messages that are displayed in the prompt.
// this notification messages.
// that means, if a message is already in the message buffer, it is not added again.
// this is useful if you want to avoid that the same message is displayed multiple times what is
// not really useful at some point.
// but on default we do not set this behavior, so any message is populated to the message buffer.
// also keep in mind, that this will only affect the message buffer. so it is still possible to
// display the same message multiple times. it just reduces spamming the message buffer.
func (t *Cshell) SetNoMessageDuplication(b bool) *Cshell {
	t.noMessageDuplication = b
	return t
}

// set a function that is called if the shell is shutdown. so here some cleanup can be done.
func (t *Cshell) OnShutDownFunc(fn func()) *Cshell {
	t.onShutDown = fn
	return t
}

// set a function that is called if the user enters an unknown command.
// this is useful if you want to implement a fallback for unknown commands.
func (t *Cshell) OnUnknownCmdFunc(fn func(string) error) *Cshell {
	t.onUnknownCmd = fn
	return t
}

// set a function that is called if an error occurs.
// this means any error that is not handled by the shell itself.
func (t *Cshell) OnErrorFunc(fn func(error)) *Cshell {
	t.onErrorFn = fn
	return t
}

// set the duration for displaying a message. this is ment for messages that are displayed
// inside the prompt. as long a message is displayed, the prompt will get updated with status UpdateByNotify.
// in this case the message could received by using GetCurrentMessage() where you get a bool if a message is
// currently displayed and the message itself.
// example:
//
//	shell := ctxshell.NewCshell()
//	shell.SetMessageDisplayTime(5 * time.Second)
//	shell.SetPromptFunc(func(reason int) string {
//		if reason == ctxshell.UpdateByNotify {
//			if haveAnMessage, msg := shell.GetCurrentMessage(); haveAnMessage {
//				return "info:[" + msg.GetMsg() + "] > "
//			}
//		}
//		return " prompt > "
//	})
//	shell.Run()
//
// this duration is setting the time a message is displayed. after this time the message is gone and the next message from the buffer is displayed.
// until any messages are handled.
// this is ment for having the messages displayed in a environment, that is constantly printing something to stdout,
// so it would be hard to see the messages in the stdout.
func (t *Cshell) SetMessageDisplayTime(d time.Duration) *Cshell {
	t.messageDisplayTime = d
	return t
}

// notify the prompt to display a message.
// the message is displayed for the time defined by SetMessageDisplayTime()
// this is also used internally to display messages nd errors
func (t *Cshell) NotifyToPrompt(message Msg) *Cshell {
	if t.noMessageDuplication {
		for _, m := range t.promptMessages {
			if m.GetMsg() == message.GetMsg() {
				return t
			}
		}
	}
	t.promptMessages = append(t.promptMessages, message)
	return t
}

// enable or disable the capture of the exit signal
func (t *Cshell) SetCaptureExitSignal(b bool) *Cshell {
	t.captureExitSignal = b
	return t
}

// enable or disable the output to stdout
func (t *Cshell) SetStopOutput(b bool) *Cshell {
	t.StopOutput = b
	return t
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

// set the duration for updating the prompt
func (t *Cshell) UpdatePromptPeriod(d time.Duration) *Cshell {
	t.updatePromptDuration = d
	return t
}

// enable or disable the periodic update of the prompt
func (t *Cshell) UpdatePromptEnabled(onoff bool) *Cshell {
	t.updatePromptEnabled = onoff
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

// returns the last input
func (t *Cshell) GetLastInput() string {
	return t.lastInput
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
	t.cobraRootCmd = cmd
	return t
}

// add a native command to the shell. native commands are just a wrapper
func (t *Cshell) AddNativeCmd(cmd *NativeCmd) *Cshell {
	t.navtiveCmds = append(t.navtiveCmds, cmd)
	return t
}

// set the prompt function. this function is called every time the prompt
// is printed. so you can change the prompt string dynamically
func (t *Cshell) SetPromptFunc(f func(int) string) *Cshell {
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

// get the readline instance
func (t *Cshell) GetReadline() *readline.Instance {
	return t.rlInstance
}

// add message to the message buffer
func (t *Cshell) Stdout(messg string) {
	t.messages.Push("stdout", messg)
}

// add a stdout message to the message buffer
// the message is terminated with a newline
func (t *Cshell) Stdoutln(messg string) {
	t.messages.Push("stdout", messg+"\n")
}

// add a stderr message to the message buffer
func (t *Cshell) Stderr(messg string) {
	t.messages.Push("stderr", messg)
}

// add a stderr message to the message buffer
// the message is terminated with a newline
func (t *Cshell) Stderrln(messg string) {
	t.messages.Push("stderr", messg+"\n")
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
	if t.cobraRootCmd != nil {
		// get all commands from the root command and iterate over them
		for _, c := range t.cobraRootCmd.Commands() {
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
		sliceRes, _ := c.ValidArgsFunction(t.cobraRootCmd, []string{}, "")

		for _, s := range sliceRes {
			// replace tab with space
			s = strings.ReplaceAll(s, "\t", " ")
			// we want only the first word of the result
			s = strings.Split(s, " ")[0]
			newCmd.Children = append(newCmd.Children, readline.PcItem(s))
		}
	}
	if c.ValidArgs != nil {
		for _, s := range c.ValidArgs {
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
	// create a tempfile name that is based os the binary name
	binFile := os.Args[0]
	// just replace any slashes, spaces and colons with underscores
	binFile = strings.ReplaceAll(binFile, "/", "_")
	binFile = strings.ReplaceAll(binFile, " ", "_")
	binFile = strings.ReplaceAll(binFile, "\\", "_")
	binFile = strings.ReplaceAll(binFile, ":", "_")

	var err error
	t.rlInstance, err = readline.NewEx(&readline.Config{
		Prompt:              " > ",
		HistoryFile:         filepath.Clean(os.TempDir() + "/cshell_history_" + binFile + ".tmp"),
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: t.inputFilterFunc, // this is handlind the key-bindings
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

// RunOnce executes the given commands and exits the shell afterwards.
// this is useful if you want to execute a command from a script.
func (t *Cshell) RunOnce(cmds []string) error {
	// skip if no cmd is given
	if len(cmds) == 0 {
		return nil
	}
	t.runOnceCmds = cmds
	return t.runShell(true)
}

// update the prompt
func (t *Cshell) updatePrompt(reason int) {
	if t.getPrompt != nil {
		if prmpt := t.getPrompt(reason); prmpt != "" {
			t.rlInstance.SetPrompt(t.getPrompt(reason))
		}
	}
}

// Error handler function to register time based messages for the prompt, and
// executing the Error function if defined
func (t *Cshell) Error(messages ...string) {
	t.NotifyToPrompt(DefaultPromptMessage(strings.Join(messages, " "), TopicError, t.messageDisplayTime))
	if t.onErrorFn != nil {
		errorMsg := strings.Join(messages, " ")
		t.onErrorFn(errors.New(errorMsg))
	}
}

// Message handler function to register time based messages for the prompt
func (t *Cshell) Message(messages ...string) {
	t.NotifyToPrompt(DefaultPromptMessage(strings.Join(messages, " "), TopicInfo, t.messageDisplayTime))
}

// Message handler function to register time based messages for the prompt
func (t *Cshell) TimedMessage(displayTime time.Duration, messages ...string) {
	t.NotifyToPrompt(DefaultPromptMessage(strings.Join(messages, " "), TopicInfo, displayTime))
}

func (t *Cshell) getOnceCmd() string {
	if len(t.runOnceCmds) > 0 {
		cmd := t.runOnceCmds[0]
		t.runOnceCmds = t.runOnceCmds[1:]
		return cmd
	}
	return ""
}

// check if we still have commands to execute
// from the list of commands they are pre defined
func (t *Cshell) haveOnce() bool {
	return len(t.runOnceCmds) > 0
}

// Run starts the shell
func (t *Cshell) Run() error {
	return t.runShell(false)
}

func (t *Cshell) runShell(once bool) error {
	err := t.init()
	if err != nil {
		return err
	}
	defer t.rlInstance.Close()
	// if we want to capture the exit signal, we have to do it here
	if t.captureExitSignal {
		t.rlInstance.CaptureExitSignal()
	}
	log.SetOutput(t.rlInstance.Stderr())
	t.updatePrompt(UpdateByInit)
	// start the message provider they prints the messages
	// any time defined by tickTimerDuration
	t.StartMessageProvider()
	t.StartBackgroundPromptUpate()
	// the main loop
	for {
		cmdPreset := ""
		if once {
			if cmd := t.getOnceCmd(); cmd != "" {
				cmdPreset = cmd
			} else {
				break
			}
		}
		var ln *readline.Result
		if cmdPreset != "" {
			ln = &readline.Result{
				Line:  cmdPreset,
				Error: nil,
			}
		} else {
			ln = t.rlInstance.Line()
		}
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

		// cleanup the input
		line = strings.TrimSpace(line)
		lineCmd := strings.Split(line, " ")[0]
		fullArgs := strings.Split(line, " ")
		t.lastInput = line
		t.executeHooksBefore(t.GetHooksByPattern(lineCmd))
		// reset the flag that indicates if we did something
		weDidSomething := false
		// native commands just a wrapper for the completion
		// together with the exec function
		if c := t.getNativeCmd(lineCmd); c != nil {
			weDidSomething = true
			if t.asyncNativeCmd && !t.isNeverAsyncCmd(lineCmd) {
				go func() {
					if err := c.Exec(fullArgs); err != nil {
						t.Error("error executing native command:", err.Error())
					}
					t.Message(lineCmd, "done")
				}()
			} else {
				if err := c.Exec(fullArgs); err != nil {
					t.Error("error executing native command:", err.Error())
				}
				t.Message(lineCmd, "done")
			}
			continue
		}
		// check if we deal with an cobra command
		// so we do not execute the root command, because we would not
		// know if this is an valid command or not
		if t.cobraRootCmd != nil {
			for _, c := range t.cobraRootCmd.Commands() {
				// the name is in the list of cobra commands
				// rest is the args

				if c.Name() == lineCmd {
					weDidSomething = true
					t.cobraRootCmd.SetArgs(strings.Split(line, " "))
					if t.asyncCobraExec && !t.isNeverAsyncCmd(c.Name()) {
						go func() {
							if err := t.cobraRootCmd.Execute(); err != nil {
								t.Error("error executing cobra command:", err.Error())
							}
							t.Message(lineCmd, "done")
						}()
					} else {
						if err := t.cobraRootCmd.Execute(); err != nil {
							t.Error("error executing cobra command:", err.Error())
						}
						t.Message(lineCmd, "done")
					}
					continue
				}
			}
		}
		t.executeHooksAfter(t.GetHooksByPattern(lineCmd))

		// if we are here, we have no idea what to do
		if !weDidSomething {
			// if we have a function that is called on unknown command, we call it
			// and uses the return value as error if returned
			if t.onUnknownCmd != nil {
				if err := t.onUnknownCmd(line); err != nil {
					t.Error(err.Error())
				}
			} else {
				t.Error("unknown command:", lineCmd)
			}
		}
		// move to the next line
		t.rlInstance.Write([]byte("\n"))
		t.updatePrompt(UpdateByInput)

		if once && !t.haveOnce() {
			break
		}

	}
	// leave the shell after the loop
	if t.onShutDown != nil {
		t.onShutDown()
	}
	return nil
}

func (t *Cshell) tryGetPromptMessage() (Msg, bool) {
	if len(t.promptMessages) > 0 {
		msg := t.promptMessages[0]
		t.promptMessages = t.promptMessages[1:]
		t.currentMessage = msg
		t.currentMsgExpire = time.Now().Add(msg.GetTimeToDisplay())
		return msg, true
	}
	t.currentMessage = Msg{}
	t.currentMsgExpire = time.Now().Add(-1 * time.Second)
	return Msg{}, false
}

// Gettig the Message from the message buffer that is currently displayed
func (t *Cshell) GetCurrentMessage() (bool, Msg) {
	if time.Now().Before(t.currentMsgExpire) {
		return true, t.currentMessage
	}
	return false, Msg{}
}

// this is the promt update loop
// it updates the prompt every updatePromptDuration
// use UpdatePromptPeriod to set the updatePromptDuration
// use UpdatePromptEnabled to enable or disable the prompt update
func (t *Cshell) StartBackgroundPromptUpate() {
	if t.rlInstance == nil {
		return
	}

	done := make(chan struct{})
	go func(tp *Cshell) {

	promptLoop:
		for {
			select {
			case <-time.After(time.Duration(tp.updatePromptDuration)):
				// only update the prompt if we are not in complete mode and is it enabled
				if tp.updatePromptEnabled && !tp.rlInstance.Operation.IsInCompleteMode() {

					// lets check if we have to provide some messages to the prompt
					// first we need to check if a message is currently displayed
					// and still valid. if this is the case we do nothing
					// if not, we check if we have a message in the buffer
					promtUpdated := false
					if haveAnMessage, _ := tp.GetCurrentMessage(); !haveAnMessage {
						// so currently there is no message active. just check if we have a message in the buffer
						if _, found := tp.tryGetPromptMessage(); found {
							promtUpdated = true
							tp.updatePrompt(UpdateByNotify)
						}
					} else {
						// we have a message active. this will be also published by the message provider
						promtUpdated = true
						tp.updatePrompt(UpdateByNotify)
					}
					if !promtUpdated {
						tp.updatePrompt(UpdateByPeriod)
					}
					tp.rlInstance.Refresh()
				}
			case <-done:
				break promptLoop
			}
		}
		done <- struct{}{}
	}(t)

}

// filters the key bindings and executes the function
// if the key is not in the list, the key is returned and the bool is true
func (t *Cshell) inputFilterFunc(r rune) (rune, bool) {
	// check if we have a key binding
	for _, k := range t.keyBindings {
		if k.Key == r {
			return r, k.Fn()
		}
	}
	return r, true
}
