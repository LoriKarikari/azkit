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
complete -C pimctl pimctl
`

const zshCompletion = `#compdef pimctl
autoload -U +X bashcompinit && bashcompinit
complete -o nospace -C pimctl pimctl
`

const fishCompletion = `function __complete_pimctl
    set -lx COMP_LINE (commandline -cp)
    test -z (commandline -ct)
    and set COMP_LINE "$COMP_LINE "
    set -lx COMP_POINT (string length -- "$COMP_LINE")
    pimctl
end
complete -f -c pimctl -a "(__complete_pimctl)"
`

const powershellCompletion = `# PowerShell completion for pimctl
Register-ArgumentCompleter -Native -CommandName pimctl -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    $oldLine = $env:COMP_LINE
    $oldPoint = $env:COMP_POINT
    try {
        $env:COMP_LINE = $commandAst.ToString()
        $env:COMP_POINT = [string]$cursorPosition
        pimctl | ForEach-Object {
            if ($_ -like "$wordToComplete*") {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
        }
    } finally {
        $env:COMP_LINE = $oldLine
        $env:COMP_POINT = $oldPoint
    }
}
`
