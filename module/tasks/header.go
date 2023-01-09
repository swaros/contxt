package tasks

import (
	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
)

const (
	Target                = "target"
	Arguments             = "arguments"
	Script                = "script"
	RunCfg                = "runCfg"
	StopReason            = "stopReason"
	MainCmd               = "mainCmd"
	MainCmdArgs           = "mainCmdArgs"
	PlaceholderHandler    = "placeholderHandler"
	PlaceholderSetHandler = "placeholderSetHandler"
	OutputHandler         = "outputHandler"
	DataMapHandl          = "dataMapHandler"
)

type MainCmdSetter interface {
	GetMainCmd() (string, []string)
}

var (
	emptyMainCmdSetter MainCmdSetter = emptyCmd{}
)

type targetExecuter struct {
	target          string
	arguments       map[string]string
	runCfg          configure.RunConfig
	mainCmd         string
	mainCmdArgs     []string
	phHandler       PlaceHolder
	outputHandler   func(msg ...interface{})
	reasonCheck     func(checkReason configure.Trigger, output string, e error) (bool, string)
	checkReqs       func(require configure.Require) (bool, string)
	Logger          *logrus.Logger
	dataHandler     DataMapHandler
	watch           *Watchman
	commandFallback MainCmdSetter
}

type emptyCmd struct{}

func (e emptyCmd) GetMainCmd() (string, []string) {
	return "", []string{}
}

func New(target string, arguments map[string]string, any ...interface{}) *targetExecuter {

	t := &targetExecuter{
		target:    target,
		arguments: arguments,
	}

	for i := 0; i < len(any); i++ {
		switch any[i].(type) {

		case configure.RunConfig:
			t.runCfg = any[i].(configure.RunConfig)

		case PlaceHolder:
			t.phHandler = any[i].(PlaceHolder)
			// check if if any[i] also implements the DataMapHandler interface
			// if so, and we do not have a data handler set yet
			// we set it to the one from the PlaceHolder
			if t.dataHandler == nil {
				if dm, ok := any[i].(DataMapHandler); ok {
					t.dataHandler = dm
				}
			}
		case func(msg ...interface{}):
			t.outputHandler = any[i].(func(msg ...interface{}))
		case func(checkReason configure.Trigger, output string, e error) (bool, string):
			t.reasonCheck = any[i].(func(checkReason configure.Trigger, output string, e error) (bool, string))
		case func(require configure.Require) (bool, string):
			t.checkReqs = any[i].(func(require configure.Require) (bool, string))
		case DataMapHandler:
			t.dataHandler = any[i].(DataMapHandler)
			// check if if any[i] also implements the PlaceHolder interface
			// if so, and we do not have a placeholder handler set yet
			// we set it to the one from the DataMapHandler
			if t.phHandler == nil {
				if ph, ok := any[i].(PlaceHolder); ok {
					t.phHandler = ph
				}
			}
		case *Watchman:
			t.watch = any[i].(*Watchman)
		case MainCmdSetter:
			t.commandFallback = any[i].(MainCmdSetter)
		default:
			panic("Invalid type passed to New")
		}
	}

	t.reInitialize()
	return t
}

func (t *targetExecuter) SetMainCmd(mainCmd string, args ...string) *targetExecuter {
	t.mainCmd = mainCmd
	t.mainCmdArgs = args
	return t
}

// reInitialize is used to reinitialize the targetExecuter
// so it assigns the required fields depending the given arguments
// and also make sure, any required field is set
// if they can have a default value.
func (t *targetExecuter) reInitialize() {
	// this just returns the emptyCmd struct
	// so we can use it as a fallback
	// but will not usable so we have to warn the user
	if t.commandFallback == nil {
		t.commandFallback = emptyMainCmdSetter
		t.getLogger().Warn("No MainCmdSetter provided, using empty fallback")
	}
	// if no task watcher is set, we create a new one
	if t.watch == nil {
		t.watch = NewWatchman()
	}
	// assign the Tasks to the targetExecuter

}

func (t *targetExecuter) CopyToTarget(target string) *targetExecuter {
	copy := New(
		target,
		t.arguments,
		t.runCfg,
		t.mainCmd,
		t.mainCmdArgs,
		t.phHandler,
		t.outputHandler,
		t.reasonCheck,
		t.checkReqs,
	)
	copy.watch = t.watch
	copy.dataHandler = t.dataHandler
	return copy
}

func (t *targetExecuter) SetLogger(logger *logrus.Logger) *targetExecuter {
	t.Logger = logger
	return t
}

func (t *targetExecuter) SetDataHandler(handler DataMapHandler) *targetExecuter {
	t.dataHandler = handler
	return t
}

func (t *targetExecuter) SetPlaceholderHandler(handler PlaceHolder) *targetExecuter {
	t.phHandler = handler
	return t
}

func (t *targetExecuter) SetWatchman(watch *Watchman) *targetExecuter {
	t.watch = watch
	return t
}

// Create Setter for any Property from the targetExecuter
func (t *targetExecuter) SetProperty(property string, value interface{}) *targetExecuter {
	switch property {
	case Target:
		t.target = value.(string)
	case Arguments:
		t.arguments = value.(map[string]string)
	case RunCfg:
		t.runCfg = value.(configure.RunConfig)
	case MainCmd:
		t.mainCmd = value.(string)
	case MainCmdArgs:
		t.mainCmdArgs = value.([]string)
	case PlaceholderHandler:
		t.phHandler = value.(PlaceHolder)
	case OutputHandler:
		t.outputHandler = value.(func(msg ...interface{}))
	case DataMapHandl:
		t.dataHandler = value.(DataMapHandler)
	default:
		panic("Unknown Property: " + property)
	}
	return t
}
