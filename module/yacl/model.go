package yacl

import (
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
	setFile       string // sets a specific filename. so this is the only one that will be loaded
	useSpecialDir int
	structure     *any
	version       VersionRealtion
	reader        []yamc.DataReader
	subDirs       []string
	usedFile      string
	loadedFiles   []string
}

func NewConfig(structure any, read ...yamc.DataReader) *ConfigModel {
	return &ConfigModel{
		useSpecialDir: PATH_UNSET,
		structure:     &structure,
		reader:        read,
	}
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

func (c *ConfigModel) Load() error {
	dir := c.GetConfigPath()

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			var basename = filepath.Base(path)
			if c.setFile == "" || (basename == c.setFile) {
				var extension = filepath.Ext(path)
				return c.tryLoad(path, extension)

			}
		}
		return nil
	})

	return err
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

func (c *ConfigModel) tryLoad(path, ext string) error {
	for _, loader := range c.reader {
		for _, ex := range loader.SupportsExt() {
			if strings.EqualFold("."+ex, ext) {
				if err := loader.FileDecode(path, &c.structure); err == nil {
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
