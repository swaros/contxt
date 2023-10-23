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
		`__contxt_index_of_word()
	{
		local w word=$1
		shift
		index=0
		for w in "$@"; do
			[[ $w = "$word" ]] && return
			index=$((index+1))
		done
		index=-1
	}`,
		`__contxt_init_completion()
		{
			COMPREPLY=()
			_get_comp_words_by_ref "$@" cur prev words cword
		}`,
		`__contxt_contains_word()
		{
			local w word=$1; shift
			for w in "$@"; do
				[[ $w = "$word" ]] && return
			done
			return 1
		}`,
		`__contxt_handle_go_custom_completion()
		{
			__contxt_debug "${FUNCNAME[0]}: cur is ${cur}, words[*] is ${words[*]}, #words[@] is ${#words[@]}"
		
			local shellCompDirectiveError=1
			local shellCompDirectiveNoSpace=2
			local shellCompDirectiveNoFileComp=4
			local shellCompDirectiveFilterFileExt=8
			local shellCompDirectiveFilterDirs=16
		
			local out requestComp lastParam lastChar comp directive args
		
			# Prepare the command to request completions for the program.
			# Calling ${words[0]} instead of directly contxt allows to handle aliases
			args=("${words[@]:1}")
			# Disable ActiveHelp which is not supported for bash completion v1
			requestComp="CONTXT_ACTIVE_HELP=0 ${words[0]} __completeNoDesc ${args[*]}"
		
			lastParam=${words[$((${#words[@]}-1))]}
			lastChar=${lastParam:$((${#lastParam}-1)):1}
			__contxt_debug "${FUNCNAME[0]}: lastParam ${lastParam}, lastChar ${lastChar}"
		
			if [ -z "${cur}" ] && [ "${lastChar}" != "=" ]; then
				# If the last parameter is complete (there is a space following it)
				# We add an extra empty parameter so we can indicate this to the go method.
				__contxt_debug "${FUNCNAME[0]}: Adding extra empty parameter"
				requestComp="${requestComp} \"\""
			fi
		
			__contxt_debug "${FUNCNAME[0]}: calling ${requestComp}"
			# Use eval to handle any environment variables and such
			out=$(eval "${requestComp}" 2>/dev/null)
		
			# Extract the directive integer at the very end of the output following a colon (:)
			directive=${out##*:}
			# Remove the directive
			out=${out%:*}
			if [ "${directive}" = "${out}" ]; then
				# There is not directive specified
				directive=0
			fi
			__contxt_debug "${FUNCNAME[0]}: the completion directive is: ${directive}"
			__contxt_debug "${FUNCNAME[0]}: the completions are: ${out}"
		
			if [ $((directive & shellCompDirectiveError)) -ne 0 ]; then
				# Error code.  No completion.
				__contxt_debug "${FUNCNAME[0]}: received error from custom completion go code"
				return
			else
				if [ $((directive & shellCompDirectiveNoSpace)) -ne 0 ]; then
					if [[ $(type -t compopt) = "builtin" ]]; then
						__contxt_debug "${FUNCNAME[0]}: activating no space"
						compopt -o nospace
					fi
				fi
				if [ $((directive & shellCompDirectiveNoFileComp)) -ne 0 ]; then
					if [[ $(type -t compopt) = "builtin" ]]; then
						__contxt_debug "${FUNCNAME[0]}: activating no file completion"
						compopt +o default
					fi
				fi
			fi
		
			if [ $((directive & shellCompDirectiveFilterFileExt)) -ne 0 ]; then
				# File extension filtering
				local fullFilter filter filteringCmd
				# Do not use quotes around the $out variable or else newline
				# characters will be kept.
				for filter in ${out}; do
					fullFilter+="$filter|"
				done
		
				filteringCmd="_filedir $fullFilter"
				__contxt_debug "File filtering command: $filteringCmd"
				$filteringCmd
			elif [ $((directive & shellCompDirectiveFilterDirs)) -ne 0 ]; then
				# File completion for directories only
				local subdir
				# Use printf to strip any trailing newline
				subdir=$(printf "%s" "${out}")
				if [ -n "$subdir" ]; then
					__contxt_debug "Listing directories in $subdir"
					__contxt_handle_subdirs_in_dir_flag "$subdir"
				else
					__contxt_debug "Listing directories in ."
					_filedir -d
				fi
			else
				while IFS='' read -r comp; do
					COMPREPLY+=("$comp")
				done < <(compgen -W "${out}" -- "$cur")
			fi
		}`,
		`__contxt_handle_reply()
		{
			__contxt_debug "${FUNCNAME[0]}"
			local comp
			case $cur in
				-*)
					if [[ $(type -t compopt) = "builtin" ]]; then
						compopt -o nospace
					fi
					local allflags
					if [ ${#must_have_one_flag[@]} -ne 0 ]; then
						allflags=("${must_have_one_flag[@]}")
					else
						allflags=("${flags[*]} ${two_word_flags[*]}")
					fi
					while IFS='' read -r comp; do
						COMPREPLY+=("$comp")
					done < <(compgen -W "${allflags[*]}" -- "$cur")
					if [[ $(type -t compopt) = "builtin" ]]; then
						[[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
					fi
		
					# complete after --flag=abc
					if [[ $cur == *=* ]]; then
						if [[ $(type -t compopt) = "builtin" ]]; then
							compopt +o nospace
						fi
		
						local index flag
						flag="${cur%=*}"
						__contxt_index_of_word "${flag}" "${flags_with_completion[@]}"
						COMPREPLY=()
						if [[ ${index} -ge 0 ]]; then
							PREFIX=""
							cur="${cur#*=}"
							${flags_completion[${index}]}
							if [ -n "${ZSH_VERSION:-}" ]; then
								# zsh completion needs --flag= prefix
								eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
							fi
						fi
					fi
		
					if [[ -z "${flag_parsing_disabled}" ]]; then
						# If flag parsing is enabled, we have completed the flags and can return.
						# If flag parsing is disabled, we may not know all (or any) of the flags, so we fallthrough
						# to possibly call handle_go_custom_completion.
						return 0;
					fi
					;;
			esac
		
			# check if we are handling a flag with special work handling
			local index
			__contxt_index_of_word "${prev}" "${flags_with_completion[@]}"
			if [[ ${index} -ge 0 ]]; then
				${flags_completion[${index}]}
				return
			fi
		
			# we are parsing a flag and don't have a special handler, no completion
			if [[ ${cur} != "${words[cword]}" ]]; then
				return
			fi
		
			local completions
			completions=("${commands[@]}")
			if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
				completions+=("${must_have_one_noun[@]}")
			elif [[ -n "${has_completion_function}" ]]; then
				# if a go completion function is provided, defer to that function
				__contxt_handle_go_custom_completion
			fi
			if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
				completions+=("${must_have_one_flag[@]}")
			fi
			while IFS='' read -r comp; do
				COMPREPLY+=("$comp")
			done < <(compgen -W "${completions[*]}" -- "$cur")
		
			if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
				while IFS='' read -r comp; do
					COMPREPLY+=("$comp")
				done < <(compgen -W "${noun_aliases[*]}" -- "$cur")
			fi
		
			if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
				if declare -F __contxt_custom_func >/dev/null; then
					# try command name qualified custom func
					__contxt_custom_func
				else
					# otherwise fall back to unqualified for compatibility
					declare -F __custom_func >/dev/null && __custom_func
				fi
			fi
		
			# available in bash-completion >= 2, not always present on macOS
			if declare -F __ltrim_colon_completions >/dev/null; then
				__ltrim_colon_completions "$cur"
			fi
		
			# If there is only 1 completion and it is a flag with an = it will be completed
			# but we don't want a space after the =
			if [[ "${#COMPREPLY[@]}" -eq "1" ]] && [[ $(type -t compopt) = "builtin" ]] && [[ "${COMPREPLY[0]}" == --*= ]]; then
			   compopt -o nospace
			fi
		}`,
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
