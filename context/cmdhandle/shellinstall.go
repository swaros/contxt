package cmdhandle

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"github.com/spf13/cobra"
	"github.com/swaros/contxt/context/output"

	"github.com/swaros/contxt/context/dirhandle"
)

func UserDirectory() (string, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir, err
}

func FishUpdate(cmd *cobra.Command) {
	FishFunctionUpdate()
	FishCompletionUpdate(cmd)
}

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
	WriteFileIfNotExists(usrDir+"/.config/fish/completions/contxt.fish", origin)
	WriteFileIfNotExists(usrDir+"/.config/fish/completions/ctx.fish", ctxCmpltn)

}

func WriteFileIfNotExists(filename, content string) (int, error) {
	funcExists, funcErr := dirhandle.Exists(filename)
	if funcErr == nil && !funcExists {
		ioutil.WriteFile(filename, []byte(content), 0644)
		return 0, nil
	} else if funcExists {
		return 1, nil
	}
	return 2, funcErr

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
	cd (contxt dir -i $argv)
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
			ioutil.WriteFile(usrDir+"/.config/fish/functions/ctx.fish", []byte(fishFunc), 0644)
		} else if funcExists {
			fmt.Println("ctx function already exists. did not change that")
		}

		funcExists, funcErr = dirhandle.Exists(usrDir + "/.config/fish/functions/cn.fish")
		if funcErr == nil && !funcExists {
			ioutil.WriteFile(usrDir+"/.config/fish/functions/cn.fish", []byte(cnFunc), 0644)
		} else if funcExists {
			fmt.Println("cn function already exists. did not change that")
		}
	}
}

func BashUser() {
	bashrcAdd := `
### begin contxt bashrc
function cn() { cd $(contxt dir -i "$@"); }
function ctx() {        
	contxt "$@";
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
				output.Error("bashrc update failed", errmsg)
			} else {
				fmt.Println(output.MessageCln(output.ForeGreen, "success", output.CleanTag, " to update bash run ", output.ForeCyan, " source ~/.bashrc"))
			}
		} else {
			output.Error("missing .bashrc", "could not find expected "+usrDir+"/.bashrc")
		}
	}

}

func updateExistingFile(filename, content, doNotContain string) (bool, string) {
	ok, errDh := dirhandle.Exists(filename)
	errmsg := ""
	if errDh == nil && ok {
		byteCnt, err := ioutil.ReadFile(filename)
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
