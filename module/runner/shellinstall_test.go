package runner_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/runner"
)

func TestBashRcInstallFails(t *testing.T) {
	installer := runner.NewShellInstall("./test", mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test")
	if err := installer.BashUser(); err == nil {
		t.Error("should return an error, because the test folder do not contains a .bashrc file")
	}
}

func TestBashRcInstall(t *testing.T) {
	defer os.RemoveAll("./test/fakehome/.bashrc")
	installer := runner.NewShellInstall("./test", mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test/fakehome")
	os.WriteFile("./test/fakehome/.bashrc", []byte("# a fake bashrc"), 0644)
	if err := installer.BashUser(); err != nil {
		t.Error("should not return an error, bot got:", err)
	}
	expected := `# a fake bashrc
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
	AssertFileContent(t, "./test/fakehome/.bashrc", expected, AcceptContainsNoSpecials)
}

func TestZshRcInstallFails(t *testing.T) {
	installer := runner.NewShellInstall("./test", mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test")
	if err := installer.ZshUser(); err == nil {
		t.Error("should return an error, because the test folder do not contains a .zshrc file")
	}
}

func TestZshRcInstall(t *testing.T) {
	defer os.RemoveAll("./test/fakehome/.zshrc")
	installer := runner.NewShellInstall("./test", mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test/fakehome")
	os.WriteFile("./test/fakehome/.zshrc", []byte("# a fake zshrc"), 0644)
	if err := installer.ZshUser(); err != nil {
		t.Error("should not return an error, bot got:", err)
	}
	expected := `# a fake zshrc
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
	AssertFileContent(t, "./test/fakehome/.zshrc", expected, AcceptContainsNoSpecials)
}

func TestFishRcInstall(t *testing.T) {
	defer os.RemoveAll("./test/fakehome/.config")
	installer := runner.NewShellInstall("./test", mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test/fakehome")
	//os.WriteFile("./test/fakehome/.config/fish/config.fish", []byte("# a fake fishrc"), 0644)
	if err := installer.FishFunctionUpdate(); err != nil {
		t.Error("should not return an error, bot got:", err)
	}
	functionFile := "test/fakehome/.config/fish/functions/ctx.fish"
	cnFunctionFile := "test/fakehome/.config/fish/functions/cn.fish"
	AssertFileExists(t, functionFile)
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
	AssertFileContent(t, functionFile, fishFunc, AcceptContainsNoSpecials)
	AssertFileContent(t, cnFunctionFile, cnFunc, AcceptContainsNoSpecials)
}

func TestZshFuncDir(t *testing.T) {
	ChangeToRuntimeDir(t)
	installer := runner.NewShellInstall("./test", mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test/fakehome")
	fpath := "[ABS]/test/fakehome/.zfunc:[ABS]/test/fakehome/zFuncExists:[ABS]/test/fakehome/zFuncNotExists"
	abs, _ := filepath.Abs(".")
	fpath = strings.ReplaceAll(fpath, "[ABS]", abs)
	os.Setenv("FPATH", fpath)

	path, err := installer.ZshFuncDir()
	if err != nil {
		t.Error("should not return an error, bot got:", err)
	}
	expectedpath, err := filepath.Abs("./test/fakehome/zFuncExists")
	if err != nil {
		t.Error("should not return an error, bot got:", err)
	} else if path != expectedpath {
		t.Error("should return \n", expectedpath, ", but got:\n", path, "\n")
	}
}
