package cmdhandle

import (
	"strings"
	"sync"
)

var keyValue sync.Map

// SetPH add key value pair
func SetPH(key, value string) {
	keyValue.Store(key, value)
}

// GetPH the content og the key
func GetPH(key string) string {
	result, ok := keyValue.Load(key)
	if ok {
		return result.(string)
	}
	return ""
}

// HandlePlaceHolder replaces all defined placeholders
func HandlePlaceHolder(line string) string {
	for {
		return handlePlaceHolder(line)
	}
}

func handlePlaceHolder(line string) string {

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
				}
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
