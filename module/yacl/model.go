package yacl

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/swaros/contxt/module/yamc"
)

const (
	PATH_UNSET            = 0
	PATH_HOME             = 1
	PATH_CONFIG           = 2
	ERROR_PATH_NOT_EXISTS = 101
	NO_CONFIG_FILES_FOUND = 102
)

type VersionRealtion struct {
	main   int
	mid    int
	min    int
	prefix string
}

type ConfigModel struct {
	setFile          string // sets a specific filename. so this is the only one that will be loaded
	useSpecialDir    int
	structure        any
	version          VersionRealtion
	reader           []yamc.DataReader
	subDirs          []string
	usedFile         string
	loadedFiles      []string
	supportMigrate   bool
	expectNoFiles    bool
	noConfigFilesFn  func(errCode int) error
	fileLoadCallback func(path string, cfg interface{})
	initFn           func(strct *any)
}

func NewConfig(structure any, read ...yamc.DataReader) *ConfigModel {
	return &ConfigModel{
		useSpecialDir: PATH_UNSET,
		expectNoFiles: false,
		structure:     structure,
		reader:        read,
	}
}

func (c *ConfigModel) Init(initFn func(strct *any), noConfigFn func(errCode int) error) *ConfigModel {
	c.initFn = initFn
	c.noConfigFilesFn = noConfigFn
	return c
}

func (c *ConfigModel) SetExpectNoConfigFiles() *ConfigModel {
	c.expectNoFiles = true
	return c
}

func (c *ConfigModel) UseHomeDir() *ConfigModel {
	c.useSpecialDir = PATH_HOME
	return c
}

func (c *ConfigModel) UseConfigDir() *ConfigModel {
	c.useSpecialDir = PATH_CONFIG
	return c
}

func (c *ConfigModel) UseRelativeDir() *ConfigModel {
	c.useSpecialDir = PATH_UNSET
	return c
}

func (c *ConfigModel) SetSubDirs(dirs ...string) *ConfigModel {
	c.subDirs = dirs
	return c
}

func (c *ConfigModel) SetSingleFile(filename string) *ConfigModel {
	c.setFile = filename
	return c
}

func (c *ConfigModel) SetVersion(prefix string, main, mid, minor int) *ConfigModel {
	c.version.main = main
	c.version.mid = mid
	c.version.min = minor
	c.version.prefix = prefix
	return c
}
func (c *ConfigModel) Empty() *ConfigModel {
	if c.initFn != nil {
		c.initFn(&c.structure)
	}
	return c
}

func (c *ConfigModel) LoadFile(path string) error {
	c.Empty()
	c.setFile = path
	var extension = filepath.Ext(path)
	return c.tryLoad(filepath.Clean(c.GetConfigPath()+"/"+path), extension)
}

func (c *ConfigModel) checkDir(path string) error {
	exists, err := c.verifyPath(path)
	if exists {
		return nil
	}
	// not exists but also no error. so it is just not existing
	// thate means we have to call the handler they is reponsible
	// to handle not existing paths and other things.
	// if this handler is not exists, we will handle this as an error
	// but make a hint how to handle expected "not existings directories"
	if err == nil {
		if c.noConfigFilesFn != nil {
			return c.noConfigFilesFn(NO_CONFIG_FILES_FOUND)
		}
		return errors.New("the path " + path + " not exists. is this a expected behavior, and/or should somethig being done, create use the Inithandler and react to ERROR_PATH_NOT_EXISTS ")
	}

	return err // any other error that is different to dir not exists
}

func (c *ConfigModel) Load() error {
	c.Empty()

	// do we have loaders?
	if len(c.reader) == 0 {
		return errors.New("no loaders assigned. add add least one Reader on NewConfig(&cf, ...)")
	}
	dir := c.GetConfigPath()

	if dErr := c.checkDir(dir); dErr != nil {
		return dErr
	}

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			var basename = filepath.Base(path)
			var extension = filepath.Ext(path)
			if c.setFile == "" || (basename == c.setFile) { // loading a single file
				return c.tryLoad(path, extension)

			} else if c.setFile == "" { // loading all files and override anything by the last used config. but not if we expect a single file is used
				c.tryLoad(path, extension)
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
	if c.usedFile != "" {
		filename = c.usedFile
	} else if c.setFile != "" {
		filename = c.setFile
	}
	if filename != "" {
		return filepath.Clean(c.GetConfigPath() + "/" + filename)
	}
	return filename
}

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

func (c *ConfigModel) GetConfigPath() string {
	dir := "."

	switch c.useSpecialDir {
	case PATH_HOME:
		if usrDir, err := os.UserHomeDir(); err != nil {
			panic(err) // if this fails, there is something terrible wrong. a good reason for panic
		} else {
			dir = usrDir
		}
	case PATH_CONFIG:
		if usrCfgDir, err := os.UserConfigDir(); err != nil {
			panic(err) // if this fails, there is something terrible wrong. a good reason for panic
		} else {
			dir = usrCfgDir
		}
	}

	if len(c.subDirs) > 0 {
		dir += "/" + strings.Join(c.subDirs, "/")
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

func (c *ConfigModel) GetAsYmac() (*yamc.Yamc, error) {

	return c.CreateYamc(yamc.NewJsonReader())
}

func (c *ConfigModel) GetValue(dotedPath string) (any, error) {
	if ym, err := c.GetAsYmac(); err == nil {
		return ym.FindValue(dotedPath)
	} else {
		return nil, err
	}
}

func (c *ConfigModel) ToString(reader yamc.DataReader) (string, error) {
	if ym, err := c.CreateYamc(reader); err == nil {
		return ym.ToString(reader)
	} else {
		return "", err
	}
}

func (c *ConfigModel) CreateYamc(reader yamc.DataReader) (*yamc.Yamc, error) {
	asYamc := yamc.NewYmac()
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

func (c *ConfigModel) GetLoadedFile() string {
	return c.usedFile
}

func (c *ConfigModel) GetAllParsedFiles() []string {
	return c.loadedFiles
}
