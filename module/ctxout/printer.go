package ctxout

import (
	"fmt"
	"strings"
)

// PreHook is a function that can be used to intercept the message before it is printed
var PreHook func(msg ...interface{}) bool = nil

var (
	// output is the interface that will be used to print the message
	output   StreamInterface = nil
	initDone bool            = false
)

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
	Filter(msg interface{}) interface{}
}

// StreamInterface is an interface that can be used to stream the message
type StreamInterface interface {
	Stream(msg ...interface{})
	StreamLn(msg ...interface{})
}

func initCtxOut() {
	if !initDone {
		initDone = true
	}
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
		case StreamInterface: // we got an interface that can stream the message. so we set it as the output
			output = ctrl
		default:
			newMsh = filterExec(newMsh, filters, chk)
		}

	}
	return newMsh
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
			fmt.Println(msg)
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
			fmt.Print(msg)
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
