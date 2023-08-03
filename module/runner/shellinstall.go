// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// # Licensed under the MIT License
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
package runner

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"os"
	"os/user"
	"strings"

	"github.com/spf13/cobra"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
)

// here we have all the functions to install the shell completion and
// function files for the shell

type shellInstall struct {
	contxtHomePath    string
	pwrShellPathCache string
	userHomePath      string
	logger            mimiclog.Logger
}

func UserDirectory() (string, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir, err
}

// NewShellInstall returns a new shellInstall struct
func NewShellInstall(basePath string, logger mimiclog.Logger) *shellInstall {
	userPath, _ := UserDirectory()
	return &shellInstall{
		contxtHomePath: basePath,
		userHomePath:   userPath,
		logger:         logger,
	}
}

func (si *shellInstall) SetContxtBasePath(basePath string) {
	si.contxtHomePath = basePath
}

func (si *shellInstall) SetUserHomePath(userHomePath string) {
	si.userHomePath = userHomePath
}

func updateExistingFile(filename, content, doNotContain string) (bool, string) {
	ok, errDh := dirhandle.Exists(filename)
	errmsg := ""
	if errDh == nil && ok {
		byteCnt, err := os.ReadFile(filename)
		if err != nil {
			return false, "file not readable " + filename
		}
		strContent := string(byteCnt)
		if strings.Contains(strContent, doNotContain) {
			return false, "it seems file is already updated. it contains: " + doNotContain
		} else {
			file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				log.Println(err)
				return false, "error while opening file " + filename
			}
			defer file.Close()
			if _, err := file.WriteString(content); err != nil {
				log.Fatal(err)
				return false, "error adding content to file " + filename
			}
			return true, ""
		}

	} else {
		errmsg = "file update error: file not exists " + filename
	}
	return false, errmsg
}

// FishFunctionUpdate updates the fish function file
// and adds code completion for the fish shell
func (si *shellInstall) FishUpdate(cmd *cobra.Command) {
	si.FishFunctionUpdate()
	si.FishCompletionUpdate(cmd)
}

// FishCompletionUpdate updates the fish completion file
func (si *shellInstall) FishCompletionUpdate(cmd *cobra.Command) {

	if si.userHomePath != "" {
		// completion dir Exists ?
		exists, err := dirhandle.Exists(si.userHomePath + "/.config/fish/completions")
		if err == nil && !exists {
			mkErr := os.Mkdir(si.userHomePath+"/.config/fish/completions/", os.ModePerm)
			if mkErr != nil {
				si.logger.Critical(mkErr)
				systools.Exit(systools.ErrorBySystem)
			}
		}
	}
	cmpltn := new(bytes.Buffer)
	cmd.Root().GenFishCompletion(cmpltn, true)

	origin := cmpltn.String()
	ctxCmpltn := strings.ReplaceAll(origin, "contxt", "ctx")
	systools.WriteFileIfNotExists(si.userHomePath+"/.config/fish/completions/contxt.fish", origin)
	systools.WriteFileIfNotExists(si.userHomePath+"/.config/fish/completions/ctx.fish", ctxCmpltn)

}

func (si *shellInstall) FishFunctionUpdate() error {

	fishFunc := `function ctx
    contxt $argv
    switch $argv[1]
       case switch
          cd (contxt dir --last)
          contxt dir paths --coloroff --nohints
    end
end`
	cnFunc := `function cn
	cd (contxt dir find $argv)
end`

	usrDir := si.userHomePath
	if usrDir != "" {
		// functions dir Exists ?
		exists, err := dirhandle.Exists(usrDir + "/.config/fish/functions")
		if err == nil && !exists {
			usrDir, err := filepath.Abs(usrDir)
			if err != nil {
				return err
			}

			mkErr := os.MkdirAll(usrDir+"/.config/fish/functions", os.ModePerm)
			if mkErr != nil {
				return mkErr
			}
		}

		funcExists, funcErr := dirhandle.Exists(usrDir + "/.config/fish/functions/ctx.fish")
		if funcErr == nil && !funcExists {
			os.WriteFile(usrDir+"/.config/fish/functions/ctx.fish", []byte(fishFunc), 0644)
		} else if funcExists {
			return errors.New("ctx function already exists. did not change that")

		}

		funcExists, funcErr = dirhandle.Exists(usrDir + "/.config/fish/functions/cn.fish")
		if funcErr == nil && !funcExists {
			os.WriteFile(usrDir+"/.config/fish/functions/cn.fish", []byte(cnFunc), 0644)
		} else if funcExists {
			return errors.New("cn function already exists. did not change that")
		}
	}
	return nil
}

func (si *shellInstall) BashUser() error {
	bashrcAdd := `
### begin contxt bashrc
function cn() { cd $(contxt dir find "$@"); }
function ctx() {        
	contxt "$@";
	[ $? -eq 0 ]  || return 1
        case $1 in
          switch)          
          cd $(contxt dir --last);		  
          contxt dir paths --coloroff --nohints
          ;;
        esac
}
function ctxcompletion() {        
        ORIG=$(contxt completion bash)
        CM="contxt"
        CT="ctx"
        CTXC="${ORIG//$CM/$CT}"
        echo "$CTXC"
}
source <(contxt completion bash)
source <(ctxcompletion)
### end of contxt bashrc
	`
	usrDir := si.userHomePath
	if usrDir != "" {
		ok, errDh := dirhandle.Exists(usrDir + "/.bashrc")
		if errDh == nil && ok {
			fine, errmsg := updateExistingFile(usrDir+"/.bashrc", bashrcAdd, "### begin contxt bashrc")
			if !fine {
				return errors.New("bashrc update failed: " + errmsg)
			}
		} else {
			ctxout.PrintLn(ctxout.ForeRed, "missing .bashrc", ctxout.ForeWhite, "could not find expected "+usrDir+"/.bashrc")
			return errors.New("missing .bashrc")
		}
	}
	return nil
}

func (si *shellInstall) ZshUpdate(cmd *cobra.Command) error {
	if err := si.ZshUser(); err != nil {
		return err
	}
	return si.updateZshFunctions(cmd)
}

// try to get the best path by reading the permission
// because zsh seems not be used in windows, we stick to linux related
// permission check
func (si *shellInstall) ZshFuncDir() (string, error) {
	fpath := os.Getenv("FPATH")
	if fpath != "" {
		paths := strings.Split(fpath, ":")
		for _, path := range paths {
			fileStats, err := os.Stat(path)
			if err != nil {
				fmt.Println(err)
				continue
			}
			permissions := fileStats.Mode().Perm()
			if permissions&0b110000000 == 0b110000000 {
				return path, nil
			}
		}
		return "", errors.New("could not find zsh function dir that is accessible for writing. [" + fpath + "]")
	}
	return "", errors.New("could not find zsh function path. please set FPATH")
}

func (si *shellInstall) updateZshFunctions(cmd *cobra.Command) error {
	funcDir, err := si.ZshFuncDir()
	if err != nil {
		return err
	}
	if funcDir != "" {
		contxtPath := funcDir + "/_contxt"
		ctxPath := funcDir + "/_ctx"

		cmpltn := new(bytes.Buffer)
		cmd.Root().GenZshCompletion(cmpltn)

		origin := cmpltn.String()
		ctxCmpltn := strings.ReplaceAll(origin, "contxt", "ctx")

		systools.WriteFileIfNotExists(contxtPath, origin)
		systools.WriteFileIfNotExists(ctxPath, ctxCmpltn)
	} else {
		return errors.New("could not find zsh function dir")
	}
	return nil
}

func (si *shellInstall) ZshUser() error {
	zshrcAdd := `
### begin contxt zshrc
function cn() { cd $(contxt dir find "$@"); }
function ctx() {        
	contxt "$@";
	[ $? -eq 0 ]  || return $?
        case $1 in
          switch)          
          cd $(contxt dir --last);
          contxt dir paths --coloroff --nohints
          ;;
        esac
}
### end of contxt zshrc
	`
	usrDir := si.userHomePath
	if usrDir != "" {
		ok, errDh := dirhandle.Exists(usrDir + "/.zshrc")
		if errDh == nil && ok {
			fine, errmsg := updateExistingFile(usrDir+"/.zshrc", zshrcAdd, "### begin contxt zshrc")
			if !fine {
				return errors.New("zshrc update failed: " + errmsg)
			}
		} else {
			return errors.New("missing .zshrc. could not find expected " + usrDir + "/.zshrc")
		}
	}
	return nil
}

func (si *shellInstall) PwrShellUpdate(cmd *cobra.Command) error {
	forceProfile, _ := cmd.Flags().GetBool("create-profile")
	if forceProfile {
		si.PwrShellForceCreateProfile()
	}
	if err := si.PwrShellUser(); err != nil {
		return err
	}
	return si.PwrShellCompletionUpdate(cmd)
}

func (si *shellInstall) PwrShellUser() error {
	pwrshrcAdd := `
### begin contxt pwrshrc
function cn($path) {
	Set-Location $(contxt dir find $path)
}
function ctx {
	& contxt $args
}
### end of contxt pwrshrc
`
	if found, pwrshProfile := si.FindPwrShellProfile(); found {
		fine, errmsg := updateExistingFile(pwrshProfile, pwrshrcAdd, "### begin contxt pwrshrc")
		if !fine {
			return errors.New("pwrshrc update failed: " + errmsg)
		}
	} else {
		return errors.New("missing pwrshrc")
	}
	return nil
}

func (si *shellInstall) FindPwrShellProfile() (bool, string) {
	if si.pwrShellPathCache != "" {
		return true, si.pwrShellPathCache
	}
	pwrshProfile := os.Getenv("PROFILE")
	// retry by using powershell as host
	if pwrshProfile == "" {
		pwrShellRunner := tasks.GetShellRunnerForOs("windows")
		pwrshProfile, _ = pwrShellRunner.ExecSilentAndReturnLast(PWRSHELL_CMD_VERSION)
		//pwrshProfile = PwrShellExec(PWRSHELL_CMD_PROFILE)
	}
	if pwrshProfile != "" {
		fileStats, err := os.Stat(pwrshProfile)
		if err == nil {

			permissions := fileStats.Mode().Perm()
			if permissions&0b110000000 == 0b110000000 {
				si.pwrShellPathCache = pwrshProfile
				return true, pwrshProfile
			}
		}
	}
	return false, pwrshProfile
}

func (si *shellInstall) PwrShellForceCreateProfile() {
	if !si.PwrShellTestProfile() {
		pwrShellRunner := tasks.GetShellRunnerForOs("windows")
		pwrShellRunner.ExecSilentAndReturnLast(PWRSHELL_CMD_PROFILE_CREATE)
	}
}

func (si *shellInstall) PwrShellCompletionUpdate(cmd *cobra.Command) error {
	if !si.PwrShellTestProfile() {
		errormsg := `missing powershell profile. you can create a profile by running 'New-Item -Type File -Path $PROFILE -Force'`
		return errors.New(errormsg)
	}
	ok, profile := si.FindPwrShellProfile()
	if ok {
		cmpltn := new(bytes.Buffer)
		cmd.Root().GenPowerShellCompletion(cmpltn)
		origin := cmpltn.String()

		ctxCmpltn := strings.ReplaceAll(origin, "contxt", "ctx")

		ctxPowerShellPath := si.contxtHomePath + "/powershell"
		if exists, err := systools.Exists(ctxPowerShellPath); err != nil || !exists {
			if err := os.MkdirAll(ctxPowerShellPath, 0755); err != nil {
				return err
			}
		}
		systools.WriteFileIfNotExists(ctxPowerShellPath+"/contxt.ps1", origin)
		systools.WriteFileIfNotExists(ctxPowerShellPath+"/ctx.ps1", ctxCmpltn)

		profileAdd := `
### begin contxt powershell profile
. "` + ctxPowerShellPath + `/contxt.ps1"
. "` + ctxPowerShellPath + `/ctx.ps1"
### end of contxt powershell profile
`

		fine, errmsg := updateExistingFile(profile, profileAdd, "### begin contxt powershell profile")
		if !fine {
			return errors.New("powershell profile update failed: " + errmsg)
		}

	} else {
		return errors.New("could not find a writable path for powershell completion")
	}
	return nil
}

func (si *shellInstall) PwrShellTestProfile() bool {
	pwrShellRunner := tasks.GetShellRunnerForOs("windows")
	foundBool, _ := pwrShellRunner.ExecSilentAndReturnLast(PWRSHELL_CMD_TEST_PROFILE)
	return strings.ToLower(foundBool) == "true"
}
