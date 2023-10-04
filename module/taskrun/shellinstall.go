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
package taskrun

import (
	"bytes"
	"fmt"

	"os"
	"os/user"
	"strings"

	"github.com/spf13/cobra"

	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/manout"
)

// here we have all the functions to install the shell completion and
// function files for the shell

var pwrShellPathCache string = "" // cache the path to the powershell profile

func UserDirectory() (string, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir, err
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
func FishUpdate(cmd *cobra.Command) {
	FishFunctionUpdate()
	FishCompletionUpdate(cmd)
}

// FishCompletionUpdate updates the fish completion file
func FishCompletionUpdate(cmd *cobra.Command) {
	usrDir, err := UserDirectory()
	if err == nil && usrDir != "" {
		// completion dir Exists ?
		exists, err := dirhandle.Exists(usrDir + "/.config/fish/completions")
		if err == nil && !exists {
			mkErr := os.Mkdir(usrDir+"/.config/fish/completions/", os.ModePerm)
			if mkErr != nil {
				log.Fatal(mkErr)
			}
		}
	}
	cmpltn := new(bytes.Buffer)
	cmd.Root().GenFishCompletion(cmpltn, true)

	origin := cmpltn.String()
	ctxCmpltn := strings.ReplaceAll(origin, "contxt", "ctx")
	systools.WriteFileIfNotExists(usrDir+"/.config/fish/completions/contxt.fish", origin)
	systools.WriteFileIfNotExists(usrDir+"/.config/fish/completions/ctx.fish", ctxCmpltn)

}

func FishFunctionUpdate() {

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

	usrDir, err := UserDirectory()
	if err == nil && usrDir != "" {
		// functions dir Exists ?
		exists, err := dirhandle.Exists(usrDir + "/.config/fish/functions")
		if err == nil && !exists {
			mkErr := os.Mkdir(usrDir+"/.config/fish/functions", os.ModePerm)
			if mkErr != nil {
				log.Fatal(mkErr)
			}
		}

		funcExists, funcErr := dirhandle.Exists(usrDir + "/.config/fish/functions/ctx.fish")
		if funcErr == nil && !funcExists {
			os.WriteFile(usrDir+"/.config/fish/functions/ctx.fish", []byte(fishFunc), 0644)
		} else if funcExists {
			fmt.Println("ctx function already exists. did not change that")
		}

		funcExists, funcErr = dirhandle.Exists(usrDir + "/.config/fish/functions/cn.fish")
		if funcErr == nil && !funcExists {
			os.WriteFile(usrDir+"/.config/fish/functions/cn.fish", []byte(cnFunc), 0644)
		} else if funcExists {
			fmt.Println("cn function already exists. did not change that")
		}
	}
}

func BashUser() {
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
	usrDir, err := UserDirectory()
	if err == nil && usrDir != "" {
		ok, errDh := dirhandle.Exists(usrDir + "/.bashrc")
		if errDh == nil && ok {
			fmt.Println(usrDir + "/.bashrc")
			fine, errmsg := updateExistingFile(usrDir+"/.bashrc", bashrcAdd, "### begin contxt bashrc")
			if !fine {
				manout.Error("bashrc update failed", errmsg)
			} else {
				fmt.Println(manout.MessageCln(manout.ForeGreen, "success", manout.CleanTag, " to update bash run ", manout.ForeCyan, " source ~/.bashrc"))
			}
		} else {
			manout.Error("missing .bashrc", "could not find expected "+usrDir+"/.bashrc")
		}
	}

}

func ZshUpdate(cmd *cobra.Command) {
	ZshUser()
	updateZshFunctions(cmd)
}

// try to get the best path by reading the permission
// because zsh seems not be used in windows, we stick to linux related
// permission check
func ZshFuncDir() string {
	fpath := os.Getenv("FPATH")
	if fpath != "" {
		paths := strings.Split(fpath, ":")
		for _, path := range paths {
			fileStats, err := os.Stat(path)
			if err != nil {
				continue
			}
			permissions := fileStats.Mode().Perm()
			if permissions&0b110000000 == 0b110000000 {
				return path
			}
		}
		return ""
	}
	return fpath
}

func updateZshFunctions(cmd *cobra.Command) {
	funcDir := ZshFuncDir()
	if funcDir != "" {
		contxtPath := funcDir + "/_contxt"
		ctxPath := funcDir + "/_ctx"
		fmt.Println(funcDir)

		cmpltn := new(bytes.Buffer)
		cmd.Root().GenZshCompletion(cmpltn)

		origin := cmpltn.String()
		ctxCmpltn := strings.ReplaceAll(origin, "contxt", "ctx")

		systools.WriteFileIfNotExists(contxtPath, origin)
		systools.WriteFileIfNotExists(ctxPath, ctxCmpltn)
	} else {
		manout.Error("could not find a writable path for zsh functions in fpath")
	}
}

func ZshUser() {
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
	usrDir, err := UserDirectory()
	if err == nil && usrDir != "" {
		ok, errDh := dirhandle.Exists(usrDir + "/.zshrc")
		if errDh == nil && ok {
			fmt.Println(usrDir + "/.zshrc")
			fine, errmsg := updateExistingFile(usrDir+"/.zshrc", zshrcAdd, "### begin contxt zshrc")
			if !fine {
				manout.Error("zshrc update failed", errmsg)
			} else {
				fmt.Println(manout.MessageCln(manout.ForeGreen, "success", manout.CleanTag, "  ", manout.ForeCyan, " "))
			}
		} else {
			manout.Error("missing .zshrc", "could not find expected "+usrDir+"/.zshrc")
		}
	}

}

func PwrShellUpdate(cmd *cobra.Command) {
	forceProfile, _ := cmd.Flags().GetBool("create-profile")
	if forceProfile {
		PwrShellForceCreateProfile()
	}
	PwrShellUser()
	PwrShellCompletionUpdate(cmd)
}

func PwrShellUser() {
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
	if found, pwrshProfile := FindPwrShellProfile(); found {
		fine, errmsg := updateExistingFile(pwrshProfile, pwrshrcAdd, "### begin contxt pwrshrc")
		if !fine {
			manout.Error("pwrshrc update failed", errmsg)
		} else {
			fmt.Println(manout.MessageCln(manout.ForeGreen, "success", manout.CleanTag, "  ", manout.ForeCyan, " "))
		}
	} else {
		manout.Error("missing pwrshrc", "could not find expected powershell profile")
	}
}

func FindPwrShellProfile() (bool, string) {
	if pwrShellPathCache != "" {
		return true, pwrShellPathCache
	}
	pwrshProfile := os.Getenv("PROFILE")
	// retry by using powershell as host
	if pwrshProfile == "" {
		pwrshProfile = PwrShellExec(PWRSHELL_CMD_PROFILE)
	}
	if pwrshProfile != "" {
		fileStats, err := os.Stat(pwrshProfile)
		if err == nil {

			permissions := fileStats.Mode().Perm()
			if permissions&0b110000000 == 0b110000000 {
				pwrShellPathCache = pwrshProfile
				return true, pwrshProfile
			}
		}
	}
	return false, pwrshProfile
}

func PwrShellForceCreateProfile() {
	if !PwrShellTestProfile() {
		PwrShellExec(PWRSHELL_CMD_PROFILE_CREATE)
	}
}

func PwrShellCompletionUpdate(cmd *cobra.Command) {
	if !PwrShellTestProfile() {
		manout.Error("missing powershell profile", "could not find expected powershell profile")
		manout.Om.Println("you can create a profile by running 'New-Item -Type File -Path $PROFILE -Force'")
		return
	}
	ok, profile := FindPwrShellProfile()
	if ok {
		cmpltn := new(bytes.Buffer)
		cmd.Root().GenPowerShellCompletion(cmpltn)
		origin := cmpltn.String()
		if ctxBasePath, err := GetContxtBasePath(); err == nil {
			ctxCmpltn := strings.ReplaceAll(origin, "contxt", "ctx")

			ctxPowerShellPath := ctxBasePath + "/powershell"
			if exists, err := systools.Exists(ctxPowerShellPath); err != nil || !exists {
				if err := os.MkdirAll(ctxPowerShellPath, 0755); err != nil {
					manout.Error("error", "could not create the contxt base path")
					return
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
				manout.Error("powershell profile update failed", errmsg)
			} else {
				manout.MessageCln(manout.ForeGreen, "success", manout.CleanTag, "  ", manout.ForeCyan, " ")
			}

		} else {
			manout.Error("error", "could not find the contxt base path")
		}
	} else {
		manout.Error("could not find a writable path for powershell completion")
	}
}

func PwrShellTestProfile() bool {
	foundBool := PwrShellExec(PWRSHELL_CMD_TEST_PROFILE)
	return strings.ToLower(foundBool) == "true"
}
