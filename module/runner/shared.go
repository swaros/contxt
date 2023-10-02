package runner

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
	"github.com/swaros/manout"
)

const (
	DefaultSubPath     = "/.contxt/shared/"
	DefaultVersionConf = "version.conf"
)

type SharedHelper struct {
	basePath       string
	defaultSubPath string
	versionConf    string
	logger         mimiclog.Logger
}

// NewSharedHelper returns a new instance of the SharedHelper depending on the user home dir
func NewSharedHelper() *SharedHelper {
	if path, err := os.UserHomeDir(); err != nil {
		panic(err)
	} else {
		return NewSharedHelperWithPath(path)
	}
}

// NewSharedHelperWithPath returns a new instance of the SharedHelper depending on the given path
func NewSharedHelperWithPath(basePath string) *SharedHelper {
	return &SharedHelper{basePath, DefaultSubPath, DefaultVersionConf, mimiclog.NewNullLogger()}
}

func (sh *SharedHelper) SetLogger(logger mimiclog.Logger) {
	sh.logger = logger
}

// GetBasePath returns the base path of the shared folder
func (sh *SharedHelper) GetBasePath() string {
	return sh.basePath
}

// GetSharedPath returns the full path of the given shared name
func (sh *SharedHelper) GetSharedPath(sharedName string) string {
	fileName := systools.SanitizeFilename(sharedName, true) // make sure we have an valid filename
	return filepath.Clean(filepath.FromSlash(sh.basePath + sh.defaultSubPath + fileName))
}

// CheckOrCreateUseConfig get a usecase like swaros/ctx-git and checks
// if a local copy of them exists.
// if they not exists it creates the local directoy and uses git to
// clone the content.
// afterwards it writes a version.conf, in the forlder above of content,
// and stores the current hashes
func (sh *SharedHelper) CheckOrCreateUseConfig(externalUseCase string) (string, error) {
	sh.logger.Info("trying to solve usecase", externalUseCase)
	path := ""                                      // just as default
	sharedPath := sh.GetSharedPath(externalUseCase) // get the main path for shared content
	if sharedPath != "" {                           // no error and not an empty path
		isThere, dirError := dirhandle.Exists(sharedPath) // do we have the main shared directory?
		sh.logger.Info("using shared contxt tasks", sharedPath)
		if dirError != nil { // this is NOT related to not exists. it is an error while checking if the path exists
			return "", dirError
		} else {
			if !isThere { // directory not exists
				sh.logger.Info("shared directory not exists. try to checkout by git (github)")
				path = sh.createUseByGit(externalUseCase, sharedPath) // create dirs and checkout content if possible. fit the path also

			} else { // directory exists
				path = sh.getSourcePath(sharedPath)
				exists, _ := dirhandle.Exists(path)
				if !exists {
					manout.Error("USE Error", "shared usecase not exist and can not be downloaded", " ", path)
					systools.Exit(systools.ErrorBySystem)
				}
				sh.logger.Debug("shared directory exists. use them", path)
			}
		}
	}
	return path, nil
}

func (sh *SharedHelper) createUseByGit(usecase, pathTouse string) string {
	usecase, version := sh.GetUseInfo(usecase, pathTouse) // get needed git ref and usecase by the requested usage (like from swaros/ctx-gt@v0.0.1)
	sh.logger.Info("trying to checkout", usecase, "by git.", pathTouse, " version:", version)
	path := ""
	gitCmd := "git ls-remote --refs https://github.com/" + usecase

	var gitInfo []string
	shellRunner := tasks.GetShellRunner()
	internalExitCode, cmdError, _ := shellRunner.Exec(gitCmd, func(feed string, e error) bool {
		gitInfo = strings.Split(feed, "\t")
		if len(gitInfo) >= 2 && gitInfo[1] == version {
			sh.logger.Debug("found matching version")
			cfg, versionErr := sh.getOrCreateRepoConfig(gitInfo[1], gitInfo[0], usecase, pathTouse)
			if versionErr == nil {
				cfg = sh.takeCareAboutRepo(pathTouse, cfg)
				path = cfg.Path
			}
		}
		return true
	}, func(process *os.Process) {
		pidStr := fmt.Sprintf("%d", process.Pid)
		sh.logger.Debug("git process id", pidStr)
	})

	if internalExitCode != systools.ExitOk {
		// git info was failing. so we did not create anything right now by using git
		// so now we have to check if this is a local repository
		sh.logger.Warn("failed get version info from git", internalExitCode, cmdError)
		exists, _ := dirhandle.Exists(pathTouse)
		if exists {
			existsSource, _ := dirhandle.Exists(sh.getSourcePath(pathTouse))
			if existsSource {
				return sh.getSourcePath(pathTouse)
			}
		}

		sh.logger.Critical("Local Usage folder not exists (+ ./source)", pathTouse)
		systools.Exit(internalExitCode)
	}
	return path
}

func (sh *SharedHelper) GetUseInfo(usecase, _ string) (string, string) {
	parts := strings.Split(usecase, "@")
	version := "refs/heads/main"
	if len(parts) > 1 {
		usecase = parts[0]
		version = "refs/tags/" + parts[1]
	}
	return usecase, version
}

func (sh *SharedHelper) GetSharedPathForUseCase(usecase string) string {
	return sh.GetSharedPath(usecase)
}

func (sh *SharedHelper) getSourcePath(pathTouse string) string {
	return fmt.Sprintf("%s%s%s", pathTouse, string(os.PathSeparator), "source")
}

func (sh *SharedHelper) getVersionOsPath(pathTouse string) string {
	return fmt.Sprintf("%s%s%s", pathTouse, string(os.PathSeparator), sh.versionConf)
}

func (sh *SharedHelper) getOrCreateRepoConfig(ref, hash, usecase, pathTouse string) (configure.GitVersionInfo, error) {
	var versionConf configure.GitVersionInfo
	versionFilename := sh.getVersionOsPath(pathTouse)

	// check if the useage folder exists and create them if not
	if pathWErr := sh.createSharedUsageDir(pathTouse); pathWErr != nil {
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

		sh.logger.Info("try to create version info", versionFilename)
		if werr := sh.writeGitConfig(versionFilename, versionConf); werr != nil {
			sh.logger.Error("unable to create version info ", versionFilename, werr)
			return versionConf, werr
		}
		sh.logger.Debug("created version info", versionConf)
	} else {
		versionConf, vErr := sh.loadGitConfig(versionFilename, versionConf)
		sh.logger.Debug("loaded version info", versionConf)
		return versionConf, vErr
	}
	return versionConf, nil
}

func (sh *SharedHelper) createSharedUsageDir(sharedPath string) error {
	exists, _ := dirhandle.Exists(sharedPath)
	if !exists {
		// create dir
		sh.logger.Info("shared directory not exists. try to create them", sharedPath)
		err := os.MkdirAll(sharedPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
			return err
		}
	}
	sh.logger.Info("shared directory exists already", sharedPath)
	return nil
}

func (sh *SharedHelper) HandleUsecase(externalUseCase string) string {
	path, _ := sh.CheckOrCreateUseConfig(externalUseCase)
	return path
}

func (sh *SharedHelper) StripContxtUseDir(path string) string {
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

func (sh *SharedHelper) UpdateUseCase(fullPath string) {
	//usecase, version := getUseInfo("", fullPath)
	exists, config, _ := sh.getRepoConfig(fullPath)
	if exists {
		sh.logger.Debug("update shared usecase", fullPath, config)
		fmt.Println(manout.MessageCln(" remote:", manout.ForeLightBlue, " ", config.Repositiory))
		sh.updateGitRepo(config, true, fullPath)

	} else {
		fmt.Println(manout.MessageCln(" local shared:", manout.ForeYellow, " ", fullPath, manout.ForeDarkGrey, "(not updatable. ignored)"))
	}
}

func (sh *SharedHelper) ListUseCases(fullPath bool) ([]string, error) {
	var sharedDirs []string
	sharedPath := sh.GetSharedPath("")

	errWalk := filepath.Walk(sharedPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			var basename = filepath.Base(path)
			var directory = filepath.Dir(path)

			if basename == ".contxt.yml" {
				if fullPath {
					sharedDirs = append(sharedDirs, sh.StripContxtUseDir(directory))
				} else {
					releative := strings.Replace(sh.StripContxtUseDir(directory), sharedPath, "", 1)
					sharedDirs = append(sharedDirs, releative)
				}
			}
		}
		return nil
	})
	return sharedDirs, errWalk

}

func (sh *SharedHelper) getRepoConfig(pathTouse string) (bool, configure.GitVersionInfo, error) {
	hashChk, hashError := dirhandle.Exists(sh.getVersionOsPath(pathTouse))
	var versionConf configure.GitVersionInfo
	if hashError != nil {
		return false, versionConf, hashError
	} else if hashChk {
		versionConf, err := sh.loadGitConfig(sh.getVersionOsPath(pathTouse), versionConf)
		return err == nil, versionConf, err
	}
	sh.logger.Warn("no version info. seems to be a local shared.", pathTouse)
	return false, versionConf, nil
}

func (sh *SharedHelper) loadGitConfig(path string, config configure.GitVersionInfo) (configure.GitVersionInfo, error) {

	file, _ := os.Open(path)
	defer file.Close()
	decoder := json.NewDecoder(file)

	err := decoder.Decode(&config)
	return config, err

}

func (sh *SharedHelper) updateGitRepo(config configure.GitVersionInfo, doUpdate bool, workDir string) bool {
	if config.Repositiory != "" {
		fmt.Print(manout.MessageCln(" Reference:", manout.ForeLightBlue, " ", config.Reference))
		fmt.Print(manout.MessageCln(" Current:", manout.ForeLightBlue, " ", config.HashUsed))
		returnBool := false
		sh.checkGitVersionInfo(config.Repositiory, func(hash, reference string) {
			if reference == config.Reference {
				fmt.Print(manout.MessageCln(manout.ForeLightGreen, "[EXISTS]"))
				if hash == config.HashUsed {
					fmt.Print(manout.MessageCln(manout.ForeLightGreen, " [up to date]"))
				} else {
					fmt.Print(manout.MessageCln(manout.ForeYellow, " [update found]"))
					if doUpdate {
						gCode := sh.executeGitUpdate(sh.getSourcePath(workDir))
						if gCode == systools.ExitOk {
							config.HashUsed = hash
							if werr := sh.writeGitConfig(workDir+"/"+sh.versionConf, config); werr != nil {
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

func (sh *SharedHelper) checkGitVersionInfo(usecase string, callback func(string, string)) (int, int, error) {
	gitCmd := "git ls-remote --refs https://github.com/" + usecase
	shellRunner := tasks.GetShellRunner()
	internalExitCode, cmdError, err := shellRunner.Exec(gitCmd, func(feed string, e error) bool {
		gitInfo := strings.Split(feed, "\t")
		if len(gitInfo) >= 2 {
			callback(gitInfo[0], gitInfo[1])
		}
		return true
	}, func(process *os.Process) {
		pidStr := fmt.Sprintf("%d", process.Pid)
		sh.logger.Debug("git process id", pidStr)
	})
	return internalExitCode, cmdError, err
}

func (sh *SharedHelper) executeGitUpdate(path string) int {
	currentDir, _ := dirhandle.Current()
	os.Chdir(path)
	gitCmd := "git pull"

	shellRunner := tasks.GetShellRunner()
	exitCode, _, _ := shellRunner.Exec(gitCmd, func(feed string, e error) bool {
		fmt.Println(manout.MessageCln("\tgit: ", manout.ForeLightYellow, feed))
		return true
	}, func(process *os.Process) {
		pidStr := fmt.Sprintf("%d", process.Pid)
		sh.logger.Debug("git process id", pidStr)
	})
	os.Chdir(currentDir)
	return exitCode
}

func (sh *SharedHelper) writeGitConfig(path string, config configure.GitVersionInfo) error {
	b, _ := json.MarshalIndent(config, "", " ")
	if err := os.WriteFile(path, b, 0644); err != nil {
		sh.logger.Error("can not create file ", path, " ", err)
		return err
	}
	return nil
}

func (sh *SharedHelper) takeCareAboutRepo(pathTouse string, config configure.GitVersionInfo) configure.GitVersionInfo {
	exists, _ := dirhandle.Exists(sh.getSourcePath(pathTouse))
	if !exists { // source folder not exists
		if config.Repositiory != "" { // no repository info exists
			sh.createSharedUsageDir(pathTouse) // check if the usage folder exists and create them if not
			gitCmd := "git clone https://github.com/" + config.Repositiory + ".git " + sh.getSourcePath(pathTouse)
			sh.logger.Info("using git to create new checkout from repo", gitCmd)
			shellRunner := tasks.GetShellRunner()
			codeInt, codeCmd, err := shellRunner.Exec(gitCmd, func(feed string, e error) bool {
				fmt.Println(manout.MessageCln("\tgit: ", manout.ForeLightYellow, feed))
				return true
			}, func(process *os.Process) {
				pidStr := fmt.Sprintf("%d", process.Pid)
				sh.logger.Debug("git process id", pidStr)
			})
			sh.logger.Debug("git execution result", codeInt, codeCmd, err)
		} else {
			sh.logger.Debug("no repository info exists. seems to be a local shared.", pathTouse)
		}
	}
	config.Path = sh.getSourcePath(pathTouse)
	return config
}
