package configure

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/swaros/contxt/context/systools"
)

const (
	// DefaultConfigFileName is the main config json file name.
	DefaultConfigFileName = "config.json"
	// DefaultPath this is the default path to store gocd configurations
	DefaultPath = "/.contxt/"
	// DefaultWorkspace this is the main configuration workspace
	DefaultWorkspace = "default"
)

// Config is the the current used configuration
var Config Configuration

func getUserDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir, err
}

// InitConfig initilize the configuration files
func InitConfig() error {
	err := checkSetup()
	return err
}

// ClearPaths removes all paths from current Workspace
func ClearPaths() {
	Config.Paths = nil
	SaveDefaultConfiguration(true)
}

func createDefaultConfig() {
	Config.CurrentSet = DefaultWorkspace
}

// ChangeWorkspace changing workspace
func ChangeWorkspace(workspace string) {
	// save current configuration
	SaveDefaultConfiguration(true)
	// change set name
	Config.CurrentSet = workspace
	SaveDefaultConfiguration(false)
	err := checkSetup()
	if err != nil {
		fmt.Println(err)
	}
	SaveDefaultConfiguration(true)
}

// RemoveWorkspace removes a workspace
func RemoveWorkspace(name string) {
	if name == Config.CurrentSet {
		fmt.Println("can not remove current workspace")
		return
	}
	path, err := getConfigPath(name + ".json")
	if err == nil {
		var cfgExists, err = exists(path)
		if err == nil && cfgExists == true {
			os.Remove(path)
		} else {
			fmt.Println("no configuration found")
		}
	} else {
		fmt.Println(err)
	}
}

// SaveDefaultConfiguration stores the current configuration as default
func SaveDefaultConfiguration(workSpaceConfigUpdate bool) {
	path, err := getConfigPath(DefaultConfigFileName)
	if err == nil {
		SaveConfiguration(Config, path)
		// save workspace config too
		if workSpaceConfigUpdate {
			pathWorkspace, secErr := getConfigPath(Config.CurrentSet + ".json")
			if secErr == nil {
				SaveConfiguration(Config, pathWorkspace)
			}
		}
	} else {
		fmt.Println(err)
	}
}

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
	var files []string
	files = ListWorkSpaces()

	if len(files) > 0 {
		fmt.Println("  ", systools.Yellow("all existing workspaces"))
		for _, file := range files {
			var basePath = filepath.Base(file)
			var extension = filepath.Ext(file)
			// display json files only they are not the default config
			if extension == ".json" && basePath != DefaultConfigFileName {
				basePath = strings.TrimSuffix(basePath, extension)
				// we are also not interested in the default workspace
				if basePath != DefaultWorkspace {
					fmt.Println("\t", systools.Teal(basePath))
				}
			}
		}
	} else {
		fmt.Println("\t", systools.Yellow("No workspaces exists"))
	}
}

// ShowPaths : display all stored paths in the workspace
func ShowPaths(current string) int {

	PathWorker(current, func(index int, path string) {
		if path == current {
			fmt.Println("\t[", systools.Yellow(index), "]\t", path)
		} else {
			fmt.Println("\t ", systools.Purple(index), " \t", path)
		}
	})
	return len(Config.Paths)
}

// PathWorker executes a callback function in a path
func PathWorker(current string, callback func(int, string)) {
	cnt := len(Config.Paths)
	if cnt < 1 {
		fmt.Println("\t", systools.Warn("no paths actually stored"))
		return
	}
	for index, path := range Config.Paths {
		callback(index, path)
	}
}

func loadConfigurationFile(path string) {
	file, _ := os.Open(path)
	defer file.Close()
	decoder := json.NewDecoder(file)

	err := decoder.Decode(&Config)
	if err != nil {
		fmt.Println("error:", err)
	}
}

// SaveConfiguration : stores configuration in given path
func SaveConfiguration(config Configuration, path string) {
	b, _ := json.MarshalIndent(config, "", " ")
	err := ioutil.WriteFile(path, b, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func getConfigPath(fileName string) (string, error) {
	homeDir, err := getUserDir()
	if err == nil {
		var configPath = homeDir + DefaultPath + fileName
		return configPath, err
	}
	return homeDir, err
}

// AddPath adding a path if they not already exists
func AddPath(path string) {
	if pathExists(path) {
		fmt.Println(systools.Warn("\terror"), systools.Info(path), "already in set", systools.Green(Config.CurrentSet))
		return
	}

	Config.Paths = append(Config.Paths, path)
}

func pathExists(pathSearch string) bool {
	for _, path := range Config.Paths {
		if pathSearch == path {
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
		pathExists, err := exists(dirPath)
		// path dos not exists. create it
		if pathExists == false && err == nil {
			err := os.Mkdir(dirPath, os.ModePerm)
			if err != nil {
				log.Fatal(err)
				return err
			}
		}

		configFileExists, err := exists(configPath)
		// no config file exists. create default config
		if configFileExists == false && err == nil {
			createDefaultConfig()
		}

		// load config file
		if configFileExists == true && err == nil {
			loadConfigurationFile(configPath)
			if Config.CurrentSet == "" {
				Config.CurrentSet = DefaultWorkspace
				SaveDefaultConfiguration(false)
			}
			// now copy content of set
			pathWorkspace, secErr := getConfigPath(Config.CurrentSet + ".json")
			if secErr == nil {
				confPathExists, _ := exists(pathWorkspace)
				if confPathExists {
					loadConfigurationFile(pathWorkspace)
				} else {
					Config.Paths = nil
				}
			}
		}

	}
	return err
}

func getUserConfig() (string, error) {
	homeDir, err := getUserDir()
	if err == nil {
		log.Println(homeDir)
	}
	return homeDir, err
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
