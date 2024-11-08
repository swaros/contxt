package tasks_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/mimiclog"
)

// mapping the logrus levels to mimiclog levels
type Fields map[string]interface{}

// the LogrusLogger is a wrapper around the mimiclog.Logger interface
// and the logrus.Logger implementation
// it is mainly taken from the logrus_test.go file in the mimiclog module

type testLogger struct {
	allLevelsEnabled bool
	traces           []string
	debugs           []string
	infos            []string
	warns            []string
	errors           []string
	criticals        []string
	testing          *testing.T
	logLevel         string
}

// then we will create the logrus implementation

func NewTestLogger(test *testing.T) *testLogger {

	return &testLogger{
		testing:          test,
		allLevelsEnabled: true,
	}
}

// we need to implement the mimiclog.Logger interface

func (l *testLogger) makeNice(level string, args ...interface{}) {

	msg := ""
	firstIsString := false
	if len(args) == 0 {
		return
	}
	switch val := args[0].(type) {
	case string:
		firstIsString = true
		msg = val
	case error:
		msg = val.Error()
	default:
		msg = fmt.Sprintf("%v", val)
	}

	if firstIsString {
		if len(args) > 1 {
			l.logWithLevel(level, msg, args[1:]...)
		} else {
			l.logWithLevel(level, msg)
		}
	} else {
		l.logWithLevel(level, msg, args...)
	}
}

func (l *testLogger) logWithLevel(level string, msg string, args ...interface{}) {

	// GET ARGS AS STRING
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	switch level {
	case mimiclog.LevelTrace:
		l.traces = append(l.traces, msg)
	case mimiclog.LevelDebug:
		l.debugs = append(l.debugs, msg)
	case mimiclog.LevelInfo:
		l.infos = append(l.infos, msg)
	case mimiclog.LevelWarn:
		l.warns = append(l.warns, msg)
	case mimiclog.LevelError:
		l.errors = append(l.errors, msg)
	case mimiclog.LevelCritical:
		l.criticals = append(l.criticals, msg)
	default:
		l.infos = append(l.infos, msg)
	}

}
func (l *testLogger) Trace(args ...interface{}) {
	l.makeNice(mimiclog.LevelTrace, args...)
}

func (l *testLogger) Debug(args ...interface{}) {
	l.makeNice(mimiclog.LevelDebug, args...)
}

func (l *testLogger) Error(args ...interface{}) {
	l.makeNice(mimiclog.LevelError, args...)
}

func (l *testLogger) Warn(args ...interface{}) {
	l.makeNice(mimiclog.LevelWarn, args...)
}

func (l *testLogger) Info(args ...interface{}) {
	l.makeNice(mimiclog.LevelInfo, args...)
}

func (l *testLogger) Critical(args ...interface{}) {
	l.makeNice(mimiclog.LevelCritical, args...)
}

func (l *testLogger) IsLevelEnabled(level string) bool {
	if l.allLevelsEnabled {
		return true
	}
	switch level {
	case mimiclog.LevelTrace:
		return l.IsTraceEnabled()
	case mimiclog.LevelDebug:
		return l.IsDebugEnabled()
	case mimiclog.LevelInfo:
		return l.IsInfoEnabled()
	case mimiclog.LevelWarn:
		return l.IsWarnEnabled()
	case mimiclog.LevelError:
		return l.IsErrorEnabled()
	case mimiclog.LevelCritical:
		return l.IsCriticalEnabled()
	default:
		return false
	}
}

func (l *testLogger) IsTraceEnabled() bool {
	return l.logLevel == mimiclog.LevelTrace || l.allLevelsEnabled
}

func (l *testLogger) IsDebugEnabled() bool {
	return l.logLevel >= mimiclog.LevelDebug || l.allLevelsEnabled
}

func (l *testLogger) IsInfoEnabled() bool {
	return l.logLevel >= mimiclog.LevelInfo || l.allLevelsEnabled
}

func (l *testLogger) IsWarnEnabled() bool {
	return l.logLevel >= mimiclog.LevelWarn || l.allLevelsEnabled
}

func (l *testLogger) IsErrorEnabled() bool {
	return l.logLevel >= mimiclog.LevelError || l.allLevelsEnabled
}

func (l *testLogger) IsCriticalEnabled() bool {
	return l.logLevel >= mimiclog.LevelCritical || l.allLevelsEnabled
}

func (l *testLogger) SetLevelByString(level string) {
	switch level {
	case mimiclog.LevelTrace:
		l.logLevel = mimiclog.LevelTrace
	case mimiclog.LevelDebug:
		l.logLevel = mimiclog.LevelDebug
	case mimiclog.LevelInfo:
		l.logLevel = mimiclog.LevelInfo
	case mimiclog.LevelWarn:
		l.logLevel = mimiclog.LevelWarn
	case mimiclog.LevelError:
		l.logLevel = mimiclog.LevelError
	case mimiclog.LevelCritical:
		l.logLevel = mimiclog.LevelCritical
	default:
		l.logLevel = mimiclog.LevelInfo
	}
}

func (l *testLogger) SetLevel(level interface{}) {
	switch lvl := level.(type) {
	case logrus.Level:
		l.SetLevel(lvl)
	case string:
		l.SetLevelByString(lvl)
	default:
		l.SetLevel(logrus.InfoLevel)
	}
}

func (l *testLogger) GetLevel() string {
	return l.logLevel
}
func (l *testLogger) SetOutput(output io.Writer) {
	// do nothing
}

func (l *testLogger) GetLogger() interface{} {
	return l.logLevel
}

type LogrusMimic struct {
}

func (l *LogrusMimic) Init(args ...interface{}) (mimiclog.Logger, error) {
	return NewTestLogger(args[0].(*testing.T)), nil
}

func (l *LogrusMimic) Name() string {
	return "TestLogger"
}
