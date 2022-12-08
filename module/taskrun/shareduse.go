// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Licensed under the MIT License
//
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
package taskrun

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/manout"
)

var config_file string = "version.conf"

// CheckOrCreateUseConfig get a usecase like swaros/ctx-git and checks
// if a local copy of them exists.
// if they not exists it creates the local directoy and uses git to
// clone the content.
// afterwards it writes a version.conf in the forlder above of content
// and stores the current hashes
func CheckOrCreateUseConfig(externalUseCase string) (string, error) {
	GetLogger().WithField("usage", externalUseCase).Info("trying to solve usecase")
	path := ""                                                  // just as default
	sharedPath, err := configure.GetSharedPath(externalUseCase) // get the main path for shared content
	if err == nil && sharedPath != "" {                         // no error and not an empty path
		isThere, dirError := dirhandle.Exists(sharedPath) // do we have the main shared directory?
		GetLogger().WithFields(logrus.Fields{"path": sharedPath, "exists": isThere, "err": dirError}).Info("using shared contxt tasks")
		if dirError != nil { // this is NOT related to not exists. it is an error while checking if the path exists
			log.Fatal(dirError)
		} else {
			if !isThere { // directory not exists
				GetLogger().WithField("path", sharedPath).Info("shared directory not exists. try to checkout by git (github)")
				path = createUseByGit(externalUseCase, sharedPath) // create dirs and checkout content if possible. fit the path also

			} else { // directory exists
				path = getSourcePath(sharedPath)
				exists, _ := dirhandle.Exists(path)
				if !exists {
					manout.Error("USE Error", "shared usecase not exist and can not be downloaded", " ", path)
					systools.Exit(systools.ErrorBySystem)
				}
				GetLogger().WithField("shared-path", path).Debug("use existing shared path")
			}
		}
	}
	return path, nil
}

func createSharedUsageDir(sharedPath string) error {
	exists, _ := dirhandle.Exists(sharedPath)
	if !exists {
		// create dir
		GetLogger().WithField("path", sharedPath).Info("shared directory not exists. try to create them")
		err := os.MkdirAll(sharedPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
			return err
		}
	}
	GetLogger().WithField("path", sharedPath).Info("shared directory exists already")
	return nil
}

func HandleUsecase(externalUseCase string) string {
	path, _ := CheckOrCreateUseConfig(externalUseCase)
	return path
}

func StripContxtUseDir(path string) string {
	sep := fmt.Sprintf("%c", os.PathSeparator)
	newpath := strings.TrimSuffix(path, sep)

	parts := strings.Split(newpath, sep)
	cleanDir := ""
	if len(parts) > 1 && parts[len(parts)-1] == "source" {
		parts = parts[:len(parts)-1]
	}
	for _, subpath := range parts {
		if subpath != "" {
			cleanDir = cleanDir + sep + subpath
		}

	}
	return cleanDir
}

func UpdateUseCase(fullPath string) {
	//usecase, version := getUseInfo("", fullPath)
	exists, config, _ := getRepoConfig(fullPath)
	if exists {
		GetLogger().WithFields(logrus.Fields{"config": config}).Debug("version info")
		fmt.Println(manout.MessageCln(" remote:", manout.ForeLightBlue, " ", config.Repositiory))
		updateGitRepo(config, true, fullPath)

	} else {
		fmt.Println(manout.MessageCln(" local shared:", manout.ForeYellow, " ", fullPath, manout.ForeDarkGrey, "(not updatable. ignored)"))
	}
}

func ListUseCases(fullPath bool) ([]string, error) {
	var sharedDirs []string
	sharedPath, perr := configure.GetSharedPath("")
	if perr == nil {
		errWalk := filepath.Walk(sharedPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				var basename = filepath.Base(path)
				var directory = filepath.Dir(path)

				if basename == ".contxt.yml" {
					if fullPath {
						sharedDirs = append(sharedDirs, StripContxtUseDir(directory))
					} else {
						releative := strings.Replace(StripContxtUseDir(directory), sharedPath, "", 1)
						sharedDirs = append(sharedDirs, releative)
					}
				}
			}
			return nil
		})
		return sharedDirs, errWalk
	}
	return sharedDirs, perr
}

func GetUseInfo(usecase, _ string) (string, string) {
	parts := strings.Split(usecase, "@")
	version := "refs/heads/main"
	if len(parts) > 1 {
		usecase = parts[0]
		version = "refs/tags/" + parts[1]
	}
	return usecase, version
}

func updateGitRepo(config configure.GitVersionInfo, doUpdate bool, workDir string) bool {
	if config.Repositiory != "" {
		fmt.Print(manout.MessageCln(" Reference:", manout.ForeLightBlue, " ", config.Reference))
		fmt.Print(manout.MessageCln(" Current:", manout.ForeLightBlue, " ", config.HashUsed))
		returnBool := false
		checkGitVersionInfo(config.Repositiory, func(hash, reference string) {
			if reference == config.Reference {
				fmt.Print(manout.MessageCln(manout.ForeLightGreen, "[EXISTS]"))
				if hash == config.HashUsed {
					fmt.Print(manout.MessageCln(manout.ForeLightGreen, " [up to date]"))
				} else {
					fmt.Print(manout.MessageCln(manout.ForeYellow, " [update found]"))
					if doUpdate {
						gCode := executeGitUpdate(getSourcePath(workDir))
						if gCode == systools.ExitOk {
							config.HashUsed = hash
							if werr := writeGitConfig(workDir+"/"+config_file, config); werr != nil {
								manout.Error("unable to create version info", werr)
								returnBool = false
							} else {
								returnBool = true
							}
						}
					}
				}
			}
		})
		fmt.Println(".")
		return returnBool
	}
	return false
}

func executeGitUpdate(path string) int {
	currentDir, _ := dirhandle.Current()
	os.Chdir(path)
	gitCmd := "git pull"
	exec, args := GetExecDefaults()
	exitCode, _, _ := ExecuteScriptLine(exec, args, gitCmd, func(feed string, e error) bool {
		fmt.Println(manout.MessageCln("\tgit: ", manout.ForeLightYellow, feed))
		return true
	}, func(process *os.Process) {
		pidStr := fmt.Sprintf("%d", process.Pid)
		GetLogger().WithFields(logrus.Fields{"process": pidStr}).Debug("git process id")
	})
	os.Chdir(currentDir)
	return exitCode
}

// first argument is the hash and the second one is the version
func checkGitVersionInfo(usecase string, callback func(string, string)) (int, int, error) {
	gitCmd := "git ls-remote --refs https://github.com/" + usecase
	exec, args := GetExecDefaults()
	internalExitCode, cmdError, err := ExecuteScriptLine(exec, args, gitCmd, func(feed string, e error) bool {
		gitInfo := strings.Split(feed, "\t")
		if len(gitInfo) >= 2 {
			callback(gitInfo[0], gitInfo[1])
		}
		return true
	}, func(process *os.Process) {
		pidStr := fmt.Sprintf("%d", process.Pid)
		GetLogger().WithFields(logrus.Fields{"process": pidStr}).Debug("git process id")
	})
	return internalExitCode, cmdError, err
}

func createUseByGit(usecase, pathTouse string) string {
	usecase, version := GetUseInfo(usecase, pathTouse) // get needed git ref and usecase by the requested usage (like from swaros/ctx-gt@v0.0.1)
	GetLogger().WithFields(logrus.Fields{"use": usecase, "path": pathTouse, "version": version}).Debug("Import Usecase")
	path := ""
	gitCmd := "git ls-remote --refs https://github.com/" + usecase
	exec, args := GetExecDefaults()
	var gitInfo []string
	internalExitCode, cmdError, _ := ExecuteScriptLine(exec, args, gitCmd, func(feed string, e error) bool {
		gitInfo = strings.Split(feed, "\t")
		if len(gitInfo) >= 2 && gitInfo[1] == version {
			GetLogger().WithFields(logrus.Fields{"git-info": gitInfo, "cnt": len(gitInfo)}).Debug("found matching version")
			cfg, versionErr := getOrCreateRepoConfig(gitInfo[1], gitInfo[0], usecase, pathTouse)
			if versionErr == nil {
				cfg = takeCareAboutRepo(pathTouse, cfg)
				path = cfg.Path
			}
		}
		return true
	}, func(process *os.Process) {
		pidStr := fmt.Sprintf("%d", process.Pid)
		GetLogger().WithFields(logrus.Fields{"process": pidStr}).Debug("git process id")
	})
	if internalExitCode != systools.ExitOk {
		// git info was failing. so we did not create anything right now by using git
		// so now we have to check if this is a local repository
		GetLogger().WithFields(logrus.Fields{
			"exitCode":      internalExitCode,
			"cmd-exit-code": cmdError,
		}).Warning("failed get version info from git")
		exists, _ := dirhandle.Exists(pathTouse)
		if exists {
			existsSource, _ := dirhandle.Exists(getSourcePath(pathTouse))
			if existsSource {
				return getSourcePath(pathTouse)
			}
		}
		GetLogger().WithField("path", pathTouse).Fatal("Local Usage folder not exists (+ ./source)")
		systools.Exit(internalExitCode)
	}
	return path
}

func getRepoConfig(pathTouse string) (bool, configure.GitVersionInfo, error) {
	hashChk, hashError := dirhandle.Exists(getVersionOsPath(pathTouse))
	var versionConf configure.GitVersionInfo
	if hashError != nil {
		return false, versionConf, hashError
	} else if hashChk {
		versionConf, err := loadGitConfig(getVersionOsPath(pathTouse), versionConf)
		return err == nil, versionConf, err
	}
	GetLogger().WithField("path", pathTouse).Warning("no version info. seems to be a local shared.")
	return false, versionConf, nil
}

func getSourcePath(pathTouse string) string {
	return fmt.Sprintf("%s%s%s", pathTouse, string(os.PathSeparator), "source")
}

func getVersionOsPath(pathTouse string) string {
	return fmt.Sprintf("%s%s%s", pathTouse, string(os.PathSeparator), config_file)
}

func getOrCreateRepoConfig(ref, hash, usecase, pathTouse string) (configure.GitVersionInfo, error) {
	var versionConf configure.GitVersionInfo
	versionFilename := getVersionOsPath(pathTouse)

	// check if the useage folder exists and create them if not
	if pathWErr := createSharedUsageDir(pathTouse); pathWErr != nil {
		manout.Error("error while create directory:", pathWErr)
		return versionConf, pathWErr
	}

	hashChk, hashError := dirhandle.Exists(versionFilename)
	if hashError != nil {
		manout.Error("error while checking directory:", hashError)
		return versionConf, hashError
	} else if !hashChk {

		versionConf.Repositiory = usecase
		versionConf.HashUsed = hash
		versionConf.Reference = ref

		GetLogger().WithField("file", versionFilename).Info("Try to create version info")
		if werr := writeGitConfig(versionFilename, versionConf); werr != nil {
			GetLogger().WithField("file", versionFilename).Error("error by create version info: ", werr)
			manout.Error("unable to create version info ", versionFilename, werr)
			return versionConf, werr
		}

		GetLogger().WithField("config", versionConf).Debug("Create new Config")
	} else {
		versionConf, vErr := loadGitConfig(versionFilename, versionConf)
		GetLogger().WithField("config", versionConf).Debug("Using existing Config")
		return versionConf, vErr
	}
	return versionConf, nil
}

func takeCareAboutRepo(pathTouse string, config configure.GitVersionInfo) configure.GitVersionInfo {
	exists, _ := dirhandle.Exists(getSourcePath(pathTouse))
	if !exists { // source folder not exists
		if config.Repositiory != "" { // no repository info exists
			createSharedUsageDir(pathTouse) // check if the usage folder exists and create them if not
			gitCmd := "git clone https://github.com/" + config.Repositiory + ".git " + getSourcePath(pathTouse)
			GetLogger().WithField("cmd", gitCmd).Info("using git to create new checkout from repo")
			exec, args := GetExecDefaults()
			codeInt, codeCmd, err := ExecuteScriptLine(exec, args, gitCmd, func(feed string, e error) bool {
				fmt.Println(feed)
				return true
			}, func(process *os.Process) {
				pidStr := fmt.Sprintf("%d", process.Pid)
				GetLogger().WithFields(logrus.Fields{"process": pidStr}).Debug("git process id")
			})

			GetLogger().WithFields(logrus.Fields{"codeInternal": codeInt, "code": codeCmd, "err": err}).Debug("git execution result")
		} else {
			GetLogger().WithFields(logrus.Fields{"folder": pathTouse}).Debug("source folder exists, but no version info")
		}
	}
	config.Path = getSourcePath(pathTouse)
	return config
}

func writeGitConfig(path string, config configure.GitVersionInfo) error {
	b, _ := json.MarshalIndent(config, "", " ")
	if err := ioutil.WriteFile(path, b, 0644); err != nil {
		GetLogger().Error("can not create file ", path, " ", err)
		return err
	}
	return nil
}

func loadGitConfig(path string, config configure.GitVersionInfo) (configure.GitVersionInfo, error) {

	file, _ := os.Open(path)
	defer file.Close()
	decoder := json.NewDecoder(file)

	err := decoder.Decode(&config)
	return config, err

}
