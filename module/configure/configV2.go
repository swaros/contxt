package configure

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yamc"
)

var (
	CfgV1             *contxtConfigure = NewContxtConfig()
	USE_SPECIAL_DIR                    = true
	CONTEXT_DIR                        = "contxt"
	CONTXT_FILE                        = "contxtv2.yml"
	CFG               ConfigMetaV2     = ConfigMetaV2{}
	MIGRATION_ENABLED                  = true
)

type contxtConfigure struct {
	UsedV2Config      *ConfigMetaV2
	DefaultV2Yacl     *yacl.ConfigModel
	migrationRequired bool
}

func NewCfgV2(c *contxtConfigure) {
	c.DefaultV2Yacl = yacl.New(&CFG, yamc.NewYamlReader()).
		Init(func(strct *any) {
			CFG.Configs = make(map[string]ConfigurationV2)

		}, func(errCode int) error {
			// set flag that we may have to migrate from v1 to v2
			c.migrationRequired = true
			// if the target folder not exists, we will create them and print a message about this
			if errCode == yacl.ERROR_PATH_NOT_EXISTS {
				fmt.Println("configuration path not exists ", c.DefaultV2Yacl.GetConfigPath(), " try to create them")
				if err := os.MkdirAll(c.DefaultV2Yacl.GetConfigPath(), os.ModePerm); err != nil {
					panic("error while create configuration folder " + err.Error()) // at this point we should have panic
				}
			}
			return nil
		}).
		SetSubDirs(CONTEXT_DIR).
		SetSingleFile(CONTXT_FILE)
	// we can use this for testing to point to a relative path
	if USE_SPECIAL_DIR {
		c.DefaultV2Yacl.UseConfigDir()
	}

	if err := c.DefaultV2Yacl.Load(); err != nil {
		// errors depending not existing folders and files should already be handled without reporting as error
		// so this is something else and a reason for panic
		panic("error while reading configuration " + err.Error())

	}
	c.UsedV2Config = &CFG
}

func NewContxtConfig() *contxtConfigure {
	var cfgV1 Configuration
	//var cfgV2 ConfigMetaV2
	c := &contxtConfigure{}
	NewCfgV2(c)

	// if migration is required
	// this is the case if the new configuration file not exists
	// and the global migration flag is set
	if MIGRATION_ENABLED && c.migrationRequired {
		contxtCfg := yacl.New(&cfgV1, yamc.NewJsonReader()).
			SetSubDirs(".contxt").
			SupportMigrate(func(path string, cfg interface{}) {
				// this one contains the name of the current used config
				if strings.Contains(filepath.Clean(path), "contxt_current_config.json") {
					c.UsedV2Config.CurrentSet = cfgV1.CurrentSet
					return // just get out. we do not need something else from this file
				}

				// this one contains the name of the current used config
				if strings.Contains(filepath.Clean(path), "default_contxt_ws.json") {
					return // just get out. we ignore the whole configuation
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
				// we need to set the current index
				cfgEntrie.CurrentIndex = strconv.Itoa(cfgV1.LastIndex)
				// we make this already in the Init func
				if c.UsedV2Config.Configs == nil {
					fmt.Println("wtf. this should be initalized already")
					os.Exit(55)
				}
				c.UsedV2Config.Configs[cfgV1.CurrentSet] = cfgEntrie
			})
		if USE_SPECIAL_DIR {
			contxtCfg.UseHomeDir()
		}
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

// getConfig helper function to get a config by the name
func (c *contxtConfigure) getConfig(name string) (ConfigurationV2, bool) {
	if cfg, ok := c.UsedV2Config.Configs[name]; ok {
		return cfg, true
	}
	return ConfigurationV2{}, false
}

// getCurrentConfig helper function
func (c *contxtConfigure) getCurrentConfig() (ConfigurationV2, bool) {
	return c.getConfig(c.UsedV2Config.CurrentSet)
}

// helper funtion to change a named config
func (c *contxtConfigure) DoConfigChange(name string, worker func(cfg *ConfigurationV2)) {
	if cfg, ok := c.UsedV2Config.Configs[name]; ok {
		worker(&cfg)
	}
}

// helper function to change the current used config
func (c *contxtConfigure) DoCurrentConfigChange(worker func(cfg *ConfigurationV2)) {
	c.DoConfigChange(c.UsedV2Config.CurrentSet, worker)
}

// ClearPaths removes all paths
func (c *contxtConfigure) ClearPaths() {
	c.DoCurrentConfigChange(func(cfg *ConfigurationV2) {
		cfg.Paths = make(map[string]WorkspaceInfoV2)
		c.UpdateCurrentConfig(*cfg)
	})
	c.DefaultV2Yacl.Save()
}

// ListWorkSpaces returns all workspaces
func (c *contxtConfigure) ListWorkSpaces() []string {
	ws := []string{}
	for indx := range c.UsedV2Config.Configs {
		ws = append(ws, indx)
	}
	return ws
}

// ExecOnWorkSpaces callback on any workspace. as argument for the callback you get the name and the configuration
func (c *contxtConfigure) ExecOnWorkSpaces(callFn func(index string, cfg ConfigurationV2)) {
	for i, c := range c.UsedV2Config.Configs {
		callFn(i, c)
	}
}

// CurrentWorkspace returns the name of the current workspace
func (c *contxtConfigure) CurrentWorkspace() string {
	return c.UsedV2Config.CurrentSet
}

// ChangeWorkspace changes the current workspace and executes callbacks for the current workspace
// and afterwards for the new one.
// the configuration willbe save
func (c *contxtConfigure) ChangeWorkspace(workspace string, oldspace func(string) bool, newspace func(string)) error {
	// triggers execution of checking old Workspace
	canChange := oldspace(c.UsedV2Config.CurrentSet)
	if canChange {
		_, exists := c.getConfig(workspace)
		if !exists {
			return errors.New("workspace does not exists")
		}

		// change set name and save
		c.UsedV2Config.CurrentSet = workspace
		if err := c.DefaultV2Yacl.Save(); err != nil {
			return err
		}
		newspace(workspace) // execute any assigned newspace callback
		return nil
	} else {
		return errors.New("changing workspace failed")
	}
}

// RemoveWorkspace a workspace from the configuration
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

// HaveWorkSpace checks if the workspace exists
func (c *contxtConfigure) HaveWorkSpace(name string) bool {
	_, found := c.getConfig(name)
	return found
}

func (c *contxtConfigure) AddWorkSpace(name string, oldspace func(string) bool, newspace func(string)) error {
	clearname, nErr := systools.CheckForCleanString(name)
	if nErr != nil {
		return nErr
	}

	// if we have a cleared string that is not equal to the submitted, then we do not continue
	if name != clearname {
		return errors.New("the workspace name [" + name + "] is invalid")
	}

	if c.HaveWorkSpace(name) {
		return errors.New("the workspace [" + clearname + "] already exists")
	}

	c.UsedV2Config.Configs[clearname] = ConfigurationV2{Paths: map[string]WorkspaceInfoV2{}}
	return c.ChangeWorkspace(clearname, oldspace, newspace)
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

			if err := os.Chdir(current); err != nil {
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

// SetCurrentPathIndex sets the current path index
func (c *contxtConfigure) SetCurrentPathIndex(index string) error {
	cfg, found := c.getCurrentConfig()
	if !found {
		return errors.New("error while getting the current configuration")
	}
	if _, ok := cfg.Paths[index]; ok {
		cfg.CurrentIndex = index
		c.UpdateCurrentConfig(cfg)
		return nil
	}
	return errors.New("index does not exists")
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

func (c *contxtConfigure) UpdateCurrentConfig(updated ConfigurationV2) error {
	// weird lint. it did not see the usage of cfgElement in body, so i added a useless check of the current index
	if cfgElement, ok := c.UsedV2Config.Configs[c.UsedV2Config.CurrentSet]; ok && cfgElement.CurrentIndex != "fake-because-of-weird-lint" {
		cfgElement = updated
		c.UsedV2Config.Configs[c.UsedV2Config.CurrentSet] = cfgElement
		return nil
	} else {
		return errors.New("error change config on entry " + c.UsedV2Config.CurrentSet)
	}
}

func (c *contxtConfigure) ChangeActivePath(index string) error {
	cfg, found := c.getCurrentConfig()
	if !found {
		return errors.New("could not change the index. error while reading the configuration")
	}
	if _, ok := cfg.Paths[index]; ok {
		cfg.CurrentIndex = index
		return c.UpdateCurrentConfig(cfg)
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
		autoSet := (len(cfg.Paths) == 0)
		cfg.Paths[newIndex] = WorkspaceInfoV2{Path: path}
		// if this is the fist path, then we also set this as used index
		if autoSet {
			cfg.CurrentIndex = newIndex
		}
		return c.UpdateCurrentConfig(cfg)
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
		if filepath.Clean(pathSearch) == filepath.Clean(path.Path) {
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
