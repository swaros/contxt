package cmdhandle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/swaros/contxt/context/output"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/context/configure"
	"github.com/swaros/contxt/context/dirhandle"
)

func CheckUseConfig(externalUseCase string) (string, error) {
	GetLogger().WithField("usage", externalUseCase).Info("trying to solve usecase")
	sharedPath, err := configure.GetSharedPath(externalUseCase)
	path := ""
	if err == nil && sharedPath != "" {
		isThere, dirError := dirhandle.Exists(sharedPath)
		GetLogger().WithFields(logrus.Fields{"path": sharedPath, "exists": isThere, "err": dirError}).Info("using shared contxt tasks")
		if dirError != nil {
			fmt.Println(dirError)
		} else {
			if !isThere {
				// create dir first
				GetLogger().WithField("path", sharedPath).Info("shared directory not exists. try to checkout by git (github)")
				path = createUseByGit(externalUseCase, sharedPath)
				GetLogger().WithField("shared-path", path).Debug("shared usage")

			} else {
				path = sharedPath + "/source"
				exists, _ := dirhandle.Exists(path)
				if !exists {
					output.Error("USE Error", "shared usecase not exist and can not be downloaded", " ", path)
					os.Exit(10)
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
	return nil
}

func HandleUsecase(externalUseCase string) string {
	path, _ := CheckUseConfig(externalUseCase)
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
		fmt.Println(output.MessageCln(" remote:", output.ForeLightBlue, " ", config.Repositiory))
		updateGitRepo(config, true, fullPath)

	} else {
		fmt.Println(output.MessageCln(" local shared:", output.ForeYellow, " ", fullPath, output.ForeDarkGrey, "(not updatable. ignored)"))
	}
}

func ListUseCases(fullPath bool) ([]string, error) {
	var sharedDirs []string
	//sep := fmt.Sprintf("%c", os.PathSeparator)
	sharedPath, perr := configure.GetSharedPath("")
	if perr == nil {
		errWalk := filepath.Walk(sharedPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				//var extension = filepath.Ext(path)
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

func getUseInfo(usecase, pathTouse string) (string, string) {
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
		fmt.Print(output.MessageCln(" Reference:", output.ForeLightBlue, " ", config.Reference))
		fmt.Print(output.MessageCln(" Current:", output.ForeLightBlue, " ", config.HashUsed))
		returnBool := false
		checkGitVersionInfo(config.Repositiory, func(hash, reference string) {
			if reference == config.Reference {
				fmt.Print(output.MessageCln(output.ForeLightGreen, "[EXISTS]"))
				if hash == config.HashUsed {
					fmt.Print(output.MessageCln(output.ForeLightGreen, " [up to date]"))
				} else {
					fmt.Print(output.MessageCln(output.ForeYellow, " [update found]"))
					if doUpdate {
						gCode := executeGitUpdate(workDir + "/source")
						if gCode == ExitOk {
							config.HashUsed = hash
							writeGitConfig(workDir+"/version.json", config)
							returnBool = true
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
	gitCmd := "git fetch"
	exitCode, _, _ := ExecuteScriptLine("bash", []string{"-c"}, gitCmd, func(feed string) bool {
		fmt.Println(output.MessageCln("\tgit: ", output.ForeLightYellow, feed))
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
	internalExitCode, cmdError, err := ExecuteScriptLine("bash", []string{"-c"}, gitCmd, func(feed string) bool {
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
	/*
		parts := strings.Split(usecase, "@")
		version := "refs/heads/main"
		if len(parts) > 1 {
			usecase = parts[0]
			version = "refs/tags/" + parts[1]
		}*/
	usecase, version := getUseInfo(usecase, pathTouse)
	GetLogger().WithFields(logrus.Fields{"use": usecase, "path": pathTouse, "version": version}).Debug("Import Usecase")
	path := ""
	gitCmd := "git ls-remote --refs https://github.com/" + usecase
	internalExitCode, cmdError, _ := ExecuteScriptLine("bash", []string{"-c"}, gitCmd, func(feed string) bool {
		gitInfo := strings.Split(feed, "\t")
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
	if internalExitCode != ExitOk {
		// git info was failing. so we did not create anything right now by using git
		// so now we have to check if this is a local repository
		GetLogger().WithFields(logrus.Fields{
			"exitCode":      internalExitCode,
			"cmd-exit-code": cmdError,
		}).Warning("failed get version info from git")
		exists, _ := dirhandle.Exists(pathTouse)
		if exists {
			existsSource, _ := dirhandle.Exists(pathTouse + "/.source")
			if existsSource {
				return pathTouse + "/.source"
			}
		}
		GetLogger().WithField("path", pathTouse).Fatal("Local Usage folder not exists (+ ./source)")
		os.Exit(internalExitCode)
	}
	return path
}

func getRepoConfig(pathTouse string) (bool, configure.GitVersionInfo, error) {
	hashChk, hashError := dirhandle.Exists(pathTouse + "/version.json")
	var versionConf configure.GitVersionInfo
	if hashError != nil {
		return false, versionConf, hashError
	} else if hashChk {
		versionConf = loadGitConfig(pathTouse+"/version.json", versionConf)
		return true, versionConf, nil
	}
	GetLogger().WithField("path", pathTouse).Warning("no version info. seems to be a local shared.")
	return false, versionConf, nil
}

func getOrCreateRepoConfig(ref, hash, usecase, pathTouse string) (configure.GitVersionInfo, error) {
	hashChk, hashError := dirhandle.Exists(pathTouse + "/version.json")
	var versionConf configure.GitVersionInfo
	if hashError != nil {
		log.Fatal(hashError)
		return versionConf, hashError
	} else if !hashChk {
		versionConf.Repositiory = usecase
		versionConf.HashUsed = hash
		versionConf.Reference = ref
		writeGitConfig(pathTouse+"/version.json", versionConf)
		GetLogger().WithField("config", versionConf).Debug("Create new Config")
	} else {
		versionConf = loadGitConfig(pathTouse+"/version.json", versionConf)
		GetLogger().WithField("config", versionConf).Debug("Using existing Config")
	}
	return versionConf, nil
}

func takeCareAboutRepo(pathTouse string, config configure.GitVersionInfo) configure.GitVersionInfo {
	exists, _ := dirhandle.Exists(pathTouse + "/source")
	// source folder not exists
	if !exists {
		// no repository info exists
		if config.Repositiory != "" {

			// check if the useage folder exists and create them if not
			createSharedUsageDir(pathTouse)

			gitCmd := "git clone https://github.com/" + config.Repositiory + ".git " + pathTouse + "/source"
			GetLogger().WithField("cmd", gitCmd).Info("using git to create ne checkout from repo")
			codeInt, codeCmd, err := ExecuteScriptLine("bash", []string{"-c"}, gitCmd, func(feed string) bool {
				fmt.Println(feed)
				return true
			}, func(process *os.Process) {
				pidStr := fmt.Sprintf("%d", process.Pid)
				GetLogger().WithFields(logrus.Fields{"process": pidStr}).Debug("git process id")
			})

			GetLogger().WithFields(logrus.Fields{"codeInternal": codeInt, "code": codeCmd, "err": err}).Debug("git execution result")
		}
	}
	config.Path = pathTouse + "/source"
	return config
}

func writeGitConfig(path string, config configure.GitVersionInfo) {
	b, _ := json.MarshalIndent(config, "", " ")
	err := ioutil.WriteFile(path, b, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func loadGitConfig(path string, config configure.GitVersionInfo) configure.GitVersionInfo {

	file, _ := os.Open(path)
	defer file.Close()
	decoder := json.NewDecoder(file)

	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
	}
	return config

}
