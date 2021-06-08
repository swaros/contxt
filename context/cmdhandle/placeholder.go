package cmdhandle

import (
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var keyValue sync.Map

// SetPH add key value pair
func SetPH(key, value string) {
	GetLogger().WithField(key, value).Trace("add/overwrite placeholder")
	keyValue.Store(key, value)
}

func SetIfNotExists(key, value string) {
	_, ok := keyValue.Load(key)
	if !ok {
		keyValue.Store(key, value)
	}

}

// GetPH the content og the key
func GetPH(key string) string {
	result, ok := keyValue.Load(key)
	if ok {
		GetLogger().WithField(key, result.(string)).Trace("deliver content from placeholder")
		return result.(string)
	}
	GetLogger().WithField("key", key).Trace("returns empty string because key is not set")
	return ""
}

// HandlePlaceHolder replaces all defined placeholders
func HandlePlaceHolder(line string) string {
	for {
		return handlePlaceHolder(line)
	}
}

func handlePlaceHolder(line string) string {

	if GetLogger().IsLevelEnabled(logrus.TraceLevel) {
		keyValue.Range(func(key, value interface{}) bool {
			keyName := "${" + key.(string) + "}"
			if strings.Contains(line, keyName) {
				GetLogger().WithField("line", line).Trace("replace: source")
				GetLogger().WithField(keyName, value.(string)).Trace("replace: variables")
			}
			line = strings.ReplaceAll(line, keyName, value.(string))
			line = handleMapVars(line)
			return true
		})
	}

	keyValue.Range(func(key, value interface{}) bool {
		keyName := "${" + key.(string) + "}"
		line = strings.ReplaceAll(line, keyName, value.(string))
		line = handleMapVars(line)
		return true
	})
	return line
}

func handleMapVars(line string) string {
	dataKeys := GetDataKeys()
	if len(dataKeys) == 0 {
		return line
	}
	GetLogger().WithField("key-count", len(dataKeys)).Trace("parsing keymap placeholder")
	for _, keyname := range dataKeys {
		lookup := "${" + keyname + ":"
		if strings.Contains(line, lookup) {
			start := strings.Index(line, lookup)
			if start > -1 {
				end := strings.Index(line[start:], "}")
				if end > -1 {
					pathLine := line[start+len(lookup) : start+end]
					if pathLine != "" {
						replace := lookup + pathLine + "}"
						GetLogger().Debug("replace ", replace)
						line = strings.ReplaceAll(line, replace, GetJSONPathValueString(keyname, pathLine))
					}
				} else {
					GetLogger().WithField("key", lookup).Warn("error by getting end position of prefix")
				}
			} else {
				GetLogger().WithField("key", lookup).Warn("error by getting start position of prefix")
			}
		}
	}
	return line
}

// ClearAll removes all entries
func ClearAll() {
	keyValue.Range(func(key, value interface{}) bool {
		keyValue.Delete(key)
		return true
	})
}
