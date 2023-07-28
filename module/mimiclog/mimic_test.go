package mimiclog_test

import (
	"testing"

	"github.com/swaros/contxt/module/mimiclog"
)

// create a user of the mimiclog.Logger interface
type logrusLoggerUsage struct {
	logger mimiclog.Logger
}

func (lu *logrusLoggerUsage) SetLogger(logger mimiclog.Logger) {
	lu.logger = logger
}

// an non mimic.Logger compatible implementation

type nonCompatibleLogger struct {
}

// this test do not need to run, it just tests the compilation
func TestLogrusUser(t *testing.T) {
	logger := NewLogrusLogger()

	lu := &logrusLoggerUsage{}
	mimiclog.ApplyLogger(logger, lu)
}

func TestLogrusTry(t *testing.T) {
	nonComp := &nonCompatibleLogger{}
	logger := NewLogrusLogger()

	if ok := mimiclog.ApplyIfPossible(logger, nonComp); ok {
		t.Error("ApplyIfPossible should return false")
	}

	lu := &logrusLoggerUsage{}
	if ok := mimiclog.ApplyIfPossible(logger, lu); !ok {
		t.Error("ApplyIfPossible should return true")
	}
}
