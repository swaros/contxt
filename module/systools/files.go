package systools

import (
	"os"
	"path/filepath"
)

func pathCompare(left, right string) bool {
	l := filepath.FromSlash(left)
	r := filepath.FromSlash(right)

	return l == r
}

func CopyFile(source, target string) error {
	r, err := os.Open(filepath.Clean(source))
	if err != nil {
		return err
	}
	defer r.Close()
	w, err := os.Create(filepath.Clean(target))
	if err != nil {
		return err
	}
	defer w.Close()
	if _, err := w.ReadFrom(r); err != nil {
		return err
	}
	return nil
}
