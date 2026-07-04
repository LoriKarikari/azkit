package cli

import (
	"fmt"
	"io"
)

type ShellInitCmd struct {
	Shell string `arg:"" enum:"bash,zsh,powershell,pwsh" help:"Shell to initialize: bash, zsh, or powershell"`
}

func (c *ShellInitCmd) Run(streams *Streams) error {
	script, err := shellInitScript(c.Shell)
	if err != nil {
		return err
	}
	_, err = io.WriteString(streams.Stdout, script)
	return err
}

func shellInitScript(shell string) (string, error) {
	switch shell {
	case "bash":
		return posixShellInit("bash"), nil
	case "zsh":
		return posixShellInit("zsh"), nil
	case "powershell", "pwsh":
		return powershellInit(), nil
	default:
		return "", fmt.Errorf("unsupported shell %q", shell)
	}
}

func posixShellInit(shell string) string {
	return fmt.Sprintf(`# azkit shell integration for %[1]s
azkit() {
  if [ "$#" -gt 0 ] && __azkit_needs_shell "$@"; then
    local __azkit_output
    __azkit_output="$(AZKIT_SHELL=%[1]s command azkit --shell-env "$@")"
    local __azkit_status=$?
    if [ $__azkit_status -ne 0 ]; then
      return $__azkit_status
    fi
    if [ -n "$__azkit_output" ]; then
      eval "$__azkit_output"
    fi
    return 0
  fi
  command azkit "$@"
}

__azkit_needs_shell() {
  case "$1" in
    ctx)
      case "${2-}" in
        add|rm|current|-l|--list|--help|-h) return 1 ;;
        *) return 0 ;;
      esac
      ;;
    sub)
      case "${2-}" in
        -l|--list|--refresh|--help|-h) return 1 ;;
        *) return 0 ;;
      esac
      ;;
  esac
  return 1
}
`, shell)
}

func powershellInit() string {
	return `# azkit shell integration for PowerShell
$script:AzkitNativeCommand = (Get-Command azkit -CommandType Application | Select-Object -First 1).Source

function global:azkit {
  if ($args.Count -gt 0 -and (Test-AzkitShellSwitch $args)) {
    $__azkitOldShell = $env:AZKIT_SHELL
    try {
      $env:AZKIT_SHELL = "powershell"
      $__azkitOutput = & $script:AzkitNativeCommand --shell-env @args
      if ($LASTEXITCODE -ne 0) {
        return
      }
      if ($__azkitOutput) {
        Invoke-Expression ($__azkitOutput -join [Environment]::NewLine)
      }
      return
    } finally {
      if ($null -eq $__azkitOldShell) {
        Remove-Item Env:AZKIT_SHELL -ErrorAction SilentlyContinue
      } else {
        $env:AZKIT_SHELL = $__azkitOldShell
      }
    }
  }
  & $script:AzkitNativeCommand @args
}

function global:Test-AzkitShellSwitch {
  param([object[]]$AzkitArgs)

  if ($AzkitArgs.Count -eq 0) {
    return $false
  }

  switch ($AzkitArgs[0]) {
    "ctx" {
      if ($AzkitArgs.Count -eq 1) { return $true }
      switch ($AzkitArgs[1]) {
        "add" { return $false }
        "rm" { return $false }
        "current" { return $false }
        "-l" { return $false }
        "--list" { return $false }
        "--help" { return $false }
        "-h" { return $false }
        default { return $true }
      }
    }
    "sub" {
      if ($AzkitArgs.Count -eq 1) { return $true }
      switch ($AzkitArgs[1]) {
        "-l" { return $false }
        "--list" { return $false }
        "--refresh" { return $false }
        "--help" { return $false }
        "-h" { return $false }
        default { return $true }
      }
    }
    default { return $false }
  }
}
`
}
