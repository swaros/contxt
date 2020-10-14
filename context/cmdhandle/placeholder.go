package cmdhandle

import (
	"strings"
)

// NewPlaceHolderMap create a new PlaceHolder Map
func NewPlaceHolderMap() map[string]string {
	placeholder := map[string]string{}
	return placeholder
}

// HandlePlaceHolder replaces all defined placeholders
func HandlePlaceHolder(placeholder map[string]string, line string) string {
	for key, value := range placeholder {
		keyName := "${" + key + "}"
		line = strings.ReplaceAll(line, keyName, value)
	}
	return line
}
