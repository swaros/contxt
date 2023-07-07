package runner

import (
	"io"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/mimiclog"
)

// mapping the logrus levels to mimiclog levels
type Fields map[string]interface{}

func (f Fields) ToLogrusFields() logrus.Fields {
	return logrus.Fields(f)
}

// the LogrusLogger is a wrapper around the mimiclog.Logger interface
// and the logrus.Logger implementation
// it is mainly taken from the logrus_test.go file in the mimiclog module

type logrusLogger struct {
	*logrus.Logger
}

// then we will create the logrus implementation

func NewLogrusLogger() *logrusLogger {

	return &logrusLogger{
		Logger: logrus.New(),
	}
}

// we need to implement the mimiclog.Logger interface

func (l *logrusLogger) Trace(args ...interface{}) {
	l.Logger.Trace(args...)
}

func (l *logrusLogger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}

func (l *logrusLogger) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

func (l *logrusLogger) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
}

func (l *logrusLogger) Info(args ...interface{}) {
	l.Logger.Info(args...)
}

func (l *logrusLogger) Critical(args ...interface{}) {
	l.Logger.Fatal(args...)
}

func (l *logrusLogger) TraceFn(fn mimiclog.LogFunc, args ...interface{}) {
	if l.IsTraceEnabled() {
		l.Trace(fn(args...))
	}
}

func (l *logrusLogger) DebugFn(fn mimiclog.LogFunc, args ...interface{}) {
	if l.IsDebugEnabled() {
		l.Debug(fn(args...))
	}
}

func (l *logrusLogger) ErrorFn(fn mimiclog.LogFunc, args ...interface{}) {
	if l.IsErrorEnabled() {
		l.Error(fn(args...))
	}
}

func (l *logrusLogger) WarnFn(fn mimiclog.LogFunc, args ...interface{}) {
	if l.IsWarnEnabled() {
		l.Warn(fn(args...))
	}
}

func (l *logrusLogger) InfoFn(fn mimiclog.LogFunc, args ...interface{}) {
	if l.IsInfoEnabled() {
		l.Info(fn(args...))
	}
}

func (l *logrusLogger) CriticalFn(fn mimiclog.LogFunc, args ...interface{}) {
	if l.IsCriticalEnabled() {
		l.Critical(fn(args...))
	}
}

func (l *logrusLogger) IsLevelEnabled(level string) bool {
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
	case mimiclog.LevelCrit:
		return l.IsCriticalEnabled()
	default:
		return false
	}
}

func (l *logrusLogger) IsTraceEnabled() bool {
	return l.Logger.IsLevelEnabled(logrus.TraceLevel)
}

func (l *logrusLogger) IsDebugEnabled() bool {
	return l.Logger.IsLevelEnabled(logrus.DebugLevel)
}

func (l *logrusLogger) IsInfoEnabled() bool {
	return l.Logger.IsLevelEnabled(logrus.InfoLevel)
}

func (l *logrusLogger) IsWarnEnabled() bool {
	return l.Logger.IsLevelEnabled(logrus.WarnLevel)
}

func (l *logrusLogger) IsErrorEnabled() bool {
	return l.Logger.IsLevelEnabled(logrus.ErrorLevel)
}

func (l *logrusLogger) IsCriticalEnabled() bool {
	return l.Logger.IsLevelEnabled(logrus.FatalLevel)
}

func (l *logrusLogger) SetLevelByString(level string) {
	switch level {
	case mimiclog.LevelTrace:
		l.Logger.SetLevel(logrus.TraceLevel)
	case mimiclog.LevelDebug:
		l.Logger.SetLevel(logrus.DebugLevel)
	case mimiclog.LevelInfo:
		l.Logger.SetLevel(logrus.InfoLevel)
	case mimiclog.LevelWarn:
		l.Logger.SetLevel(logrus.WarnLevel)
	case mimiclog.LevelError:
		l.Logger.SetLevel(logrus.ErrorLevel)
	case mimiclog.LevelCrit:
		l.Logger.SetLevel(logrus.FatalLevel)
	default:
		l.Logger.SetLevel(logrus.InfoLevel)
	}
}

func (l *logrusLogger) SetLevel(level interface{}) {
	switch lvl := level.(type) {
	case logrus.Level:
		l.Logger.SetLevel(lvl)
	case string:
		l.SetLevelByString(lvl)
	default:
		l.Logger.SetLevel(logrus.InfoLevel)
	}
}

func (l *logrusLogger) GetLevel() string {
	return l.Logger.GetLevel().String()
}
func (l *logrusLogger) SetOutput(output io.Writer) {
	l.Logger.SetOutput(output)
}

func (l *logrusLogger) GetLogger() interface{} {
	return l.Logger
}

type LogrusMimic struct {
}

func (l *LogrusMimic) Init(args ...interface{}) (mimiclog.Logger, error) {
	return NewLogrusLogger(), nil
}

func (l *LogrusMimic) Name() string {
	return "logrus"
}
