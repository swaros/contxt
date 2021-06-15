package cmdhandle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

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
				GetLogger().Info("shared directory not exists. try to create them")
				err := os.MkdirAll(sharedPath, os.ModePerm)
				if err != nil {
					log.Fatal(err)
					return "", err
				}
				path = createUseByGit(externalUseCase, sharedPath)

			} else {
				path = sharedPath + "/source"
			}
		}
	}
	return path, nil
}

func HandleUsecase(externalUseCase string) string {
	path, _ := CheckUseConfig(externalUseCase)
	return path
}

func createUseByGit(usecase, pathTouse string) string {

	parts := strings.Split(usecase, "@")
	version := "refs/heads/main"
	if len(parts) > 1 {
		usecase = parts[0]
		version = "refs/tags/" + parts[1]
	}
	path := ""
	gitCmd := "git ls-remote --refs https://github.com/" + usecase
	ExecuteScriptLine("bash", []string{"-c"}, gitCmd, func(feed string) bool {
		gitInfo := strings.Split(feed, "\t")
		if len(gitInfo) >= 2 && gitInfo[1] == version {
			GetLogger().WithFields(logrus.Fields{"git-info": gitInfo, "cnt": len(gitInfo)}).Info("found matching version")
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
	return path
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
	if !exists {
		if config.Repositiory != "" {
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
