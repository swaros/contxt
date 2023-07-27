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

package mimiclog

// mimiclog is a module that ust defines a logger interface and a default logger implementation.
// it is ment to be used as a dependency for other modules that need logging, without sticking to
// a specific logging framework.
// so the idea is to use this module as a dependency and then use the logger interface to log.
// the default logger implementation is a simple wrapper around the standard log package.

const (
	LevelTrace = "trace"
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelCrit  = "critical"
	LevelNone  = "none"
)

type LogFunc func(args ...interface{}) []interface{}

type Logger interface {
	// the logging methods depending on the log level
	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Error(args ...interface{})
	Warn(args ...interface{})
	Critical(args ...interface{})

	// maintain functions
	IsLevelEnabled(level string) bool // returns true if the given level is enabled
	IsTraceEnabled() bool             // returns true if trace is enabled
	IsDebugEnabled() bool             // returns true if debug is enabled
	IsInfoEnabled() bool              // returns true if info is enabled
	IsWarnEnabled() bool              // returns true if warn is enabled
	IsErrorEnabled() bool             // returns true if error is enabled
	IsCriticalEnabled() bool          // returns true if critical is enabled

	// the level functions
	SetLevelByString(level string) // sets the log level by a string like "trace", "debug", "info", "warn", "error", "critical"
	SetLevel(level interface{})    // set the log level by 'something' that is depending the logger implementation
	GetLevel() string              // returns the current log level as string

	GetLogger() interface{} // returns the logger instance
}

type MimicLogger interface {
	// Init initializes the logger.
	// Returns the logger instance and an error if something went wrong.
	Init(args ...interface{}) (*Logger, error)
	Name() string
}

// mapppings for the logrus logger fields
type Fields map[string]interface{}

// If a module wants to use the mimiclog module, it has to implement this interface.
// So it can be used to inject the logger into the module, without changing any other interfaces
// they my be implemented by the module.
type MimicLogUser interface {
	SetLogger(logger Logger)
}

// just a helper function to apply a logger to a module that implements the MimicLogUser interface
func ApplyLogger(logger Logger, logUser MimicLogUser) {
	logUser.SetLogger(logger)
}

// just a helper function to apply a logger to a module that implements the MimicLogUser interface
// without knowing if the logUser implements the MimicLogUser interface
// returns true if the logger was applied, false if not
func ApplyIfPossible(logger Logger, logUser interface{}) bool {
	if logUser, ok := logUser.(MimicLogUser); ok {
		logUser.SetLogger(logger)
		return true
	}
	return false
}
