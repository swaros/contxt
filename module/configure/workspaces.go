package configure

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ListWorkSpaces : list all existing workspaces
func ListWorkSpaces() []string {
	var files []string
	var fullHomeDir string
	homeDir, err := getUserDir()
	if err == nil {
		fullHomeDir = homeDir + DefaultPath
		err := filepath.Walk(fullHomeDir, func(path string, info os.FileInfo, err error) error {
			files = append(files, path)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
	return files
}

// DisplayWorkSpaces prints out all workspaces
func DisplayWorkSpaces() {
	//var files []string
	files := ListWorkSpaces()

	if len(files) > 0 {
		for _, file := range files {
			var basePath = filepath.Base(file)
			var extension = filepath.Ext(file)
			// display json files only they are not the default config
			if extension == ".json" && basePath != DefaultConfigFileName {
				basePath = strings.TrimSuffix(basePath, extension)
				// we are also not interested in the default workspace
				if basePath != DefaultWorkspace {
					fmt.Println(basePath)
				}
			}
		}
	}
}

// GetWorkSpacesAsList prints out all workspaces
func GetWorkSpacesAsList() ([]string, bool) {
	var files []string
	var workspaces []string
	found := false
	files = ListWorkSpaces()

	if len(files) > 0 {
		for _, file := range files {
			var basePath = filepath.Base(file)
			var extension = filepath.Ext(file)
			// display json files only they are not the default config
			if extension == ".json" && basePath != DefaultConfigFileName {
				basePath = strings.TrimSuffix(basePath, extension)
				// we are also not interested in the default workspace
				if basePath != DefaultWorkspace {
					workspaces = append(workspaces, basePath)
					found = true
				}
			}
		}
	}
	return workspaces, found
}

// WorkSpaces handler to iterate all workspaces
func WorkSpaces(callback func(string)) {
	//var files []string
	files := ListWorkSpaces()

	if len(files) > 0 {
		for _, file := range files {
			var basePath = filepath.Base(file)
			var extension = filepath.Ext(file)
			// display json files only they are not the default config
			if extension == ".json" && basePath != DefaultConfigFileName {
				basePath = strings.TrimSuffix(basePath, extension)
				// we are also not interested in the default workspace
				if basePath != DefaultWorkspace {
					callback(basePath)
				}
			}
		}
	}
}
