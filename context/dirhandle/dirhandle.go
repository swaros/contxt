package dirhandle

import (
	"fmt"
	"log"
	"os"

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
	for index, path := range configure.Config.Paths {
		if number == index {
			fmt.Println(path)
			return
		}
	}
	fmt.Println(".")
}

// GetDir returns the path by index
func GetDir(number int) string {
	for index, path := range configure.Config.Paths {
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
