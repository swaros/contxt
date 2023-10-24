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
	"log"
	"path/filepath"

	"os"
	"os/user"
	"strings"

	"github.com/spf13/cobra"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
)

// here we have all the functions to install the shell completion and
// function files for the shell

type shellInstall struct {
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
func NewShellInstall(logger mimiclog.Logger) *shellInstall {
	userPath, _ := UserDirectory()
	return &shellInstall{
		userHomePath: userPath,
		logger:       logger,
	}
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
func (si *shellInstall) FishUpdate(cmd *cobra.Command) error {
	if err := si.FishFunctionUpdate(); err != nil {
		return err
	}
	return si.FishCompletionUpdate(cmd)
}

// FishCompletionUpdate updates the fish completion file
func (si *shellInstall) FishCompletionUpdate(cmd *cobra.Command) error {
	shortCut, _, binName := configure.GetShortcutsAndBinaryName()
	if si.userHomePath != "" {
		// completion dir Exists ?
		exists, err := dirhandle.Exists(si.userHomePath + "/.config/fish/completions")
		if err == nil && !exists {
			mkErr := os.MkdirAll(si.userHomePath+"/.config/fish/completions/", os.ModePerm)
			if mkErr != nil {
				return mkErr
			}
		}
	}
	cmpltn := new(bytes.Buffer)
	cmd.Root().GenFishCompletion(cmpltn, true)

	origin := cmpltn.String()
	ctxCmpltn := strings.ReplaceAll(origin, binName, shortCut)
	if _, err := systools.WriteFileIfNotExists(si.userHomePath+"/.config/fish/completions/"+binName+".fish", origin); err != nil {
		return err
	}
	if _, err := systools.WriteFileIfNotExists(si.userHomePath+"/.config/fish/completions/"+shortCut+".fish", ctxCmpltn); err != nil {
		return err
	}
	return nil
}

func (si *shellInstall) FishFunctionUpdate() error {
	shortCut, cnShort, binName := configure.GetShortcutsAndBinaryName()
	fishFunc := `function ` + shortCut + `
    ` + binName + ` $argv
    switch $argv[1]
       case switch
          cd (` + binName + ` dir --last)
          ` + binName + ` dir paths --coloroff --nohints
    end
end`
	cnFunc := `function ` + cnShort + `
	cd (` + binName + ` dir find $argv)
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

		funcExists, funcErr := dirhandle.Exists(usrDir + "/.config/fish/functions/" + shortCut + ".fish")
		if funcErr == nil && !funcExists {
			os.WriteFile(usrDir+"/.config/fish/functions/"+shortCut+".fish", []byte(fishFunc), 0644)
		} else if funcExists {
			return errors.New(shortCut + " function already exists. did not change that")

		}

		funcExists, funcErr = dirhandle.Exists(usrDir + "/.config/fish/functions/" + cnShort + ".fish")
		if funcErr == nil && !funcExists {
			os.WriteFile(usrDir+"/.config/fish/functions/"+cnShort+".fish", []byte(cnFunc), 0644)
		} else if funcExists {
			return errors.New(cnShort + " function already exists. did not change that")
		}
	}
	return nil
}

func (si *shellInstall) BashUserInstall() error {
	shortCut, cnShort, binName := configure.GetShortcutsAndBinaryName()
	bashrcAdd := `
### begin ` + binName + ` bashrc
function ` + cnShort + `() { cd $(` + binName + ` dir find "$@"); }
function ` + shortCut + `() {        
	` + binName + ` "$@";
	[ $? -eq 0 ]  || return 1
        case $1 in
          switch)          
          cd $(` + binName + ` dir --last);		  
          ` + binName + ` dir paths --coloroff --nohints
          ;;
        esac
}
function ` + shortCut + `completion() {        
        ORIG=$(` + binName + ` completion bash)
        CM="` + binName + `"
        CT="` + shortCut + `"
        CTXC="${ORIG//$CM/$CT}"
        echo "$CTXC"
}
source <(` + binName + ` completion bash)
source <(` + shortCut + `completion)
### end of ` + binName + ` bashrc
	`
	usrDir := si.userHomePath
	if usrDir != "" {
		ok, errDh := dirhandle.Exists(usrDir + "/.bashrc")
		if errDh == nil && ok {
			fine, errmsg := updateExistingFile(usrDir+"/.bashrc", bashrcAdd, "### begin "+binName+" bashrc")
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
	if err := si.ZshUserInstall(); err != nil {
		return err
	}
	return si.updateZshFunctions(cmd)
}

// try to get the best path by reading the permission
// because zsh seems not be used in windows, we stick to linux related
// permission check
func (si *shellInstall) ZshFuncDir() (string, error) {
	zsh := NewZshHelper()
	return zsh.GetFirstFPath()
}

func (si *shellInstall) updateZshFunctions(cmd *cobra.Command) error {
	shortCut, _, binName := configure.GetShortcutsAndBinaryName()
	funcDir, err := si.ZshFuncDir()
	if err != nil {
		return err
	}
	if funcDir != "" {
		contxtPath := funcDir + "/_" + binName
		ctxPath := funcDir + "/_" + shortCut

		cmpltn := new(bytes.Buffer)
		cmd.Root().GenZshCompletion(cmpltn)

		origin := cmpltn.String()
		ctxCmpltn := strings.ReplaceAll(origin, binName, shortCut)

		systools.WriteFileIfNotExists(contxtPath, origin)
		systools.WriteFileIfNotExists(ctxPath, ctxCmpltn)
	} else {
		return errors.New("could not find zsh function dir")
	}
	return nil
}

func (si *shellInstall) ZshUserInstall() error {
	shortCut, cnShort, binName := configure.GetShortcutsAndBinaryName()
	zshrcAdd := `
### begin ` + binName + ` zshrc
function ` + cnShort + `() { cd $(` + binName + ` dir find "$@"); }
function ` + shortCut + `() {        
	` + binName + ` "$@";
	[ $? -eq 0 ]  || return $?
        case $1 in
          switch)          
          cd $(` + binName + ` dir --last);
          ` + binName + ` dir paths --coloroff --nohints
          ;;
        esac
}
### end of ` + binName + ` zshrc
	`
	usrDir := si.userHomePath
	if usrDir != "" {
		ok, errDh := dirhandle.Exists(usrDir + "/.zshrc")
		if errDh == nil && ok {
			fine, errmsg := updateExistingFile(usrDir+"/.zshrc", zshrcAdd, "### begin "+binName+" zshrc")
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
	shortCut, cnShort, binName := configure.GetShortcutsAndBinaryName()
	pwrshrcAdd := `
### begin ` + binName + ` pwrshrc
function ` + cnShort + `($path) {
	Set-Location $(` + binName + ` dir find $path)
}
function ` + shortCut + ` {
	& ` + binName + ` $args
}
### end of ` + binName + ` pwrshrc
`
	if found, pwrshProfile := si.FindPwrShellProfile(); found {
		fine, errmsg := updateExistingFile(pwrshProfile, pwrshrcAdd, "### begin "+binName+" pwrshrc")
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
	shortCut, _, binName := configure.GetShortcutsAndBinaryName()
	if !si.PwrShellTestProfile() {
		errormsg := `missing powershell profile. you can create a profile by running 'New-Item -Type File -Path $PROFILE -Force'`
		return errors.New(errormsg)
	}
	ok, profile := si.FindPwrShellProfile()
	if ok {
		cmpltn := new(bytes.Buffer)
		cmd.Root().GenPowerShellCompletion(cmpltn)
		origin := cmpltn.String()

		ctxCmpltn := strings.ReplaceAll(origin, binName, shortCut)

		ctxPowerShellPath := si.userHomePath + "/." + binName + "/powershell"
		if exists, err := systools.Exists(ctxPowerShellPath); err != nil || !exists {
			if err := os.MkdirAll(ctxPowerShellPath, 0755); err != nil {
				return err
			}
		}
		systools.WriteFileIfNotExists(ctxPowerShellPath+"/"+binName+".ps1", origin)
		systools.WriteFileIfNotExists(ctxPowerShellPath+"/"+shortCut+".ps1", ctxCmpltn)

		profileAdd := `
### begin ` + binName + ` powershell profile
. "` + ctxPowerShellPath + `/` + binName + `.ps1"
. "` + ctxPowerShellPath + `/` + shortCut + `.ps1"
### end of ` + binName + ` powershell profile
`

		fine, errmsg := updateExistingFile(profile, profileAdd, "### begin "+binName+" powershell profile")
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
