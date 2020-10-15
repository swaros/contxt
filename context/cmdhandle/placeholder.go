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
		return true
	})

	return line
}
