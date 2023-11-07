package process_test

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/swaros/contxt/module/process"
)

func bashPs(cmd string) ([]string, error) {
	cmdString := fmt.Sprintf(`ps -eo cmd | grep "%s"`, cmd)
	process := process.NewProcess("bash", "-c", cmdString)
	outputs := []string{}
	process.SetOnOutput(func(msg string, err error) bool {
		outputs = append(outputs, msg)
		return true
	})
	if _, _, err := process.Exec(); err != nil {
		return []string{}, err
	}
	return outputs, nil
}

func CmdIsRunning(t *testing.T, cmd string) bool {
	t.Helper()
	if runtime.GOOS == "windows" {
		return true // TODO: implement this
	}
	output, err := bashPs(cmd)
	if err != nil {
		t.Error(err)
		return false
	}
	for _, line := range output {
		// we just need to check if we hit the grep command
		// if we hit something without the grep command
		// it is the process we are looking for
		if !strings.Contains(line, "grep") {
			return true
		}
	}
	return false
}

// helper function to skip tests on github ci.
// we need to skip tests that require some special setup.
// be careful to not skip tests that are important. double check if these tests are working elsewhere.
// and make sure the issue is not related to the github ci setup.
func SkipOnGithubCi(t *testing.T) {
	t.Helper()
	// check if one of the github ci env vars is set
	if os.Getenv("GITHUB_CI") != "" || os.Getenv("GITHUB_ENV") != "" || os.Getenv("GITHUB_WORKSPACE") != "" {
		t.Skip("Skip on Github CI")
	}
	t.Log("Not on Github CI. continue testing...")
}

// mimiclogger implementation that collects all the logs in memory
// we need to implement the MimiLogger interface

type MimicLogEntry struct {
	Level string
	Msg   string
}

const (
	TraceLevel = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	CriticalLevel
)

type MimicLogger struct {
	logs    []MimicLogEntry
	syncLoc sync.Mutex
	level   int
}

func NewMimicTestLogger() *MimicLogger {
	return &MimicLogger{
		logs: []MimicLogEntry{},
	}
}

func (m *MimicLogger) GetLogs() []string {
	logs := []string{}
	m.syncLoc.Lock()
	defer m.syncLoc.Unlock()
	for _, entry := range m.logs {
		logs = append(logs, entry.Level+"\t"+entry.Msg)
	}
	return logs
}

func (m *MimicLogger) LogsToTestLog(t *testing.T) []string {
	t.Helper()

	logs := []string{}
	for _, entry := range m.logs {
		logs = append(logs, t.Name()+":"+entry.Level+"\t"+entry.Msg)
		t.Log("[" + t.Name() + ":" + entry.Level + "]\t" + entry.Msg)
	}
	return logs
}

func (m *MimicLogger) ResetLogs() {
	m.logs = []MimicLogEntry{}
}

func (m *MimicLogger) GetLevelInt() int {
	return m.level
}

func (m *MimicLogger) SetLevelInt(level int) {
	m.level = level
}

func (m *MimicLogger) GetLevelString() string {
	switch m.level {
	case TraceLevel:
		return "TRACE"
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case CriticalLevel:
		return "CRITICAL"
	default:
		return ""
	}
}

func (m *MimicLogger) GetLevel() string {
	return m.GetLevelString()
}

func (m *MimicLogger) SetLevelString(level string) {
	switch strings.ToUpper(level) {
	case "TRACE":
		m.level = TraceLevel
	case "DEBUG":
		m.level = DebugLevel
	case "INFO":
		m.level = InfoLevel
	case "WARN":
		m.level = WarnLevel
	case "ERROR":
		m.level = ErrorLevel
	case "CRITICAL":
		m.level = CriticalLevel
	default:
		m.level = InfoLevel
	}
}

func (m *MimicLogger) addToLogs(level string, msg string) {
	if !m.IsAllowedLevelString(level) {
		return
	}
	m.syncLoc.Lock()
	defer m.syncLoc.Unlock()
	entry := MimicLogEntry{
		Level: level,
		Msg:   msg,
	}
	m.logs = append(m.logs, entry)
}

func (m *MimicLogger) Trace(args ...interface{}) {
	m.addToLogs("TRACE", fmt.Sprint(args...))
}

func (m *MimicLogger) Debug(args ...interface{}) {
	m.addToLogs("DEBUG", fmt.Sprint(args...))
}

func (m *MimicLogger) Info(args ...interface{}) {
	m.addToLogs("INFO", fmt.Sprint(args...))
}

func (m *MimicLogger) Error(args ...interface{}) {
	m.addToLogs("ERROR", fmt.Sprint(args...))
}

func (m *MimicLogger) Warn(args ...interface{}) {
	m.addToLogs("WARN", fmt.Sprint(args...))
}

func (m *MimicLogger) Critical(args ...interface{}) {
	m.addToLogs("CRITICAL", fmt.Sprint(args...))
}

func (m *MimicLogger) IsLevelEnabled(level string) bool {
	return m.IsAllowedLevelString(level)
}

func (m *MimicLogger) IsTraceEnabled() bool {
	return m.IsAllowedLevelInt(TraceLevel)
}

func (m *MimicLogger) IsDebugEnabled() bool {
	return m.IsAllowedLevelInt(DebugLevel)
}

func (m *MimicLogger) IsInfoEnabled() bool {
	return m.IsAllowedLevelInt(InfoLevel)
}

func (m *MimicLogger) IsWarnEnabled() bool {
	return m.IsAllowedLevelInt(WarnLevel)
}

func (m *MimicLogger) IsErrorEnabled() bool {
	return m.IsAllowedLevelInt(ErrorLevel)
}

func (m *MimicLogger) IsCriticalEnabled() bool {
	return m.IsAllowedLevelInt(CriticalLevel)
}

func (m *MimicLogger) IsAllowedLevel(level interface{}) bool {
	switch level := level.(type) {
	case int:
		return m.IsAllowedLevelInt(level)
	case string:
		return m.IsAllowedLevelString(level)
	default:
		return false
	}

}

func (m *MimicLogger) IsAllowedLevelInt(level int) bool {
	return level >= m.level
}

func (m *MimicLogger) IsAllowedLevelString(level string) bool {
	switch strings.ToUpper(level) {
	case "TRACE":
		return m.IsAllowedLevelInt(TraceLevel)
	case "DEBUG":
		return m.IsAllowedLevelInt(DebugLevel)
	case "INFO":
		return m.IsAllowedLevelInt(InfoLevel)
	case "WARN":
		return m.IsAllowedLevelInt(WarnLevel)
	case "ERROR":
		return m.IsAllowedLevelInt(ErrorLevel)
	case "CRITICAL":
		return m.IsAllowedLevelInt(CriticalLevel)
	default:
		return false
	}
}

func (m *MimicLogger) SetLevelByString(level string) {
}

func (m *MimicLogger) SetLevel(level interface{}) {
	switch level := level.(type) {
	case int:
		m.SetLevelInt(level)
	case string:
		m.SetLevelString(level)
	}
}

func (m *MimicLogger) GetLogger() interface{} {
	return nil
}
