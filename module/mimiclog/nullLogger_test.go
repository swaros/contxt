package mimiclog_test

import (
	"testing"

	"github.com/swaros/contxt/module/mimiclog"
)

// create a user of the mimiclog.Logger interface
// and just test the compilation
func TestNullLogger(t *testing.T) {
	logger := mimiclog.NewNullLogger()
	logger.Trace("trace")
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
	logger.Critical("critical")

	if ok := mimiclog.ApplyIfPossible(logger, nil); ok {
		t.Error("ApplyIfPossible should return false")
	}

	if ok := logger.IsCriticalEnabled(); ok {
		t.Error("IsCriticalEnabled should return false")
	}

	if ok := logger.IsErrorEnabled(); ok {
		t.Error("IsErrorEnabled should return false")
	}

	if ok := logger.IsWarnEnabled(); ok {
		t.Error("IsWarnEnabled should return false")
	}

	if ok := logger.IsInfoEnabled(); ok {
		t.Error("IsInfoEnabled should return false")
	}

	if ok := logger.IsDebugEnabled(); ok {
		t.Error("IsDebugEnabled should return false")
	}

	if ok := logger.IsTraceEnabled(); ok {
		t.Error("IsTraceEnabled should return false")
	}

	if ok := logger.IsLevelEnabled(mimiclog.LevelTrace); ok {
		t.Error("IsLevelEnabled should return false")
	}

	if ok := logger.IsLevelEnabled(mimiclog.LevelDebug); ok {
		t.Error("IsLevelEnabled should return false")
	}

	if ok := logger.IsLevelEnabled(mimiclog.LevelInfo); ok {
		t.Error("IsLevelEnabled should return false")
	}

	if ok := logger.IsLevelEnabled(mimiclog.LevelWarn); ok {
		t.Error("IsLevelEnabled should return false")
	}

	if ok := logger.IsLevelEnabled(mimiclog.LevelError); ok {
		t.Error("IsLevelEnabled should return false")
	}

	if ok := logger.IsLevelEnabled(mimiclog.LevelCritical); ok {
		t.Error("IsLevelEnabled should return false")
	}

	if ok := logger.IsLevelEnabled("unknown"); ok {
		t.Error("IsLevelEnabled should return false")
	}

	if ok := logger.IsLevelEnabled(""); ok {
		t.Error("IsLevelEnabled should return false")
	}

	getLg := logger.GetLogger()
	if getLg != nil {
		t.Error("GetLogger should return nil")
	}

	if level := logger.GetLevel(); level != mimiclog.LevelNone {
		t.Error("GetLevel should return LevelNone")
	}
}
