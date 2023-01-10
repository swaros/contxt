package tasks

import "github.com/tidwall/gjson"

// DataMapHandler is the interface for the data handlers
// they are used to store and retrieve data
type DataMapHandler interface {
	GetJSONPathResult(key, path string) (gjson.Result, bool) // returns the result of a json path query
	GetDataAsJson(key string) (string, bool)                 // returns the data as json string
	GetDataAsYaml(key string) (string, bool)                 // returns the data as yaml string
	AddJSON(key, jsonString string) error                    // adds data by parsing a json string
	SetJSONValueByPath(key, path, value string) error        // sets a value by a json path using
}

// PlaceHolder is the interface for the placeholder handler
// they are used to store and retrieve placeholder
type PlaceHolder interface {
	SetPH(key, value string)                                                    // sets a placeholder and its value
	AppendToPH(key, value string) bool                                          // appends a value to a placeholder
	SetIfNotExists(key, value string)                                           // sets a placeholder and its value if it does not exist
	GetPHExists(key string) (string, bool)                                      // returns the value of a placeholder if it exists
	GetPH(key string) string                                                    // returns the value of a placeholder if it exists, otherwise it returns the placeholder itself
	GetPlaceHoldersFnc(inspectFunc func(phKey string, phValue string))          // iterates over all placeholders and calls the inspectFunc for each placeholder
	HandlePlaceHolder(line string) string                                       // replaces all placeholders in a string
	HandlePlaceHolderWithScope(line string, scopeVars map[string]string) string // replaces all placeholders in a string with a scope
	ClearAll()                                                                  // clears all placeholders
	ExportVarToFile(variable string, filename string) error                     // exports a placeholder to a file
}
