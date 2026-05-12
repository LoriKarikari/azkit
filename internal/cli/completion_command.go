package cli

import (
	"fmt"
	"io"
	"strings"
)

type CompletionCmd struct {
	Shell string `arg:"" required:"" help:"Shell to generate completions for (bash, zsh, fish, powershell)"`
}

func (c *CompletionCmd) Run(streams *Streams) error {
	script, err := completionScript(strings.ToLower(c.Shell))
	if err != nil {
		return err
	}
	_, err = io.WriteString(streams.Stdout, script)
	return err
}

func completionScript(shell string) (string, error) {
	switch shell {
	case "bash":
		return bashCompletion, nil
	case "zsh":
		return zshCompletion, nil
	case "fish":
		return fishCompletion, nil
	case "powershell", "pwsh":
		return powershellCompletion, nil
	default:
		return "", fmt.Errorf("unknown shell %q: supported shells are bash, zsh, fish, powershell", shell)
	}
}

const bashCompletion = `#!/usr/bin/env bash
_pimctl_completions() {
    local cur prev opts
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="list status activate completion --verbose --config --help"

    if [[ ${cur} == -* ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi

    case "${prev}" in
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish powershell" -- "${cur}") )
            return 0
            ;;
    esac
}
complete -F _pimctl_completions pimctl
`

const zshCompletion = `#compdef pimctl
_pimctl() {
    local curcontext="$curcontext" state line
    typeset -A opt_args

    _arguments -C \
        '(-v --verbose)'{-v,--verbose}'[Enable debug logging to stderr]' \
        '(--config)'--config'[Path to config file]:config:_files' \
        '1: :->command' \
        '*:: :->args'

    case "$state" in
        command)
            _values 'command' \
                'list[List eligible PIM role assignments]' \
                'status[List active PIM role assignments]' \
                'activate[Activate an eligible PIM role assignment]' \
                'completion[Generate shell completion script]'
            ;;
        args)
            case "$line[1]" in
                completion)
                    _values 'shell' 'bash' 'zsh' 'fish' 'powershell'
                    ;;
            esac
            ;;
    esac
}
compdef _pimctl pimctl
`

const fishCompletion = `complete -c pimctl -f
complete -c pimctl -s v -l verbose -d "Enable debug logging to stderr"
complete -c pimctl -l config -d "Path to config file"
complete -c pimctl -n "not __fish_seen_subcommand_from list status activate completion" -a "list" -d "List eligible PIM role assignments"
complete -c pimctl -n "not __fish_seen_subcommand_from list status activate completion" -a "status" -d "List active PIM role assignments"
complete -c pimctl -n "not __fish_seen_subcommand_from list status activate completion" -a "activate" -d "Activate an eligible PIM role assignment"
complete -c pimctl -n "not __fish_seen_subcommand_from list status activate completion" -a "completion" -d "Generate shell completion script"
complete -c pimctl -n "__fish_seen_subcommand_from completion" -a "bash zsh fish powershell"
`

const powershellCompletion = `# PowerShell completion for pimctl
Register-ArgumentCompleter -Native -CommandName pimctl -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $commands = @("list", "status", "activate", "completion")
    $flags = @("--verbose", "--config", "--help")

    if ($wordToComplete -like "--*") {
        $flags | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
        }
        return
    }

    $tokens = $commandAst.ToString().Split()
    if ($tokens.Count -eq 1 -or ($tokens.Count -eq 2 -and $wordToComplete -ne "")) {
        $commands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'Command', $_)
        }
    }
}
`
