package ctxshell

import "github.com/chzyer/readline"

type NativeCmd struct {
	Name          string
	Help          string
	ExecFunc      func(args []string) error
	CompleterFunc readline.DynamicCompleteFunc
}

func NewNativeCmd(name, help string, execFunc func(args []string) error) *NativeCmd {
	return &NativeCmd{
		Name:     name,
		Help:     help,
		ExecFunc: execFunc,
	}
}

func (t *NativeCmd) SetCompleterFunc(f readline.DynamicCompleteFunc) *NativeCmd {
	t.CompleterFunc = f
	return t
}

func (t *NativeCmd) Exec(args []string) error {
	return t.ExecFunc(args)
}

func (t *NativeCmd) GetHelp() string {
	return t.Help
}

func (t *NativeCmd) GetName() string {
	return t.Name
}
