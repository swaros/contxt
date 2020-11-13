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

	"github.com/swaros/contxt/context/output"
)

const (
	// DefaultConfigFileName is the main config json file name.
	DefaultConfigFileName = "config.json"
	// DefaultPath this is the default path to store gocd configurations
	DefaultPath = "/.contxt/"
	// DefaultWorkspace this is the main configuration workspace
	DefaultWorkspace = "default"
	// MirrorPath path for local mirror
	MirrorPath = "mirror/"
)

// UsedConfig is the the current used configuration
var UsedConfig Configuration

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
	UsedConfig.Paths = nil
	SaveDefaultConfiguration(true)
}

func createDefaultConfig() {
	UsedConfig.CurrentSet = DefaultWorkspace
}

// ChangeWorkspace changing workspace
func ChangeWorkspace(workspace string, oldspace func(string) bool, newspace func(string)) {
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
		fmt.Println(output.MessageCln("current workspace is now:", output.BackBlue, output.ForeWhite, workspace))
	} else {
		fmt.Println(output.MessageCln(output.ForeLightYellow, "changing workspace failed."))
	}
}

// RemoveWorkspace removes a workspace
func RemoveWorkspace(name string) {
	if name == UsedConfig.CurrentSet {
		fmt.Println("can not remove current workspace")
		return
	}
	path, err := getConfigPath(name + ".json")
	if err == nil {
		var cfgExists, err = exists(path)
		if err == nil && cfgExists == true {
			os.Remove(path)
		} else {
			fmt.Println("no workspace exists with name: ", output.MessageCln(output.ForeLightYellow, name))
			os.Exit(5)
		}
	} else {
		fmt.Println(err)
	}
}

// SaveDefaultConfiguration stores the current configuration as default
func SaveDefaultConfiguration(workSpaceConfigUpdate bool) {
	path, err := getConfigPath(DefaultConfigFileName)
	if err == nil {
		SaveConfiguration(UsedConfig, path)
		// save workspace config too
		if workSpaceConfigUpdate {
			pathWorkspace, secErr := getConfigPath(UsedConfig.CurrentSet + ".json")
			if secErr == nil {
				SaveConfiguration(UsedConfig, pathWorkspace)
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

// WorkSpaces handler to iterate all workspaces
func WorkSpaces(callback func(string)) {
	var files []string
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
					callback(basePath)
				}
			}
		}
	}
}

// ShowPaths : display all stored paths in the workspace
func ShowPaths(current string) int {

	PathWorker(func(index int, path string) {
		if path == current {

			fmt.Println(output.MessageCln("\t[", output.ForeLightYellow, index, output.CleanTag, "]\t", output.BoldTag, path))
		} else {
			fmt.Println(output.MessageCln("\t ", output.ForeLightBlue, index, output.CleanTag, " \t", path))
		}

	})
	return len(UsedConfig.Paths)
}

// PathWorker executes a callback function in a path
func PathWorker(callback func(int, string)) {
	cnt := len(UsedConfig.Paths)
	if cnt < 1 {
		fmt.Println(output.MessageCln("\t", output.ForeRed, "no paths actually stored"))
		return
	}
	for index, path := range UsedConfig.Paths {
		os.Chdir(path)
		callback(index, path)
	}
}

func loadConfigurationFile(path string) {
	file, _ := os.Open(path)
	defer file.Close()
	decoder := json.NewDecoder(file)

	err := decoder.Decode(&UsedConfig)
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
		fmt.Println(output.MessageCln(output.ForeYellow, "\tthe path is already in set ", output.BoldTag, UsedConfig.CurrentSet))
		return
	}

	UsedConfig.Paths = append(UsedConfig.Paths, path)
}

func pathExists(pathSearch string) bool {
	for _, path := range UsedConfig.Paths {
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
			if UsedConfig.CurrentSet == "" {
				UsedConfig.CurrentSet = DefaultWorkspace
				SaveDefaultConfiguration(false)
			}
			// now copy content of set
			pathWorkspace, secErr := getConfigPath(UsedConfig.CurrentSet + ".json")
			if secErr == nil {
				confPathExists, _ := exists(pathWorkspace)
				if confPathExists {
					loadConfigurationFile(pathWorkspace)
				} else {
					UsedConfig.Paths = nil
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
