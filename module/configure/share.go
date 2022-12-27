package configure

import (
	"os"
	"path/filepath"

	"github.com/swaros/contxt/module/systools"
)

// GetSharedPath returns the full path to the shared repository
func GetSharedPath(sharedName string) (string, error) {
	fileName := systools.SanitizeFilename(sharedName, true) // make sure we have an valid filename
	if path, err := os.UserHomeDir(); err != nil {          // get the user home dir
		return "", err
	} else {
		return filepath.FromSlash(path + "/.contxt/" + Sharedpath + fileName), nil // add the filename. sharedDir have the pathSeperator
	}
}
