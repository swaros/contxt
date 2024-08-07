// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

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
	cfgV1                *contxtConfigure                  // global config
	USE_SPECIAL_DIR                       = true           // if true, we will use some of the special dirs like user home or other. defined in the config model
	CONTEXT_DIR                           = "contxt"       // default directory name for the config files
	CONTXT_FILE                           = "contxtv2.yml" // default file name for the config file
	CFG                  ConfigMetaV2     = ConfigMetaV2{} // the config model
	MIGRATION_ENABLED                     = true           // if true, we will try to migrate from v1 to v2
	CONFIG_PATH_CALLBACK GetPathCallback  = nil            // if set, we will use this callback to get the absolte path to the config file
)

type GetPathCallback func() string

type contxtConfigure struct {
	UsedV2Config      *ConfigMetaV2
	DefaultV2Yacl     *yacl.ConfigModel
	migrationRequired bool
}

func GetGlobalConfig() *contxtConfigure {
	if cfgV1 == nil {
		cfgV1 = NewContxtConfig()
	}
	return cfgV1
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

	// if an callback is set, we will use this to get the path
	// what will also forces the absolute path usage
	if CONFIG_PATH_CALLBACK != nil {
		c.DefaultV2Yacl.SetFileAndPathsByFullFilePath(CONFIG_PATH_CALLBACK())

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
	c.ResetConfig()
	return nil
}

// Resets the global config so it forces a reload
func (c *contxtConfigure) ResetConfig() {
	cfgV1 = nil
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

// ListPaths returns all paths as array
func (c *contxtConfigure) ListPaths() []string {
	paths := []string{}
	c.PathWorkerNoCd(func(index, path string) {
		paths = append(paths, path)
	})
	return paths
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

// ChangeWorkspaceNotSaved changes the current workspace without saving the configuration.
// it also do not execute any callbacks while changing the workspace
func (c *contxtConfigure) ChangeWorkspaceNotSaved(workspace string) error {
	_, exists := c.getConfig(workspace)
	if !exists {
		return errors.New("workspace does not exists")
	}
	// change the current workspace
	c.UsedV2Config.CurrentSet = workspace
	return nil

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

	// we need at least 3 chars
	if len(clearname) < 3 {
		return errors.New("the workspace name [" + name + "] too short (min 3 chars)")
	}

	// allowed length is 125 chars
	if len(clearname) > 125 {
		return errors.New("the workspace name [" + name + "] too long. (max 125 chars)")
	}

	// check if the workspace already exists
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
			return errors.New("no paths actually stored in workspace " + c.UsedV2Config.CurrentSet)
		}
		prepmap := make(map[string]any)
		for index, path := range cfg.Paths {
			prepmap[index] = path
		}
		var errorWhileLoop error
		systools.MapRangeSortedFn(prepmap, func(index string, path any) {
			localpath := path.(WorkspaceInfoV2)
			if err := os.Chdir(localpath.Path); err == nil {
				callbackInDirectory(index, localpath.Path)
			} else {
				errorWhileLoop = err
				return
			}
			if err := os.Chdir(current); err != nil {
				errorWhileLoop = err
				return
			}
		})
		if errorWhileLoop != nil {
			return errorWhileLoop
		}

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
	start := len(cfg.Paths)
	// we will try to find an index that is not used
	// and we add the amount of paths to the maxTrys
	for i := 0; i <= maxTrys+start; i++ {
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
