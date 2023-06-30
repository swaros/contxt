package yaclint

import (
	"fmt"
	"strings"
	"time"

	"github.com/swaros/contxt/module/ctxout"
)

// lets create a simple and dirty logger for the linting process
// right now i don't like the idea to use a logger for this.
// this will (maybe) be changed in the future
// but for now i did not find anything matching my needs.
// even the logger interfaces i saw, have a different theorie of logging then
// i have.
// the logger should track all messages and print them at the end of the linting process.
// but there you need to know the context of the message.
// for this you don't need different levels of logging, like Error or info. and you dont't need to push
// the messages to a file or something else.
// this type of logging is for development only. while you are writing code, and not while you try to
// figure out, why the app crashes in a pod running in the cloud.
// you need to know anything what depends on the entry "key" and the context.
// and you don't casre about the level. and you dont need any other information, what dosn't help you to find the issue.

type DirtyLoggerDef interface {
	Trace(args ...interface{})
	SetTraceFn(traceFn TraceFn)
	GetTrace(find ...string) []string
	Print(patter ...string) string
}

type TraceFn func(args ...interface{}) // thats all we need right now
const (
	DirtyLogMaxLen = 500
)

type DirtyLogger struct {
	traceFn  TraceFn
	logStuff []string
	addTime  bool
}

func NewDirtyLogger() *DirtyLogger {
	return &DirtyLogger{
		traceFn: nil,
		addTime: true,
	}
}

func (dl *DirtyLogger) SetAddTime(addTime bool) *DirtyLogger {
	dl.addTime = addTime
	return dl
}

func (dl *DirtyLogger) Trace(args ...interface{}) {
	if dl.traceFn != nil {
		dl.traceFn(args...)
	} else {
		dl.CreateSimpleTracer()
		dl.traceFn("CreateSimpleTracer: there was no traceFn set. so we create a simple one.")
		dl.traceFn(args...)
	}
}

func (dl *DirtyLogger) GetTrace(find ...string) []string {
	if len(find) == 0 {
		return dl.logStuff
	}
	// lets find the strings
	var ret []string
	for _, s := range dl.logStuff {
		for _, f := range find {
			if strings.Contains(strings.ToLower(s), strings.ToLower(f)) {
				ret = append(ret, s)
			}
		}
	}
	return ret
}

func (dl *DirtyLogger) Print(pattern ...string) string {
	return strings.Join(dl.GetTrace(pattern...), "\n")
}

func (dl *DirtyLogger) SetTraceFn(traceFn TraceFn) {
	dl.traceFn = traceFn
}

func (dl *DirtyLogger) CreateSimpleTracer() *DirtyLogger {
	dl.traceFn = func(args ...interface{}) {
		addStr := ""
		// add the current time first
		if dl.addTime {
			addStr += time.Now().Format("[15:04:05.000] - ")
		}

		for _, a := range args {
			switch a := a.(type) {
			case string:
				addStr += a
			case MatchToken:
			case *MatchToken:
				addorRm := "-"
				if a.Added {
					addorRm = "+"
				}
				addStr += "[" + addorRm + a.keyToString() + "] "
			case []string: // better readable instead of the fmt.Sprint
				addStr += "["
				for _, s := range a {
					addStr += "'" + s + "',"
				}
				addStr += "]"
			default:
				addStr += fmt.Sprintf("%v", a)
			}
		}
		if addStr != "" {
			// lets add the string to the logStuff
			// but respect the max length
			if len(dl.logStuff)+1 == DirtyLogMaxLen {
				// remove the first element
				dl.logStuff = dl.logStuff[1:]
			}
			// append the new one
			dl.logStuff = append(dl.logStuff, addStr)
		}
	}
	return dl
}

func (dl *DirtyLogger) CreateCtxoutTracer() *DirtyLogger {
	dl.traceFn = func(args ...interface{}) {
		addStr := ""
		// add the current time first
		if dl.addTime {
			addStr += ctxout.ForeDarkGrey + ctxout.BackWhite + time.Now().Format("[15:04:05.000]") + ctxout.CleanTag + " - "
		}

		for _, a := range args {
			switch a := a.(type) {
			case bool:
				if a {
					addStr += ctxout.ForeGreen + ctxout.BackBlack + "true" + ctxout.CleanTag
				} else {
					addStr += ctxout.ForeRed + ctxout.BackBlack + "false" + ctxout.CleanTag
				}
			case string:
				addStr += ctxout.ForeLightBlue + a + ctxout.CleanTag

			case MatchToken:
			case *MatchToken:
				addorRm := ctxout.ForeRed + "-"
				if a.Added {
					addorRm = ctxout.ForeGreen + "+"
				}
				addStr += ctxout.ForeCyan + "[" + addorRm + a.keyToString() + ctxout.CleanTag + ctxout.ForeCyan + "] " + ctxout.CleanTag
				addStr += ctxout.ForeDarkGrey + fmt.Sprintf("%v", a.Value) + ctxout.CleanTag
			case []string: // better readable instead of the fmt.Sprint
				addStr += ctxout.ForeMagenta + ctxout.BackBlack + "["
				joinStr := []string{}
				for _, s := range a {
					joinStr = append(joinStr, ctxout.ForeMagenta+"'"+ctxout.ForeCyan+s+ctxout.ForeMagenta+"'")
				}
				addStr += strings.Join(joinStr, ctxout.ForeMagenta+",") + "]" + ctxout.CleanTag
			default:
				addStr += fmt.Sprintf("%v", a)
			}
		}
		if addStr != "" {
			// lets add the string to the logStuff
			// but respect the max length
			if len(dl.logStuff)+1 == DirtyLogMaxLen {
				// remove the first element
				dl.logStuff = dl.logStuff[1:]
			}
			// append the new one
			dl.logStuff = append(dl.logStuff, ctxout.ToString(ctxout.NewMOWrap(), addStr))
		}
	}
	return dl
}
