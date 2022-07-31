package taskrun

import "os"

// CreateMirror creates nested directories
func CreateMirror(path string) {
	os.MkdirAll(path, os.ModePerm)
}
