// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

package runner

import (
	"fmt"
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

func (l *logrusLogger) makeNice(level string, args ...interface{}) {

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

func (l *logrusLogger) logWithLevel(level string, msg string, args ...interface{}) {

	entry := logrus.NewEntry(l.Logger)
	if len(args) > 0 {
		switch argtype := args[0].(type) {
		case map[string]interface{}:
			entry = l.Logger.WithFields(argtype)
		case Fields:
			entry = l.Logger.WithFields(argtype.ToLogrusFields())
		case mimiclog.Fields:
			entry = l.Logger.WithFields(logrus.Fields(argtype))
		default:
			entry = l.Logger.WithField("data", args)
		}
	}

	switch level {
	case mimiclog.LevelTrace:
		entry.Trace(msg)
	case mimiclog.LevelDebug:
		entry.Debug(msg)
	case mimiclog.LevelInfo:
		entry.Info(msg)
	case mimiclog.LevelWarn:
		entry.Warn(msg)
	case mimiclog.LevelError:
		entry.Error(msg)
	case mimiclog.LevelCrit:
		entry.Fatal(msg)
	default:
		entry.Trace(msg)
	}

}
func (l *logrusLogger) Trace(args ...interface{}) {
	l.makeNice(mimiclog.LevelTrace, args...)
}

func (l *logrusLogger) Debug(args ...interface{}) {
	l.makeNice(mimiclog.LevelDebug, args...)
}

func (l *logrusLogger) Error(args ...interface{}) {
	l.makeNice(mimiclog.LevelError, args...)
}

func (l *logrusLogger) Warn(args ...interface{}) {
	l.makeNice(mimiclog.LevelWarn, args...)
}

func (l *logrusLogger) Info(args ...interface{}) {
	l.makeNice(mimiclog.LevelInfo, args...)
}

func (l *logrusLogger) Critical(args ...interface{}) {
	l.makeNice(mimiclog.LevelCrit, args...)
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
