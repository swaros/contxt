package runner_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/runner"
)

func TestBashCompletionOrigin(t *testing.T) {
	SetDefaultValues()
	cobra := runner.NewCobraCmds()
	cmpltn := new(bytes.Buffer)
	cobra.RootCmd.GenBashCompletion(cmpltn)
	cobraResult := cmpltn.String()
	err := cobra.RootCmd.Execute()

	if err != nil {
		t.Errorf("Error executing completion command: %s", err)
	}

	if cobraResult == "" {
		t.Errorf("Completion result is empty")
	}

	// testing some of the expected methods exists
	// __contxt_handle_reply
	// __contxt_handle_go_custom_completion
	// __contxt_handle_subdirs_in_dir_flag
	expectedSlice := []string{
		`__contxt_index_of_word()`,
		`__contxt_init_completion()`,
		`__contxt_contains_word()`,
		`__contxt_handle_go_custom_completion()`,
		`__contxt_handle_reply()`,
	}

	assertStringSliceInContent(t, cobraResult, expectedSlice, IgnoreMultiSpaces, IgnoreTabs, IgnoreNewLines, IgnoreTabs, IgnoreSpaces)

}

// testing renamed binary name for bash completion
func TestBashCompletionRenamed(t *testing.T) {
	configure.SetBinaryName("RenamedBin")
	cobra := runner.NewCobraCmds()
	cmpltn := new(bytes.Buffer)
	cobra.RootCmd.GenBashCompletion(cmpltn)
	cobraResult := cmpltn.String()
	err := cobra.RootCmd.Execute()

	if err != nil {
		t.Errorf("Error executing completion command: %s", err)
	}

	if cobraResult == "" {
		t.Errorf("Completion result is empty")
	}

	// testing some of the expected methods exists
	// __contxt_handle_reply
	// __contxt_handle_go_custom_completion
	// __contxt_handle_subdirs_in_dir_flag
	expectedSlice := []string{
		`__RenamedBin_index_of_word()`,
		`__RenamedBin_init_completion()`,
		`__RenamedBin_contains_word()`,
		`__RenamedBin_handle_go_custom_completion()`,
		`__RenamedBin_debug "${FUNCNAME[0]}`,
		`__RenamedBin_handle_reply()`,
	}

	assertStringSliceInContent(t, cobraResult, expectedSlice, IgnoreMultiSpaces, IgnoreTabs, IgnoreNewLines, IgnoreTabs, IgnoreSpaces)

}

// testing zsh completion with default values
func TestCompletionZshOrgin(t *testing.T) {
	SetDefaultValues()
	result := runCobraCommand(func(cobra *runner.SessionCobra, w io.Writer) {
		cobra.RootCmd.GenZshCompletion(w)
	})

	if result == "" {
		t.Errorf("Completion result is empty")
	}

	expectedSlice := []string{
		`#compdef contxt`,
		`__contxt_debug`,
	}
	assertStringSliceInContent(t, result, expectedSlice)
}

// testing renamed binary name for zsh completion
func TestCompletionZshRenamedBin(t *testing.T) {
	configure.SetBinaryName("contxtV2")
	result := runCobraCommand(func(cobra *runner.SessionCobra, w io.Writer) {
		cobra.RootCmd.GenZshCompletion(w)
	})

	if result == "" {
		t.Errorf("Completion result is empty")
	}

	expectedSlice := []string{
		`#compdef contxtV2`,
		`__contxtV2_debug`,
	}
	assertStringSliceInContent(t, result, expectedSlice)
}

// testing fish completion with default values
func TestCompletionFishOrgin(t *testing.T) {
	SetDefaultValues()
	result := runCobraCommand(func(cobra *runner.SessionCobra, w io.Writer) {
		cobra.RootCmd.GenFishCompletion(w, true)
	})
	expectedSlice := []string{
		`fish completion for contxt`,
		`function __contxt_debug`,
		`__contxt_perform_completion`,
		`__contxt_perform_completion_once_result`,
		`function __contxt_prepare_completions`,
	}
	assertStringSliceInContent(t, result, expectedSlice)
}

// testing renamed binary name for fish completion
func TestCompletionFishRenamedBin(t *testing.T) {
	configure.SetBinaryName("contxtV2")
	result := runCobraCommand(func(cobra *runner.SessionCobra, w io.Writer) {
		cobra.RootCmd.GenFishCompletion(w, true)
	})
	expectedSlice := []string{
		`fish completion for contxtV2`,
		`function __contxtV2_debug`,
		`__contxtV2_perform_completion`,
		`__contxtV2_perform_completion_once_result`,
		`function __contxtV2_prepare_completions`,
	}
	assertStringSliceInContent(t, result, expectedSlice)
}

// testing powershell completion with default values
func TestCompletionPowershellOrgin(t *testing.T) {
	SetDefaultValues()
	result := runCobraCommand(func(cobra *runner.SessionCobra, w io.Writer) {
		cobra.RootCmd.GenPowerShellCompletion(w)
	})
	expectedSlice := []string{
		`powershell completion for contxt`,
		`function __contxt_debug`,
		`CONTXT_ACTIVE_HELP`,
	}
	assertStringSliceInContent(t, result, expectedSlice)
}

// testing renamed binary name for powershell completion
func TestCompletionPowershellRenamedBin(t *testing.T) {
	configure.SetBinaryName("contxtV2")
	result := runCobraCommand(func(cobra *runner.SessionCobra, w io.Writer) {
		cobra.RootCmd.GenPowerShellCompletion(w)
	})
	expectedSlice := []string{
		`powershell completion for contxtV2`,
		`function __contxtV2_debug`,
		`CONTXTV2_ACTIVE_HELP`,
	}
	assertStringSliceInContent(t, result, expectedSlice)
}
