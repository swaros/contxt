package mimiclog_test

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/swaros/contxt/module/mimiclog"
)

// test an logrus implementation of the mimiclog.Logger interface

// first we will create the logrus implementation
type logrusLogger struct {
	*logrus.Logger
}

var testHook *test.Hook

// then we will create the logrus implementation

func NewLogrusLogger() *logrusLogger {
	testLogger, hook := test.NewNullLogger()
	testHook = hook

	return &logrusLogger{
		Logger: testLogger,
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
	case mimiclog.LevelCritical:
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
	case mimiclog.LevelCritical:
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

// this test do not need to be runned, it is just to see
// if the logrus implementation is having something not implemented
func TestLogrusIsMimicInterface(t *testing.T) {
	t.Skip("there is no need to run this test")
	var _ mimiclog.Logger = NewLogrusLogger()
}

func TestBasics(t *testing.T) {
	// we will create a new logger
	mimicLogger := &LogrusMimic{}
	logger, err := mimicLogger.Init()
	if err != nil {
		t.Fatal(err)
	}
	// and set the level to trace
	logger.SetLevel(mimiclog.LevelTrace)

	// now we will log something

	logger.Warn("warn")
	logger.Debug("debug")
	logger.Info("info")
	logger.Trace("trace")

	if testHook.LastEntry().Message != "trace" {
		t.Errorf("expected trace, got %s", testHook.LastEntry().Message)
	}

	logger.SetLevel(mimiclog.LevelInfo)

	logger.Warn("2 warn")
	logger.Debug("2 debug")
	logger.Info("2 info")
	logger.Trace("2 trace")

	if testHook.LastEntry().Message != "2 info" {
		t.Errorf("expected 2 info, got %s", testHook.LastEntry().Message)
	}

	logger.SetLevel(mimiclog.LevelWarn)

	logger.Warn("3 warn")
	logger.Debug("3 debug")
	logger.Info("3 info")
	logger.Trace("3 trace")

	if testHook.LastEntry().Message != "3 warn" {
		t.Errorf("expected 3 warn, got %s", testHook.LastEntry().Message)
	}

}
