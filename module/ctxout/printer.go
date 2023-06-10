package ctxout

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"golang.org/x/term"
)

var (
	// PreHook is a function that can be used to intercept the message before it is printed
	PreHook func(msg ...interface{}) bool = nil
	// output is the interface that will be used to print the message
	output      StreamInterface = nil
	initDone    bool            = false
	postFilters []PostFilter    = []PostFilter{}
	termInfo    PostFilterInfo
	behavior    CtxOutBehavior = CtxOutBehavior{
		NoColored: false,
		ANSI:      true,
		ANSI256:   false,
		ANSI16M:   false,
		Info:      &termInfo,
	}
	runInfos []string // contains informartion what filters are run
	mu       sync.Mutex
)

type CtxOutBehavior struct {
	NoColored bool
	ANSI      bool
	ANSI256   bool
	ANSI16M   bool
	Info      *PostFilterInfo
}

// CtxOutCtrl is a control structure that can be used to control the output
type CtxOutCtrl struct {
	IgnoreCase bool
}

// CtxOutLabel is a control structure that can be used to control the output
type CtxOutLabel struct {
	Message interface{}
	FColor  string
}

// PrintInterface is an interface that can be used to filter the message
type PrintInterface interface {
	Update(info CtxOutBehavior)
	Filter(msg interface{}) interface{}
}

// StreamInterface is an interface that can be used to stream the message
type StreamInterface interface {
	Stream(msg ...interface{})
	StreamLn(msg ...interface{})
}

// PostFilter is an interface that can be used to filter the message after the markup filter
// they works only on strings
type PostFilter interface {
	Update(info PostFilterInfo)
	CanHandleThis(text string) bool
	Command(cmd string) string
}

type PostFilterInfo struct {
	IsTerminal bool // if the output is a terminal
	Colored    bool // if the output can be colored
	Disabled   bool // if the whole filter is enabled. the filter is still called, but it should not change the message. but remove the markup
	Width      int  // the width of the terminal
	Height     int  // the height of the terminal

}

// AddPostFilter adds a post filter to the list of post filters
// these filters are called after the markup filter
// they works only on strings
func AddPostFilter(filter PostFilter) {
	initCtxOut()
	postFilters = append(postFilters, filter)
	filter.Update(termInfo)
}

// GetPostFilters returns the list of post filters
func GetPostFilters() []PostFilter {
	return postFilters
}

// ClearPostFilters clears the list of post filters
func ClearPostFilters() {
	postFilters = []PostFilter{}
}

// initCtxOut initializes the ctxout package
// but only once
func initCtxOut() {
	if initDone {
		return
	}
	fd := int(os.Stdout.Fd())
	isTerm := term.IsTerminal(fd)
	info := PostFilterInfo{
		IsTerminal: isTerm,
		Colored:    isTerm,
		Disabled:   false,
		Width:      80,
		Height:     24,
	}
	if info.IsTerminal {
		w, h, err := term.GetSize(fd)
		if err == nil {
			info.Width = w
			info.Height = h
		}
	}
	initDone = true
	termInfo = info
	behavior.Info = &info
	behavior.NoColored = !info.Colored
	behavior.ANSI = info.Colored

}

func SetBehavior(behave CtxOutBehavior) {
	behavior = behave
}

func GetBehavior() CtxOutBehavior {
	return behavior
}

func IsPrinterInterface(msg interface{}) bool {
	switch msg.(type) {
	case PrintInterface:
		return true
	}
	return false
}

// Message is the function that will be called by the Print and PrintLn functions
// it handles the filtering and streaming of the message depending on the type of the message.
// so here er can also inject the filters and the output stream
func Message(msg ...interface{}) []interface{} {
	mu.Lock()
	defer mu.Unlock()
	runInfos = []string{}
	initCtxOut()
	filters := []PrintInterface{} // these are filters just used in this function

	if PreHook != nil { // if the prehook is defined AND it returns true, we just stop doing anything
		if abort := PreHook(msg...); abort {
			return nil
		}
	}
	var newMsh []interface{}
	for _, chk := range msg {
		switch ctrl := chk.(type) {

		case CtxOutCtrl:
			if chk.(CtxOutCtrl).IgnoreCase { // if we have found this flag set to true, it means ignore the message
				return nil
			}
		case PrintInterface: // we got an interface that can filter the message. so we add it to the list of filters
			filters = append(filters, ctrl)
			runInfos = append(runInfos, fmt.Sprintf("added filter: %T", ctrl))
			ctrl.Update(behavior)
		case StreamInterface: // we got an interface that can stream the message. so we set it as the output
			runInfos = append(runInfos, fmt.Sprintf("set output: %T", ctrl))
			output = ctrl
		default:
			runInfos = append(runInfos, fmt.Sprintf("default hndl message to filterExec: %T", chk))
			newMsh = filterExec(newMsh, filters, chk)
		}

	}
	return PostMarkupFilter(newMsh)
}

// GetRunInfos returns the list of run infos
// while the message is processed, we add infos to this list
func GetRunInfos() []string {
	return runInfos
}

// GetRunInfosF returns the list of run infos
// but only the ones that contains the pattern
func GetRunInfosF(pattern string) []string {
	ret := []string{}
	for _, info := range runInfos {
		if strings.Contains(info, pattern) {
			ret = append(ret, info)
		}
	}
	return ret
}

func PostMarkupFilter(msgSlice []interface{}) []interface{} {
	var newMsh []interface{}
	// we want to summerarize all strings into one string
	// until we find a non string
	// then we call the MarkupFilter function
	// and then we continue with the next string
	stringSum := ""
	for _, msg := range msgSlice {
		if msg != nil {
			switch mg := msg.(type) {
			case string:
				stringSum += mg
			default:
				if stringSum != "" {
					newMsh = append(newMsh, MarkupFilter(stringSum))
					stringSum = ""
				}
				newMsh = append(newMsh, msg)
			}
		}
	}
	// take care if we never found a non string, so we never called the MarkupFilter function
	if stringSum != "" {
		newMsh = append(newMsh, MarkupFilter(stringSum))
	}
	return newMsh
}

// MarkupFilter is the function that will be called by the Message function
// it handles the filtering of the message depending on the type of the message.
func MarkupFilter(msg string) string {

	if len(postFilters) > 0 {
		for _, filter := range postFilters {
			if filter.CanHandleThis(msg) {
				runInfos = append(runInfos, fmt.Sprintf("(YES) post filter: %T", filter))
				msg = filter.Command(msg)
			} else {
				runInfos = append(runInfos, fmt.Sprintf("(NO) post filter: %T - can not handle this", filter))
			}
		}
	} else {
		runInfos = append(runInfos, "no post filters")
	}

	// if we have found this flag set to true, it means ignore the message
	if strings.HasPrefix(msg, "IGNORE") {
		runInfos = append(runInfos, "IGNORE found. message ignored")
		return ""
	}
	return msg
}

// filterExec is the function that will be called by the Message function
// it handles the filters different than the defined post filters
func filterExec(newMsh []interface{}, filters []PrintInterface, msg interface{}) []interface{} {
	initCtxOut()
	if len(filters) > 0 { // we have filters, so they do the job of filtering the message

		for _, filter := range filters {
			runInfos = append(runInfos, fmt.Sprintf("filter exec: %T", filter))
			msg = filter.Filter(msg)
			if msg == nil {
				break
			}
		}

		if msg != nil {
			newMsh = append(newMsh, msg)
		}
	} else {
		newMsh = append(newMsh, msg)
	}
	return newMsh
}

// PrintLn is parsing the message and then printing it
// by using the output interface if is defined
// or by using the fmt.Println function
func PrintLn(msg ...interface{}) {
	msg = Message(msg...)
	if msg != nil {
		if output != nil {
			output.StreamLn(msg...)
		} else {
			fmt.Println(msg...)
		}
	}
}

// Print is parsing the message and then printing it
// by using the output interface if is defined
// or by using the fmt.Print function
func Print(msg ...interface{}) {
	msg = Message(msg...)
	if msg != nil {
		if output != nil {
			output.Stream(msg...)
		} else {
			fmt.Print(msg...)
		}
	}
}

// CtxOut is a shortcut for PrintLn
func CtxOut(msg ...interface{}) {
	PrintLn(msg...)
}

// ToString is parsing the message into a string
func ToString(msg ...interface{}) string {
	msg = Message(msg...)
	var newMsh []string
	for _, chk := range msg {
		switch ctrl := chk.(type) {
		case CtxOutCtrl:
			if chk.(CtxOutCtrl).IgnoreCase { // if we have found this flag set to true, it means ignore the message
				return ""
			}
		case string:
			newMsh = append(newMsh, ctrl)
		default:
			newMsh = append(newMsh, fmt.Sprintf("%v", chk))
		}
	}
	return strings.Join(newMsh, " ")
}
