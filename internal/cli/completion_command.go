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
    local cur command opts
    cur="${COMP_WORDS[COMP_CWORD]}"
    command="${COMP_WORDS[1]}"

    case "${command}" in
        list|status)
            opts="--json --extended --help"
            ;;
        activate)
            opts="--scope --subscription --resource-group --role --reason --duration --json --help"
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish powershell" -- "${cur}") )
            return 0
            ;;
        *)
            opts="list status activate completion --verbose --config --help"
            ;;
    esac

    COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
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
                list|status)
                    _values 'flags' '--json' '--extended' '--help'
                    ;;
                activate)
                    _values 'flags' '--scope' '--subscription' '--resource-group' '--role' '--reason' '--duration' '--json' '--help'
                    ;;
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
complete -c pimctl -n "__fish_seen_subcommand_from list status" -l json -d "Output as JSON"
complete -c pimctl -n "__fish_seen_subcommand_from list status" -l extended -d "Show more details"
complete -c pimctl -n "__fish_seen_subcommand_from activate" -l scope -d "Azure resource scope ID"
complete -c pimctl -n "__fish_seen_subcommand_from activate" -l subscription -d "Subscription ID or exact name"
complete -c pimctl -n "__fish_seen_subcommand_from activate" -l resource-group -d "Resource group name"
complete -c pimctl -n "__fish_seen_subcommand_from activate" -l role -d "Role display name or definition ID"
complete -c pimctl -n "__fish_seen_subcommand_from activate" -l reason -d "Justification for the activation"
complete -c pimctl -n "__fish_seen_subcommand_from activate" -l duration -d "Activation duration"
complete -c pimctl -n "__fish_seen_subcommand_from activate" -l json -d "Output as JSON"
complete -c pimctl -n "__fish_seen_subcommand_from completion" -a "bash zsh fish powershell"
`

const powershellCompletion = `# PowerShell completion for pimctl
Register-ArgumentCompleter -Native -CommandName pimctl -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $commands = @("list", "status", "activate", "completion")
    $globalFlags = @("--verbose", "--config", "--help")
    $commandFlags = @{
        list = @("--json", "--extended", "--help")
        status = @("--json", "--extended", "--help")
        activate = @("--scope", "--subscription", "--resource-group", "--role", "--reason", "--duration", "--json", "--help")
        completion = @("bash", "zsh", "fish", "powershell")
    }

    $tokens = $commandAst.ToString().Split()
    $command = if ($tokens.Count -gt 1) { $tokens[1] } else { "" }

    if ($wordToComplete -like "--*") {
        $options = if ($commandFlags.ContainsKey($command)) { $commandFlags[$command] } else { $globalFlags }
        $options | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
        }
        return
    }

    if ($command -eq "completion") {
        $commandFlags.completion | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
        }
        return
    }

    if ($tokens.Count -eq 1 -or ($tokens.Count -eq 2 -and $wordToComplete -ne "")) {
        $commands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'Command', $_)
        }
    }
}
`
