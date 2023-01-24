#
# NoteBot Justfile
#
APP_NAME := 'Krait'
CLI_NAME := 'krait'
MAIN_CODE := 'krait.go'

alias ver := version

shebang := if os() == 'windows' { 'powershell' } else { '/bin/sh' } # Shebang for PS Desktop on Windows and PS Core everywhere else
set windows-shell := ["powershell", "-c"] # To use PowerShell Desktop instead of Core on Windows
# set shell := ["pwsh", "-c"] # PowerShell Core (Multi-Platform)

set positional-arguments


@_default: _term-wipe
	just --list

@args *args:
	echo "\$# = $#"
	echo "\$@ = $@"

# Run code with (optional) arguments
run *args: _term-wipe
	go run cmd/myapp/myapp.go {{args}}


# Set Terminal Title
_term-title title='':
	#!/bin/sh
	printf "\033]0;%s\007" "{{title}}"

# Wipe Terminal Buffer and Scrollback Buffer
@_term-wipe:
	#!/bin/sh
	if [[ ${#VISUAL_STUDIO_CODE} -gt 0 ]]; then
		clear
	elif [[ ${KITTY_WINDOW_ID} -gt 0 ]] || [[ ${#TMUX} -gt 0 ]] || [[ "${TERM_PROGRAM}" = 'vscode' ]]; then
		printf '\033c'
	elif [[ "$(uname)" == 'Darwin' ]] || [[ "${TERM_PROGRAM}" = 'Apple_Terminal' ]] || [[ "${TERM_PROGRAM}" = 'iTerm.app' ]]; then
		osascript -e 'tell application "System Events" to keystroke "k" using command down'
	elif [[ -x "$(which tput)" ]]; then
		tput reset
	elif [[ -x "$(which reset)" ]]; then
		reset
	else
		clear
	fi

	# Clear-Host


# Unit Test Code
test: _term-wipe
	@# go test ./...
	go test .

tester *args: _term-wipe
	#!/bin/sh
	echo

	if [ $# -lt 1 ]; then
		echo "==> Tester with no args"
		go run cmd/tester/tester.go
	else
		echo "==> Tester with $# args: $@"
		go run cmd/tester/tester.go "$@"
	fi

	echo


# Display version of app
@version:
	cat "{{MAIN_CODE}}" | grep AppVersion | head -1 | cut -d"'" -f2
	# ((Get-Content {{MAIN_CODE}} | Select-String 'AppVersion') -Split "'")[1]

