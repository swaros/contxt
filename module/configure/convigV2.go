package configure

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yamc"
)

var CfgV1 *contxtConfigure = NewContxtConfig()

type contxtConfigure struct {
	UsedV2Config      ConfigMetaV2
	DefaultV2Yacl     *yacl.ConfigModel
	migrationRequired bool
}

func NewCfgV2(c *contxtConfigure) ConfigMetaV2 {
	var cfgV2 ConfigMetaV2
	c.DefaultV2Yacl = yacl.New(&cfgV2, yamc.NewYamlReader()).
		Init(func(strct *any) {
			cfgV2.Configs = make(map[string]ConfigurationV2)

		}, func(errCode int) error {
			c.migrationRequired = true // set flag that we may have to migrate from v1 to v2
			if errCode == yacl.ERROR_PATH_NOT_EXISTS {
				fmt.Println("configuration path not exists ", c.DefaultV2Yacl.GetConfigPath(), " try to create them")
				if err := os.MkdirAll(c.DefaultV2Yacl.GetConfigPath(), os.ModePerm); err != nil {
					panic("error while create configuration folder " + err.Error()) // at this point we should have panic
				}
			}
			return nil
		}).
		UseConfigDir().
		SetSubDirs("contxt").
		SetSingleFile("contxtv2.yml")

	if err := c.DefaultV2Yacl.Load(); err != nil {
		// errors depending not existing folders and files should already be handled without reporting as error
		// so this is something else
		panic("error while reading configuration " + err.Error())

	}
	return cfgV2
}

func NewContxtConfig() *contxtConfigure {
	var cfgV1 Configuration
	var cfgV2 ConfigMetaV2
	c := &contxtConfigure{}
	cfgV2 = NewCfgV2(c)
	c.UsedV2Config = cfgV2
	// if migration is required
	if c.migrationRequired {
		contxtCfg := yacl.New(&cfgV1, yamc.NewJsonReader()).
			UseHomeDir().
			SetSubDirs(".contxt").
			SupportMigrate(func(path string, cfg interface{}) {
				// this one contains the name of the current used config
				if strings.Contains(filepath.Clean(path), "contxt_current_config.json") {
					c.UsedV2Config.CurrentSet = cfgV1.CurrentSet
					//c.UsedV1Config = cfgV1
					return // just get out. we do not need something else from this file
				}

				// this one contains the name of the current used config
				if strings.Contains(filepath.Clean(path), "default_contxt_ws.json") {
					return // just get out. we do not need something else from this file
				}

				// here we have the logic to convert the configuration
				var cfgEntrie ConfigurationV2 = ConfigurationV2{Name: cfgV1.CurrentSet}
				for index, inPath := range cfgV1.Paths {
					var cfgPaths WorkspaceInfoV2 = WorkspaceInfoV2{
						Path: inPath,
					}
					if cfgEntrie.Paths == nil {
						cfgEntrie.Paths = make(map[string]WorkspaceInfoV2)
					}
					keyStr := fmt.Sprintf("%d", index)
					cfgEntrie.Paths[keyStr] = cfgPaths
				}
				// we make this already in the Init func
				if c.UsedV2Config.Configs == nil {
					fmt.Println("wtf. this should be initalized already")
					os.Exit(55)
				}
				c.UsedV2Config.Configs[cfgV1.CurrentSet] = cfgEntrie
			})

		contxtCfg.Load()                                      // process the obsolet configs.
		contxtCfg.SetSingleFile("contxt_current_config.json") // after reading all files, we want to use this file  for any write operation

		//c.DefaultV1Yacl = *contxtCfg
		c.DefaultV2Yacl.Save()
	}
	return c
}

// InitConfig initilize the configuration files
func (c *contxtConfigure) InitConfig() error {
	//err := c.CheckSetup()
	return nil
}

// helper function to get a config by the name
func (c *contxtConfigure) getConfig(name string) (*ConfigurationV2, bool) {
	if cfg, ok := c.UsedV2Config.Configs[name]; ok {
		return &cfg, true
	}
	return &ConfigurationV2{}, false
}

func (c *contxtConfigure) getCurrentConfig() (*ConfigurationV2, bool) {
	return c.getConfig(c.UsedV2Config.CurrentSet)
}

// helper funtion to change a name config
func (c *contxtConfigure) doConfigChange(name string, worker func(cfg *ConfigurationV2)) {
	if cfg, ok := c.UsedV2Config.Configs[name]; ok {
		worker(&cfg)
	}
}

// helper function to change the current used config
func (c *contxtConfigure) doCurrentConfigChange(worker func(cfg *ConfigurationV2)) {
	c.doConfigChange(c.UsedV2Config.CurrentSet, worker)
}

// ClearPaths removes all paths
func (c *contxtConfigure) ClearPaths() {
	c.doCurrentConfigChange(func(cfg *ConfigurationV2) {
		cfg.Paths = make(map[string]WorkspaceInfoV2)
	})
	c.DefaultV2Yacl.Save()
}

func (c *contxtConfigure) CheckSetup() error {
	return nil
}

func (c *contxtConfigure) ListWorkSpaces() []string {
	ws := []string{}
	for indx := range c.UsedV2Config.Configs {
		ws = append(ws, indx)
	}
	return ws
}

func (c *contxtConfigure) ExecOnWorkSpaces(callFn func(index string, cfg ConfigurationV2)) {
	for i, c := range c.UsedV2Config.Configs {
		callFn(i, c)
	}
}

func (c *contxtConfigure) ChangeWorkspace(workspace string, oldspace func(string) bool, newspace func(string)) error {
	// triggers execution of checking old Workspace
	canChange := oldspace(c.UsedV2Config.CurrentSet)
	if canChange {
		// change set name and save
		c.UsedV2Config.CurrentSet = workspace
		c.DefaultV2Yacl.Save()
		newspace(workspace) // execute any assigned newspace callback
		return nil
	} else {
		return errors.New("changing workspace failed")
	}
}

func (c *contxtConfigure) RemoveWorkspace(name string) error {
	if name == c.UsedV2Config.CurrentSet {
		return errors.New("can not remove current workspace")
	}
	if _, found := c.getConfig(name); found {
		delete(c.UsedV2Config.Configs, name)
	} else {
		return errors.New("no workspace exists with name: " + name)
	}
	return nil
}

func (c *contxtConfigure) PathWorker(callbackInDirectory func(string, string), callbackBackToOrigin func(origin string)) error {
	// checking current directory. store it for going back later
	if current, err := os.Getwd(); err != nil {
		return err
	} else {
		// do we have targets? then iterate
		// on them, got to the directory, and execute the callback
		cfg, found := c.getCurrentConfig()
		if !found {
			return errors.New("error while getting the current configuration")
		}
		cnt := len(cfg.Paths)
		if cnt < 1 {
			return errors.New("no paths actually stored ")
		}
		for index, path := range cfg.Paths {
			if err := os.Chdir(path.Path); err == nil {
				callbackInDirectory(index, path.Path)
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

func (c *contxtConfigure) PathWorkerNoCd(callback func(string, string)) error {
	cfg, found := c.getCurrentConfig()
	if !found {
		return errors.New("error while getting the current configuration")
	}
	cnt := len(cfg.Paths)
	if cnt < 1 {
		return errors.New("no paths actually stored ")
	}
	for index, path := range cfg.Paths {
		callback(index, path.Path)
	}
	return nil
}

// SaveConfiguration : stores configuration in given path
func (c *contxtConfigure) SaveConfiguration() error {
	return c.DefaultV2Yacl.Save()

}

// getConfigPath returns the user related  path
// it do not checks if this exists
func (c *contxtConfigure) GetConfigPath(fileName string) (string, error) {
	return c.DefaultV2Yacl.GetConfigPath(), nil
}

// returns the path by a given index
func (c *contxtConfigure) GetPathByIndex(index, fallback string) string {
	cfg, found := c.getCurrentConfig()
	if !found {
		return fallback
	}

	if path, ok := cfg.Paths[index]; ok {
		return path.Path
	}
	return fallback
}

// returns the path by a given index or fallback
func (c *contxtConfigure) GetActivePath(fallback string) string {
	cfg, found := c.getCurrentConfig()
	if !found {
		return fallback
	}
	if path, ok := cfg.Paths[cfg.CurrentIndex]; ok {
		return path.Path
	}
	return fallback
}

func (c *contxtConfigure) ChangeActivePath(index string) error {
	cfg, found := c.getCurrentConfig()
	if !found {
		return errors.New("could not change the index. error while getting the configuration")
	}
	if _, ok := cfg.Paths[index]; ok {
		cfg.CurrentIndex = index
		return nil
	} else {
		return errors.New("could not change the index. this index " + index + " not exists")
	}
}

// AddPath adding a path if they not already exists
func (c *contxtConfigure) AddPath(path string) error {
	if c.PathExists(path) {
		return errors.New("the path is already stored in " + c.UsedV2Config.CurrentSet)
	}
	cfg, found := c.getCurrentConfig()
	if !found {
		return errors.New("error on reading current configuration")
	}
	if newIndex, err := c.getAutoInc(10); err == nil {
		cfg.Paths[newIndex] = WorkspaceInfoV2{Path: path}
		return nil
	} else {
		return err
	}

}

// create a numeric path index (as string) by checking the current index entries
func (c *contxtConfigure) getAutoInc(maxTrys int) (string, error) {
	cfg, found := c.getCurrentConfig()
	if !found {
		return "", errors.New("error on reading current configuration while detection next index")
	}
	checks := []string{}
	for i := 0; i <= maxTrys; i++ {
		for indx := range cfg.Paths {
			checks = append(checks, indx)
			i++
		}
		strI := strconv.Itoa(i)
		if !strSliceContains(checks, strI) {
			return strI, nil
		}
	}
	return "", errors.New("could not create an new index")
}

func strSliceContains(slice []string, search string) bool {
	for _, str := range slice {
		if str == search {
			return true
		}
	}
	return false
}

// RemovePath removes a path from current set.
// returns true is path was found and removed
func (c *contxtConfigure) RemovePath(path string) bool {
	cfg, found := c.getCurrentConfig()
	if !found {
		return false
	}
	for indx, p := range cfg.Paths {
		if p.Path == path {
			delete(cfg.Paths, indx)
			return true
		}
	}
	return false
}

// PathExists checks if this path is stored in current Workspace
func (c *contxtConfigure) PathExists(pathSearch string) bool {

	cfg, found := c.getCurrentConfig()
	if !found {
		return false
	}

	for _, path := range cfg.Paths {
		if pathSearch == path.Path {
			return true
		}
	}
	return false
}

// PathMeightPartOfWs checks if the path meight be a part of this workspace
func (c *contxtConfigure) PathMeightPartOfWs(pathSearch string) bool {
	cfg, found := c.getCurrentConfig()
	if !found {
		return false
	}
	for _, path := range cfg.Paths {
		if strings.Contains(pathSearch, path.Path) {
			return true
		}
	}
	return false
}
