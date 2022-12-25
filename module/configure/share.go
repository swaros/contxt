package configure

import "github.com/swaros/contxt/module/systools"

// GetSharedPath returns the full path to the shared repository
func GetSharedPath(sharedName string) (string, error) {
	fileName := systools.SanitizeFilename(sharedName, true)       // make sure we have an valid filename
	sharedDir, err := NewContxtConfig().GetConfigPath(Sharedpath) // get the path where we store shared repos
	if err == nil {
		var configPath = sharedDir + fileName // add the filename. sharedDir have the pathSeperator
		return configPath, err
	}
	return "", err
}
