package tasks

import (
	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
)

const (
	CodeLine              = "codeLine"
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

type targetExecuter struct {
	target        string
	arguments     map[string]string
	script        configure.Task
	runCfg        configure.RunConfig
	stopReason    configure.Trigger
	codeLine      string
	mainCmd       string
	mainCmdArgs   []string
	phHandler     PlaceHolder
	outputHandler func(msg ...interface{})
	reasonCheck   func(checkReason configure.Trigger, output string, e error) (bool, string)
	checkReqs     func(require configure.Require) (bool, string)
	Logger        *logrus.Logger
	dataHandler   DataMapHandler
	watch         *Watchman
}

func New(target string,
	arguments map[string]string,
	script configure.Task,
	runCfg configure.RunConfig,
	stopReason configure.Trigger,
	codeLine string,
	mainCmd string,
	mainCmdArgs []string,
	placeholderHandler PlaceHolder,
	outputHandler func(msg ...interface{}),
	reasonCheck func(checkReason configure.Trigger, output string, e error) (bool, string),
	requirementCheck func(require configure.Require) (bool, string),
) *targetExecuter {
	return &targetExecuter{
		target:        target,
		arguments:     arguments,
		script:        script,
		runCfg:        runCfg,
		stopReason:    stopReason,
		codeLine:      codeLine,
		mainCmd:       mainCmd,
		mainCmdArgs:   mainCmdArgs,
		phHandler:     placeholderHandler,
		outputHandler: outputHandler,
		reasonCheck:   reasonCheck,
		checkReqs:     requirementCheck,
		watch:         NewWatchman(),
	}
}

func (t *targetExecuter) CopyToTarget(target string) *targetExecuter {
	copy := New(
		target,
		t.arguments,
		t.script,
		t.runCfg,
		t.stopReason,
		t.codeLine,
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
	case Script:
		t.script = value.(configure.Task)
	case RunCfg:
		t.runCfg = value.(configure.RunConfig)
	case StopReason:
		t.stopReason = value.(configure.Trigger)
	case CodeLine:
		t.codeLine = value.(string)
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
