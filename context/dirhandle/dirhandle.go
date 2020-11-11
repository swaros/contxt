package dirhandle

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/swaros/contxt/context/configure"
)

// Current returns the current path
func Current() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir, err
}

// PrintDir prints the all the paths
func PrintDir(number int) {
	for index, path := range configure.UsedConfig.Paths {
		if number == index {
			fmt.Println(path)
			return
		}
	}
	fmt.Println(".")
}

// GetDir returns the path by index
func GetDir(number int) string {
	for index, path := range configure.UsedConfig.Paths {
		if number == index {

			return path
		}
	}
	return "."
}

// Exists checks if a path exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// FileTypeHandler calls function depending on file ending
// and if this fil exists
func FileTypeHandler(path string, jsonHandle func(string), yamlHandle func(string), notExists func(string, error)) {
	fileInfo, err := os.Stat(path)
	if err == nil && !fileInfo.IsDir() {
		var extension = filepath.Ext(path)
		var basename = filepath.Base(path)
		switch extension {
		case ".yaml", ".yml":
			yamlHandle(basename)
		case ".json":
			jsonHandle(basename)
		}
	} else {
		notExists(path, err)
	}
}
