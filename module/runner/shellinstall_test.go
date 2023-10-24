package runner_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/runner"
)

func TestBashRcInstallFails(t *testing.T) {
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test")
	if err := installer.BashUserInstall(); err == nil {
		t.Error("should return an error, because the test folder do not contains a .bashrc file")
	}
}

func SetDefaultValues() {
	configure.SetShortcut("ctx")
	configure.SetCnShortcut("cn")
	configure.SetBinaryName("contxt")
}

func TestBashRcInstall(t *testing.T) {
	// defaults
	SetDefaultValues()
	defer os.RemoveAll("./test/fakehome/.bashrc")
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test/fakehome")
	os.WriteFile("./test/fakehome/.bashrc", []byte("# a fake bashrc"), 0644)
	if err := installer.BashUserInstall(); err != nil {
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
	assertFileContent(t, "./test/fakehome/.bashrc", expected, AcceptContainsNoSpecials)
}

func TestBashRcInstallRenamed(t *testing.T) {
	configure.SetShortcut("ctxV2")
	configure.SetCnShortcut("cnV2")
	configure.SetBinaryName("contxtV2")
	defer os.RemoveAll("./test/fakehome/.bashrc")
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test/fakehome")
	os.WriteFile("./test/fakehome/.bashrc", []byte("# a fake bashrc"), 0644)
	if err := installer.BashUserInstall(); err != nil {
		t.Error("should not return an error, bot got:", err)
	}
	expected := `# a fake bashrc
### begin contxtV2 bashrc
function cnV2() { cd $(contxtV2 dir find "$@"); }
function ctxV2() {
	contxtV2 "$@";
	[ $? -eq 0 ]  || return 1
        case $1 in
          switch)
          cd $(contxtV2 dir --last);
          contxtV2 dir paths --coloroff --nohints
          ;;
        esac
}
function ctxV2completion() {
        ORIG=$(contxtV2 completion bash)
        CM="contxtV2"
        CT="ctxV2"
        CTXC="${ORIG//$CM/$CT}"
        echo "$CTXC"
}
source <(contxtV2 completion bash)
source <(ctxV2completion)
### end of contxtV2 bashrc
`
	assertFileContent(t, "./test/fakehome/.bashrc", expected, AcceptContainsNoSpecials)
}

func TestZshRcInstallFails(t *testing.T) {
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test")
	if err := installer.ZshUserInstall(); err == nil {
		t.Error("should return an error, because the test folder do not contains a .zshrc file")
	}
}

func TestZshRcInstall(t *testing.T) {
	// defaults
	SetDefaultValues()
	defer os.RemoveAll("./test/fakehome/.zshrc")
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test/fakehome")
	os.WriteFile("./test/fakehome/.zshrc", []byte("# a fake zshrc"), 0644)
	if err := installer.ZshUserInstall(); err != nil {
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
	assertFileContent(t, "./test/fakehome/.zshrc", expected, AcceptContainsNoSpecials)
}

func TestZshRcInstallRenamed(t *testing.T) {
	configure.SetShortcut("CNTXT") // force the shortcut to ctx
	configure.SetCnShortcut("DD")  // force the shortcut to ctx
	configure.SetBinaryName("ZULU")
	defer os.RemoveAll("./test/fakehome/.zshrc")
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test/fakehome")
	os.WriteFile("./test/fakehome/.zshrc", []byte("# a fake zshrc"), 0644)
	if err := installer.ZshUserInstall(); err != nil {
		t.Error("should not return an error, bot got:", err)
	}
	expected := `# a fake zshrc
	### begin ZULU zshrc
	function DD() { cd $(ZULU dir find "$@"); }
	function CNTXT() {        
		ZULU "$@";
		[ $? -eq 0 ]  || return $?
			case $1 in
			  switch)          
			  cd $(ZULU dir --last);
			  ZULU dir paths --coloroff --nohints
			  ;;
			esac
	}
	### end of ZULU zshrc
`
	assertFileContent(t, "./test/fakehome/.zshrc", expected, AcceptContainsNoSpecials)
}

func TestFishRcInstall(t *testing.T) {
	// defaults
	SetDefaultValues()
	defer os.RemoveAll("./test/fakehome/.config")
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test/fakehome")
	//os.WriteFile("./test/fakehome/.config/fish/config.fish", []byte("# a fake fishrc"), 0644)
	if err := installer.FishFunctionUpdate(); err != nil {
		t.Error("should not return an error, bot got:", err)
	}
	functionFile := "test/fakehome/.config/fish/functions/ctx.fish"
	cnFunctionFile := "test/fakehome/.config/fish/functions/cn.fish"
	assertFileExists(t, functionFile)
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
	assertFileContent(t, functionFile, fishFunc, AcceptContainsNoSpecials)
	assertFileContent(t, cnFunctionFile, cnFunc, AcceptContainsNoSpecials)
}

func TestFishRcInstallRenamed(t *testing.T) {
	configure.SetShortcut("FctxF") // force the shortcut to ctx
	configure.SetCnShortcut("KKK") // force the shortcut to ctx
	configure.SetBinaryName("UINMKOI")
	defer os.RemoveAll("./test/fakehome/.config")
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test/fakehome")
	//os.WriteFile("./test/fakehome/.config/fish/config.fish", []byte("# a fake fishrc"), 0644)
	if err := installer.FishFunctionUpdate(); err != nil {
		t.Error("should not return an error, bot got:", err)
	}
	functionFile := "test/fakehome/.config/fish/functions/FctxF.fish"
	cnFunctionFile := "test/fakehome/.config/fish/functions/KKK.fish"
	assertFileExists(t, functionFile)
	fishFunc := `function FctxF
    UINMKOI $argv
    switch $argv[1]
       case switch
          cd (UINMKOI dir --last)
          UINMKOI dir paths --coloroff --nohints
    end
end`
	cnFunc := `function KKK
	cd (UINMKOI dir find $argv)
end`
	assertFileContent(t, functionFile, fishFunc, AcceptContainsNoSpecials)
	assertFileContent(t, cnFunctionFile, cnFunc, AcceptContainsNoSpecials)
}

func TestFishCompletionUpdate(t *testing.T) {
	// defaults
	SetDefaultValues()
	defer os.RemoveAll("./test/fakehome/.config")
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test/fakehome")

	cobra := runner.NewCobraCmds()
	if err := installer.FishCompletionUpdate(cobra.RootCmd); err != nil {
		t.Error("should not return an error, bot got:", err)
	}

	completionFile := "test/fakehome/.config/fish/completions/ctx.fish"
	assertFileExists(t, completionFile)

	completionFile = "test/fakehome/.config/fish/completions/contxt.fish"
	assertFileExists(t, completionFile)
}

func TestFishCompletionUpdateRenamed(t *testing.T) {
	configure.SetShortcut("UNU")
	configure.SetBinaryName("UNUBIN")

	defer os.RemoveAll("./test/fakehome/.config")
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test/fakehome")

	cobra := runner.NewCobraCmds()
	if err := installer.FishCompletionUpdate(cobra.RootCmd); err != nil {
		t.Error("should not return an error, bot got:", err)
	}

	completionFile := "test/fakehome/.config/fish/completions/UNU.fish"
	assertFileExists(t, completionFile)

	completionFile = "test/fakehome/.config/fish/completions/UNUBIN.fish"
	assertFileExists(t, completionFile)
}

func TestZshFuncDir(t *testing.T) {
	// defaults
	SetDefaultValues()
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
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

func TestZshUser(t *testing.T) {
	// defaults
	SetDefaultValues()
	os.WriteFile("./test/fakehome/.zshrc", []byte("# a fake zshrc"), 0644)
	defer os.Remove("./test/fakehome/.zshrc")
	defer os.Remove("./test/fakehome/zFuncExists/_ctx")
	defer os.Remove("./test/fakehome/zFuncExists/_contxt")
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test/fakehome")
	fpath := "[ABS]/test/fakehome/.zfunc:[ABS]/test/fakehome/zFuncExists:[ABS]/test/fakehome/zFuncNotExists"
	abs, _ := filepath.Abs(".")
	fpath = strings.ReplaceAll(fpath, "[ABS]", abs)
	os.Setenv("FPATH", fpath)

	cobra := runner.NewCobraCmds()

	if err := installer.ZshUpdate(cobra.RootCmd); err != nil {
		t.Error("should not return an error, bot got:", err)
	}
	assertFileExists(t, "test/fakehome/zFuncExists/_ctx")
	assertFileExists(t, "test/fakehome/zFuncExists/_contxt")
}

func TestZshUserRenamed(t *testing.T) {
	configure.SetShortcut("UGA")    // force the shortcut to UGA
	configure.SetBinaryName("NANA") // chage binaray name to NANA
	os.WriteFile("./test/fakehome/.zshrc", []byte("# a fake zshrc"), 0644)
	defer os.Remove("./test/fakehome/.zshrc")
	defer os.Remove("./test/fakehome/zFuncExists/_UGA")
	defer os.Remove("./test/fakehome/zFuncExists/_contxt")
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
	installer.SetUserHomePath("./test/fakehome")
	fpath := "[ABS]/test/fakehome/.zfunc:[ABS]/test/fakehome/zFuncExists:[ABS]/test/fakehome/zFuncNotExists"
	abs, _ := filepath.Abs(".")
	fpath = strings.ReplaceAll(fpath, "[ABS]", abs)
	os.Setenv("FPATH", fpath)

	cobra := runner.NewCobraCmds()

	if err := installer.ZshUpdate(cobra.RootCmd); err != nil {
		t.Error("should not return an error, bot got:", err)
	}
	assertFileExists(t, "test/fakehome/zFuncExists/_UGA")
	assertFileExists(t, "test/fakehome/zFuncExists/_NANA")
}

func TestZshDirFind(t *testing.T) {
	// testing depends if zsh is installed
	zsh := runner.NewZshHelper()
	zPath, err := zsh.GetBinPath()
	if err != nil || zPath == "" {
		t.Log("skipped zsh Testing, because it seems zsh not being installed.")
		t.SkipNow()
	}
	// defaults
	SetDefaultValues()
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
	instDir, err := installer.ZshFuncDir()
	if err != nil {
		t.Error("should not return an error, bot got:", err)
	}

	if instDir == "" {
		t.Error("should return a directory, but got:", instDir)
	}

}

func TestZshDirFindWithFpathSet(t *testing.T) {
	// testing depends if zsh is installed
	zsh := runner.NewZshHelper()
	zPath, err := zsh.GetBinPath()
	if err != nil || zPath == "" {
		t.Log("skipped zsh Testing, because it seems zsh not being installed.")
		t.SkipNow()
	}
	// defaults
	SetDefaultValues()
	os.Setenv("FPATH", "fpath1:fpath2:fpath3")
	installer := runner.NewShellInstall(mimiclog.NewNullLogger())
	_, err = installer.ZshFuncDir()
	// we expect an error, because the fpath is not a directory
	if err == nil {
		t.Error("should return an error, bot got:", err)
	}

}
