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
	PATH_UNSET  = 0
	PATH_HOME   = 1
	PATH_CONFIG = 2
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
	structure        *any
	version          VersionRealtion
	reader           []yamc.DataReader
	subDirs          []string
	usedFile         string
	loadedFiles      []string
	supportMigrate   bool
	fileLoadCallback func(path string, cfg interface{})
	initFn           func(strct *any)
	asYamc           *yamc.Yamc
}

func NewConfig(structure any, read ...yamc.DataReader) *ConfigModel {
	return &ConfigModel{
		useSpecialDir: PATH_UNSET,
		structure:     &structure,
		reader:        read,
	}
}

func (c *ConfigModel) Init(initFn func(strct *any)) *ConfigModel {
	c.initFn = initFn
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
		c.initFn(c.structure)
	}
	return c
}

func (c *ConfigModel) LoadFile(path string) error {
	c.Empty()
	c.setFile = path
	var extension = filepath.Ext(path)
	return c.tryLoad(filepath.Clean(c.GetConfigPath()+"/"+path), extension)
}

func (c *ConfigModel) Load() error {
	c.Empty()
	dir := c.GetConfigPath()

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			var basename = filepath.Base(path)
			var extension = filepath.Ext(path)
			if c.setFile == "" || (basename == c.setFile) { // loading a single file
				return c.tryLoad(path, extension)

			} else { // loading all files and override anything by the last used config
				c.tryLoad(path, extension)
			}
		}
		return nil
	})

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

func (c *ConfigModel) GetAsYmac() (*yamc.Yamc, error) {
	if c.asYamc == nil {
		if err := c.createYamc(yamc.NewJsonReader()); err != nil {
			return nil, err
		}
	}
	return c.asYamc, nil
}

func (c *ConfigModel) GetValue(gjsonPath string) (any, error) {
	if ym, err := c.GetAsYmac(); err == nil {
		return ym.GetGjsonValue(gjsonPath)
	} else {
		return nil, err
	}
}

func (c *ConfigModel) ToString(reader yamc.DataReader) (string, error) {
	if ym, err := c.GetAsYmac(); err == nil {
		return ym.ToString(reader)
	} else {
		return "", err
	}
}

func (c *ConfigModel) createYamc(reader yamc.DataReader) error {
	c.asYamc = yamc.NewYmac()
	if data, err := reader.Marshal(c.structure); err != nil {
		return err
	} else {
		c.asYamc.Parse(reader, data)
	}
	return nil
}

func (c *ConfigModel) tryLoad(path, ext string) error {
	for _, loader := range c.reader {
		for _, ex := range loader.SupportsExt() {
			if strings.EqualFold("."+ex, ext) {
				if err := loader.FileDecode(path, &c.structure); err == nil {
					if c.supportMigrate { // migrate the config
						c.fileLoadCallback(path, &c.structure)
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
