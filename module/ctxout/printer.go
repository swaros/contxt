package ctxout

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// PreHook is a function that can be used to intercept the message before it is printed
var PreHook func(msg ...interface{}) bool = nil

var (
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

func AddPostFilter(filter PostFilter) {
	initCtxOut()
	postFilters = append(postFilters, filter)
	filter.Update(termInfo)
}

func initCtxOut() {
	if initDone {
		return
	}
	fd := int(os.Stdout.Fd())
	info := PostFilterInfo{
		IsTerminal: term.IsTerminal(fd),
		Colored:    term.IsTerminal(fd),
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

	termInfo = info
}

func SetBehavior(behave CtxOutBehavior) {
	behavior = behave
}

func GetBehavior() CtxOutBehavior {
	return behavior
}

func InitTerminal() {
	initCtxOut()
	behavior.Info.IsTerminal = true
}

// Message is the function that will be called by the Print and PrintLn functions
func Message(msg ...interface{}) []interface{} {
	initCtxOut()
	filters := []PrintInterface{}

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
			ctrl.Update(behavior)
		case StreamInterface: // we got an interface that can stream the message. so we set it as the output
			output = ctrl
		default:
			newMsh = filterExec(newMsh, filters, chk)
		}

	}
	return PostMarkupFilter(newMsh)
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

func MarkupFilter(msg string) string {

	if len(postFilters) > 0 {
		for _, filter := range postFilters {
			if filter.CanHandleThis(msg) {
				return filter.Command(msg)
			}
		}
	}

	// if we have found this flag set to true, it means ignore the message
	if strings.HasPrefix(msg, "IGNORE") {
		return ""
	}
	return msg
}

func filterExec(newMsh []interface{}, filters []PrintInterface, msg interface{}) []interface{} {
	initCtxOut()
	if len(filters) > 0 { // we have filters, so they do the job of filtering the message
		for _, filter := range filters {
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
