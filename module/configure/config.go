// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Licensed under the MIT License
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package configure

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/manout"
)

const (
	// DefaultConfigFileName is the main config json file name.
	DefaultConfigFileName = "contxt_current_config.json"
	// DefaultPath this is the default path to store gocd configurations
	DefaultPath = "/.contxt/"
	// DefaultWorkspace this is the main configuration workspace
	DefaultWorkspace = "default_contxt_ws"
	// MirrorPath path for local mirror
	MirrorPath = "mirror/"
	// path where shared repositories are stored
	Sharedpath = "shared/"
)

// UsedConfig is the the current used configuration
var UsedConfig Configuration

var badCharacters = []string{
	"../",
	"<!--",
	"-->",
	"<",
	">",
	"'",
	"\"",
	"&",
	"$",
	"#",
	"{", "}", "[", "]", "=",
	";", "?", "%20", "%22",
	"%3c",   // <
	"%253c", // <
	"%3e",   // >
	"",      // > -- fill in with % 0 e - without spaces in between
	"%28",   // (
	"%29",   // )
	"%2528", // (
	"%26",   // &
	"%24",   // $
	"%3f",   // ?
	"%3b",   // ;
	"%3d",   // =
}

func RemoveBadCharacters(input string, dictionary []string) string {

	temp := input

	for _, badChar := range dictionary {
		temp = strings.Replace(temp, badChar, "", -1)
	}
	return temp
}

func SanitizeFilename(name string, relativePath bool) string {

	// default settings
	var badDictionary []string = badCharacters

	if name == "" {
		return name
	}

	// if relativePath is TRUE, we preserve the path in the filename
	// If FALSE and will cause upper path foldername to merge with filename
	// USE WITH CARE!!!

	if !relativePath {
		// add additional bad characters
		badDictionary = append(badCharacters, "./")
		badDictionary = append(badDictionary, "/")
	}

	// trim(remove)white space
	trimmed := strings.TrimSpace(name)

	// trim(remove) white space in between characters
	trimmed = strings.Replace(trimmed, " ", "", -1)

	// remove bad characters from filename
	trimmed = RemoveBadCharacters(trimmed, badDictionary)

	stripped := strings.Replace(trimmed, "\\", "", -1)

	return stripped
}

func getUserDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir, err
}

// GetSharedPath returns the full path to the shared repository
func GetSharedPath(sharedName string) (string, error) {
	fileName := SanitizeFilename(sharedName, true) // make sure we have an valid filename
	sharedDir, err := GetConfigPath(Sharedpath)    // get the path where we store shared repos
	if err == nil {
		var configPath = sharedDir + fileName // add the filename. sharedDir have the pathSeperator
		return configPath, err
	}
	return "", err
}

// InitConfig initilize the configuration files
func InitConfig() error {
	err := checkSetup()
	return err
}

// ClearPaths removes all paths from current Workspace
func ClearPaths() {
	UsedConfig.Paths = nil
	SaveDefaultConfiguration(true)
}

func createDefaultConfig() {
	UsedConfig.CurrentSet = DefaultWorkspace
}

// ChangeWorkspace changing workspace
func ChangeWorkspace(workspace string, oldspace func(string) bool, newspace func(string)) error {
	// triggers execution of checking old Workspace
	canChange := oldspace(UsedConfig.CurrentSet)
	if canChange {
		// save current configuration
		SaveDefaultConfiguration(true)
		// change set name
		UsedConfig.CurrentSet = workspace
		SaveDefaultConfiguration(false)
		err := checkSetup()
		if err != nil {
			fmt.Println(err)
		}
		SaveDefaultConfiguration(true)
		newspace(workspace)
		return nil
	} else {
		return errors.New("changing workspace failed")
	}
}

// RemoveWorkspace removes a workspace
func RemoveWorkspace(name string) {
	if name == UsedConfig.CurrentSet {
		fmt.Println("can not remove current workspace")
		return
	}
	path, err := GetConfigPath(name + ".json")
	if err == nil {
		var cfgExists, err = exists(path)
		if err == nil && cfgExists {
			os.Remove(path)
		} else {
			fmt.Println("no workspace exists with name: ", manout.MessageCln(manout.ForeLightYellow, name))
			systools.Exit(systools.ErrorWhileLoadCfg)
		}
	} else {
		fmt.Println(err)
	}
}

// SaveDefaultConfiguration stores the current configuration as default
func SaveDefaultConfiguration(workSpaceConfigUpdate bool) error {
	path, err := GetConfigPath(DefaultConfigFileName)
	if err == nil {
		if err := SaveConfiguration(UsedConfig, path); err != nil {
			return err
		}
		// save workspace config too
		if workSpaceConfigUpdate {
			pathWorkspace, secErr := GetConfigPath(UsedConfig.CurrentSet + ".json")
			if secErr == nil {
				return SaveConfiguration(UsedConfig, pathWorkspace)
			}
		}
	} else {
		return err
	}
	return nil
}

func SaveActualPathByIndex(useIndex int) error {
	if useIndex >= 0 && useIndex != UsedConfig.LastIndex {
		UsedConfig.LastIndex = useIndex
		return SaveDefaultConfiguration(true)
	}
	if useIndex < 0 {
		return errors.New("invalid index number")
	}
	// just no need to save. no error.
	return nil
}

func SaveActualPathByPath(pathToSave string) error {
	for index, path := range UsedConfig.Paths {
		if path == pathToSave {
			return SaveActualPathByIndex(index)
		}
	}
	return errors.New("this path " + pathToSave + " is not part of the stored paths")
}

// PathWorkerWithCd executes a callback function in a path
func PathWorker(callbackInDirextory func(int, string), callbackBackToOrigin func(origin string)) error {
	// checking current directory. store it for going back later
	if current, err := os.Getwd(); err != nil {
		return err
	} else {
		// do we have targets? then iterate
		// on them, got to the directory, and execute the callback
		cnt := len(UsedConfig.Paths)
		if cnt < 1 {
			return errors.New("no paths actually stored ")
		}
		for index, path := range UsedConfig.Paths {
			if err := os.Chdir(path); err == nil {
				callbackInDirextory(index, path)
			} else {
				return err
			}
		}
		// now it is time for going back to the dir we was before
		if err := os.Chdir(current); err == nil {
			callbackBackToOrigin(current)
		} else {
			return err
		}
		return nil
	}
}

func PathWorkerNoCd(callback func(int, string)) error {
	cnt := len(UsedConfig.Paths)
	if cnt < 1 {
		return errors.New("no paths actually stored ")
	}
	for index, path := range UsedConfig.Paths {
		callback(index, path)
	}
	return nil
}

func loadConfigurationFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)

	return decoder.Decode(&UsedConfig)
}

func LoadExtConfiguration(path string) (Configuration, error) {
	file, err := os.Open(path)
	if err != nil {
		return Configuration{}, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var cfg = Configuration{}
	err = decoder.Decode(&cfg)

	return cfg, err
}

// SaveConfiguration : stores configuration in given path
func SaveConfiguration(config Configuration, path string) error {
	b, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0644)

}

// getConfigPath returns the user related  path
// it do not checks if this exists
func GetConfigPath(fileName string) (string, error) {
	homeDir, err := getUserDir()
	if err == nil {
		var configPath = homeDir + DefaultPath + fileName
		return configPath, err
	}
	return homeDir, err
}

// AddPath adding a path if they not already exists
func AddPath(path string) {
	if PathExists(path) {
		fmt.Println(manout.MessageCln(manout.ForeYellow, "\tthe path is already in set ", manout.BoldTag, UsedConfig.CurrentSet))
		return
	}

	UsedConfig.Paths = append(UsedConfig.Paths, path)
}

// RemovePath removes a path from current set.
// returns true is path was found and removed
func RemovePath(path string) bool {
	var newSlice []string
	found := false
	if len(UsedConfig.Paths) < 1 {
		return false
	}
	for _, pathIn := range UsedConfig.Paths {
		if pathIn != path {
			newSlice = append(newSlice, pathIn)
		} else {
			found = true
		}
	}
	if found {
		UsedConfig.Paths = newSlice
	}
	return found
}

// PathExists checks if this path is stored in current Workspace
func PathExists(pathSearch string) bool {
	for _, path := range UsedConfig.Paths {
		if pathSearch == path {
			return true
		}
	}
	return false
}

// PathMeightPartOfWs checks if the path meight be a part of this workspace
func PathMeightPartOfWs(pathSearch string) bool {
	for _, path := range UsedConfig.Paths {
		if strings.Contains(pathSearch, path) {
			return true
		}
	}
	return false
}

func checkSetup() error {
	homeDir, err := getUserDir()
	if err == nil {
		var dirPath = homeDir + DefaultPath
		var configPath = homeDir + DefaultPath + DefaultConfigFileName
		if pathExists, err := exists(dirPath); err != nil {
			return err
		} else {
			// path dos not exists. create it
			if !pathExists && err == nil {
				err := os.Mkdir(dirPath, os.ModePerm)
				if err != nil {
					log.Fatal(err)
					return err
				}
			}

			configFileExists, err := exists(configPath)
			if err != nil {
				return err
			}
			// no config file exists. create default config
			if !configFileExists {
				createDefaultConfig()
			} else {
				if lErr := loadConfigurationFile(configPath); lErr != nil {
					return lErr
				}
				if UsedConfig.CurrentSet == "" {
					UsedConfig.CurrentSet = DefaultWorkspace
					SaveDefaultConfiguration(false)
				}
				// now copy content of set
				pathWorkspace, secErr := GetConfigPath(UsedConfig.CurrentSet + ".json")
				if secErr == nil {
					confPathExists, _ := exists(pathWorkspace)
					if confPathExists {
						return loadConfigurationFile(pathWorkspace)
					} else {
						UsedConfig.Paths = nil
					}
				} else {
					return secErr
				}
			}
		}
	}
	return err
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
