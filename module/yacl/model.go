// Copyright (c) 2022 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// # Licensed under the MIT License
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
package yacl

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/swaros/contxt/module/yamc"
)

const (
	PATH_UNSET            = 0
	PATH_HOME             = 1
	PATH_CONFIG           = 2
	PATH_ABSOLUTE         = 3
	ERROR_PATH_NOT_EXISTS = 101
	NO_CONFIG_FILES_FOUND = 102
)

type ConfigModel struct {
	setFile          string                             // sets a specific filename. so this is the only one that will be loaded
	useSpecialDir    int                                // defines the behavior og the paths used like config, user home or none a simple path (relative or absolute)
	structure        any                                // points to the config struct
	reader           []yamc.DataReader                  // list of posible readers
	lastUsedReader   yamc.DataReader                    // the last used reader
	subDirs          []string                           // subdirectories relative to the basedir (defined by useSpecialDir behavior)
	usedFile         string                             // the last used configFile that is parsed
	loadedFiles      []string                           // list of all files they processed
	dirBlackList     []string                           // a blacklist of directory names they should be ignored while looking for for configurations in sub folders
	supportMigrate   bool                               // flag to enable the migration callback
	expectNoFiles    bool                               // flag to ignore the case, that no configuration files exists. if this is not set, an error will be returned wile Load
	noConfigFilesFn  func(errCode int) error            // callback that handles issues while loading configuration files. cases are ERROR_PATH_NOT_EXISTS and NO_CONFIG_FILES_FOUND
	fileLoadCallback func(path string, cfg interface{}) // needs supportMigrate enabled. then this callback will be executed for any configuration that is parsed.
	initFn           func(strct *any)                   // Init sets a callback that can handle any need to setup the structure with defaults. or create maps ..and so on
	allowSubDirs     bool                               // flag to enables scanning sub folders while looking for config files
	allowDirPattern  string                             // regex-pattern to whitelist sub folders while looking for config files
	filesPattern     string                             // file regex-pattern while looking for config files
	errorHappened    bool                               // flag to indicate, that an error happened while loading the configuration
	chainError       error                              // the last error that happened while loading the configuration
}

// New creates a New yacl ConfigModel with default properties
func New(structure any, read ...yamc.DataReader) *ConfigModel {
	return &ConfigModel{
		useSpecialDir: PATH_UNSET,
		expectNoFiles: false,
		structure:     structure,
		reader:        read,
		allowSubDirs:  false,
	}
}

// Init sets the initialization Callbacks.
// initFn will be executed to initialize the configuration structure. can be nil
// noConfigFn is the callback for the cases, the configuration directory is not exists, there are no configurations files found
func (c *ConfigModel) Init(initFn func(strct *any), noConfigFn func(errCode int) error) *ConfigModel {
	c.initFn = initFn
	c.noConfigFilesFn = noConfigFn
	return c
}

// SetExpectNoConfigFiles disable the behavior, not existing config files will be handled as error.
// this also means, it should just ignore this issue. so if this is enabled, it will also not reported
// to the noConfigFn that might be setup in the Init handler
func (c *ConfigModel) SetExpectNoConfigFiles() *ConfigModel {
	c.expectNoFiles = true
	return c
}

// SetFilePattern defines a regex-pattern for any configuration file
// any file that is not matching, will be ignored.
func (c *ConfigModel) SetFilePattern(regex string) *ConfigModel {
	c.filesPattern = regex
	return c
}

// UseHomeDir sets the User Home-dir as entrypoint (basedir)
func (c *ConfigModel) UseHomeDir() *ConfigModel {
	c.useSpecialDir = PATH_HOME
	return c
}

// UseConfigDir sets the default config dir as entrypoint (basedir)
func (c *ConfigModel) UseConfigDir() *ConfigModel {
	c.useSpecialDir = PATH_CONFIG
	return c
}

// UseRelativeDir is just defines no special path usage. mostly it means the current folder is used (relative)
// but depending on the usage of the lib, it can also still a absolute path.
func (c *ConfigModel) UseRelativeDir() *ConfigModel {
	c.useSpecialDir = PATH_UNSET
	return c
}

// add the names of the sub directories starting from the base directory.
// it will be set as string arguments. add any subdirectory separate as argument. do not add "this-dir/next-dir".
func (c *ConfigModel) SetSubDirs(dirs ...string) *ConfigModel {
	c.subDirs = dirs
	return c
}

// Limits looking for the configurations to one file (base) name. so do not add any path to filename argument.
// this is not equal to the usage of LoadFile, because if this is combined with  scanning sub dirs, any
// configuration with this file basename will be accepted.
// if you like to load one specific file so use LoadFile instead.
// if you like more flexible, depending what files should load, define a regex-pattern and use SetFilePattern
func (c *ConfigModel) SetSingleFile(filename string) *ConfigModel {
	if filepath.Base(filename) != filename {
		c.errorHappened = true
		c.chainError = fmt.Errorf("SetSingleFile: [%s] filename should not contain any path", filename)
	}
	c.setFile = filename
	return c
}

// SetFileAndPathsByFullFilePath sets the file and the path to the file. so the file will be loaded and
// the path will be used to scan for sub directories. so this is the same as SetSingleFile and SetSubDirs
// but in one call.
func (c *ConfigModel) SetFileAndPathsByFullFilePath(fullPath string) *ConfigModel {
	c.setFile = filepath.Base(fullPath)
	c.subDirs = strings.Split(filepath.Dir(fullPath), string(filepath.Separator))
	c.useSpecialDir = PATH_ABSOLUTE
	// remove the first element, if it is empty
	// this will happen if the path starts with a slash
	// so this means also we have an absolute path
	if c.subDirs[0] == "" {
		c.subDirs = c.subDirs[1:]
		c.useSpecialDir = PATH_ABSOLUTE
	}
	return c
}

// AllowSubdirs enables scanning subdirectories to find configuration files
func (c *ConfigModel) AllowSubdirs() *ConfigModel {
	c.allowSubDirs = true
	return c
}

// AllowSubdirsByRegex same as AllowSubDirs but set a regex pattern to Whitelist folder names
func (c *ConfigModel) AllowSubdirsByRegex(regex string) *ConfigModel {
	c.allowDirPattern = regex
	c.allowSubDirs = true
	return c
}

// NoSubdirs disables scanning sub folders while looking for configuration files
func (c *ConfigModel) NoSubdirs() *ConfigModel {
	c.allowSubDirs = false
	return c
}

// SetFolderBlackList defines a simple list of sub directories, they should being ignored, while looking for configuration files
func (c *ConfigModel) SetFolderBlackList(blackListedDirs []string) *ConfigModel {
	c.dirBlackList = blackListedDirs
	c.allowSubDirs = true
	return c
}

// Empty initialize the configuration without any configuration files loading.
func (c *ConfigModel) Empty() *ConfigModel {
	if c.initFn != nil {
		c.initFn(&c.structure)
	}
	return c
}

// LoadFile loads and parses a single configuration file.
// this can be called multiple times with different files.
// the content is merged (no deep copy, so no list merge for example)
func (c *ConfigModel) LoadFile(path string) error {
	if c.errorHappened {
		return c.chainError
	}
	c.Empty()
	c.setFile = path
	var extension = filepath.Ext(path)
	return c.tryLoad(filepath.Clean(c.GetConfigPath()+"/"+path), extension)
}

func (c *ConfigModel) isBlackListed(path string) bool {
	cPath := filepath.Clean(path)
	for _, badDir := range c.dirBlackList {
		if filepath.Clean(badDir) == cPath {
			return true
		}
	}
	return false
}

func (c *ConfigModel) checkDir(path string) (actionREquired bool, dirError error) {
	exists, err := c.verifyPath(path)
	if exists {
		return false, nil // regular "all fine" case. no action required
	}
	// not exists but also no error. so it is just not existing.
	// that means we have to call the handler they is responsible
	// to handle not existing paths and other things.
	// if this handler is not exists, we will handle this as an error
	// but make a hint how to handle expected "not existing directories"
	if err == nil {
		if c.noConfigFilesFn != nil {
			return true, c.noConfigFilesFn(ERROR_PATH_NOT_EXISTS)
		}
		return true, errors.New("the path " + path + " not exists. is this a expected behavior, and/or should something being done, use the Init-Handler and react to ERROR_PATH_NOT_EXISTS ")
	}
	return true, err // any other error that is different to dir not exists
}

func (c *ConfigModel) dirIsAllowed(path string) bool {
	// base config path is always allowed
	if path == c.GetConfigPath() {
		return true
	}

	if c.allowDirPattern != "" {
		if match, err := regexp.MatchString(c.allowDirPattern, path); err == nil {
			return match // an error is always false.
		}
	}
	return c.allowSubDirs // if non of the checks have decided, it just depends if we enabled using sub dirs
}

func (c *ConfigModel) filePattenCheck(path string) bool {
	if c.filesPattern != "" {
		if match, err := regexp.MatchString(c.filesPattern, path); err == nil {
			return match
		}
	} else {
		return true // empty pattern. always true
	}
	return false // this can only be reached if the regex was not working
}

// Load start loading all configuration files depends the configured behavior.
func (c *ConfigModel) Load() error {
	if c.errorHappened {
		return c.chainError
	}
	c.Empty()

	// do we have loaders?
	if len(c.reader) == 0 {
		return errors.New("no loaders assigned. add add least one Reader on New(&cf, ...)")
	}
	dir := c.GetConfigPath()

	if action, dErr := c.checkDir(dir); dErr != nil {
		return dErr
	} else if action { // action == true means we do not report a error, but the directory can not be used
		return nil // .. so we get out now
	}

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if c.filePattenCheck(path) {
				var basename = filepath.Base(path)
				var extension = filepath.Ext(path)
				if c.setFile == "" || (basename == c.setFile) { // loading a single file
					return c.tryLoad(path, extension)

				} else if c.setFile == "" { // loading all files and override anything by the last used config. but not if we expect a single file is used
					c.tryLoad(path, extension)
				}
			}
		} else {
			if c.allowSubDirs {
				if c.isBlackListed(path) || !c.dirIsAllowed(path) {
					return filepath.SkipDir
				}
			} else {
				if !c.dirIsAllowed(path) {
					return filepath.SkipDir
				}
			}

		}
		return nil
	})
	// no file could be used to load some config. if this is not expected
	// and there should exists a config, then it also depends on the handler callback return, if we
	// have a error case
	if !c.expectNoFiles && err == nil && c.usedFile == "" {
		if c.noConfigFilesFn != nil {
			return c.noConfigFilesFn(NO_CONFIG_FILES_FOUND)
		}
		return errors.New("at least one Configuration should exists. but found nothing")
	}
	return err
}

func (c *ConfigModel) detectFilename() string {
	filename := ""
	if c.usedFile != "" { // do we have a file already used?
		filename = c.usedFile

	} else if c.setFile != "" { // or do we have a file defined?
		filename = filepath.Clean(c.GetConfigPath() + "/" + c.setFile)
	}
	if filename == "" { // still no filename? then compose it
		return filepath.Clean(c.GetConfigPath() + "/" + filename)
	}
	return filename
}

// Save is try to write the current configuration on disk.
// IF we successfully loaded content at least from one configuration file, the last one is used
// IF we have setup a SingleFile and do not have a usage while loading, then this will be used instead.
// anything else will report an error
func (c *ConfigModel) Save() error {
	if len(c.reader) < 1 {
		return errors.New("we need at least one DataReader. the fist assigned will be used for write operations")
	}
	filename := c.detectFilename()
	if filename == "" {
		return errors.New("could not detect the filename for the config")
	}

	if ym, err := c.GetAsYmac(); err == nil {
		if str, sErr := ym.ToString(c.reader[0]); sErr == nil {
			data := []byte(str)
			if err := os.WriteFile(filename, data, 0644); err != nil {
				return err
			}
			return nil
		} else {
			return sErr
		}
	} else {
		return err
	}
}

// GetConfigPath compose the current used Configuration folder and returns them.
// anything what can go wrong will end up in a panic
func (c *ConfigModel) GetConfigPath() string {
	dir := "."
	startSep := ""
	switch c.useSpecialDir {
	case PATH_HOME:
		if usrDir, err := os.UserHomeDir(); err != nil {
			panic(err) // if this fails, there is something terrible wrong. a good reason for panic
		} else {
			dir = usrDir
		}
		startSep = "/"
	case PATH_CONFIG:
		if usrCfgDir, err := os.UserConfigDir(); err != nil {
			panic(err) // if this fails, there is something terrible wrong. a good reason for panic
		} else {
			dir = usrCfgDir
		}
		startSep = "/"
	case PATH_ABSOLUTE:
		dir = ""                       // this is the root of the system. we add / later
		if runtime.GOOS != "windows" { // windows does not have a root folder. it is C:\
			startSep = "/"
		}
	default:
		startSep = "/"
	}

	if len(c.subDirs) > 0 {
		dir += startSep + strings.Join(c.subDirs, "/")
	}
	return filepath.Clean(dir)
}

func (c *ConfigModel) verifyPath(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// GetValue parsing a string with dots, to use any part of them to build
// a route to a specific entry. This is a very basic path building, without
// any magic. so even a key with dots will be an issue.
// so the use case depends on the actual structure.
// Also DO NOT USE THIS FOR READING CONFIG VALUES.
// use the structure itself.
// this method is a simple helper for verify data while testing (for example)
func (c *ConfigModel) GetValue(dotedPath string) (any, error) {
	if ym, err := c.GetAsYmac(); err == nil {
		return ym.FindValue(dotedPath)
	} else {
		return nil, err
	}
}

// ToString converts the current configuration into a string, depending the
// submitted reader.
func (c *ConfigModel) ToString(reader yamc.DataReader) (string, error) {
	if ym, err := c.CreateYamc(reader); err == nil {
		return ym.ToString(reader)
	} else {
		return "", err
	}
}

// GetAsYmac creates a Yamc map from the configuration. we use Json here as Reader
func (c *ConfigModel) GetAsYmac() (*yamc.Yamc, error) {

	return c.CreateYamc(yamc.NewJsonReader())
}

// CreateYamc creates a Yamc container. it has no caching because this is needed only for
// some content creation, like saving or parsing. so there is no need to keep them.
func (c *ConfigModel) CreateYamc(reader yamc.DataReader) (*yamc.Yamc, error) {
	asYamc := yamc.New()
	if data, err := reader.Marshal(c.structure); err != nil {
		return asYamc, err
	} else {
		errParse := asYamc.Parse(reader, data)
		return asYamc, errParse
	}
}

func (c *ConfigModel) tryLoad(path, ext string) error {
	for _, loader := range c.reader {
		for _, ex := range loader.SupportsExt() {
			if strings.EqualFold("."+ex, ext) {
				if err := loader.FileDecode(path, c.structure); err == nil {
					c.lastUsedReader = loader
					if c.supportMigrate { // migrate the config
						c.fileLoadCallback(path, c.structure)
					}
					c.usedFile = path
					c.loadedFiles = append(c.loadedFiles, path)
				} else {
					return err
				}
				return nil
			}
		}
	}
	return nil
}

// return the last used reader for loading the configuration
// this is nil, if no configuration was loaded
func (c *ConfigModel) GetLastUsedReader() yamc.DataReader {
	return c.lastUsedReader
}

// GetLoadedFile returns the used configuration filename
func (c *ConfigModel) GetLoadedFile() string {
	return c.usedFile
}

// Reset the configuration. this will clear the usedFile and loadedFiles.
// this is useful for testing
func (c *ConfigModel) Reset() {
	c.usedFile = ""
	c.loadedFiles = []string{}
}

// GetAllParsedFiles returns all parsed configuration filenames
func (c *ConfigModel) GetAllParsedFiles() []string {
	return c.loadedFiles
}
