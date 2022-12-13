package yacl

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/swaros/contxt/module/yamc"
)

type VersionRealtion struct {
	main   int
	mid    int
	min    int
	prefix string
}

type ConfigModel struct {
	setFile     string // sets a specific filename. so this is the only one that will be loaded
	useHomeDir  bool
	structure   *any
	version     VersionRealtion
	reader      []yamc.DataReader
	subDirs     []string
	usedFile    string
	loadedFiles []string
}

func NewConfig(structure any, read ...yamc.DataReader) *ConfigModel {
	return &ConfigModel{
		structure: &structure,
		reader:    read,
	}
}

func (c *ConfigModel) UseHomeDir() *ConfigModel {
	c.useHomeDir = true
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
	dir := "."
	if len(c.subDirs) > 0 {
		dir += "/" + strings.Join(c.subDirs, "/")
	}

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
