package systools

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/swaros/contxt/module/dirhandle"
)

func PathCompare(left, right string) bool {
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

// WriteFileIfNotExists writes a file if it does not exist
// reports 0 if file was written, 1 if file exists, 2 if error
// on error, error is returned
func WriteFileIfNotExists(filename, content string) (int, error) {
	funcExists, funcErr := dirhandle.Exists(filename)
	if funcErr == nil && !funcExists {
		os.WriteFile(filename, []byte(content), 0644)
		return 0, nil
	} else if funcExists {
		return 1, nil
	}
	return 2, funcErr

}

// Exists reports whether the named file or directory exists.
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

// updateExistingFileIfNotContains updates a file if it does not contain a string
// this is made to avoid multiple updates of the same file
func updateExistingFileIfNotContains(filename, content, doNotContain string) (bool, error) {
	ok, errDh := Exists(filename)
	errmsg := ""
	if errDh == nil && ok {
		byteCnt, err := os.ReadFile(filename)
		if err != nil {
			return false, errors.New("file not readable " + filename + " " + err.Error())
		}
		strContent := string(byteCnt)
		if strings.Contains(strContent, doNotContain) {
			return false, errors.New("it seems file is already updated. it contains: " + doNotContain)
		} else {
			file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {

				return false, errors.New("error while opening file " + filename)
			}
			defer file.Close()
			if _, err := file.WriteString(content); err != nil {

				return false, errors.New("error adding content to file " + filename)
			}
			return true, nil
		}

	} else {
		errmsg = "file update error: file not exists " + filename
	}
	return false, errors.New(errmsg)
}
