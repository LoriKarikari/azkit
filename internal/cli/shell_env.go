package cli

import (
	"fmt"
	"strings"

	"github.com/LoriKarikari/azkit/internal/app"
)

type ShellEnvChange struct {
	Name  string
	Value string
	Unset bool
}

func (s *Streams) RequireShellIntegration(command string) error {
	if s != nil && s.ShellEnv {
		return nil
	}
	return app.ShellIntegrationRequired(command)
}

func (s *Streams) RenderShellEnv(changes []ShellEnvChange) (string, error) {
	if s == nil {
		return renderShellEnv("", changes)
	}
	return renderShellEnv(s.Shell, changes)
}

func renderShellEnv(shell string, changes []ShellEnvChange) (string, error) {
	switch shell {
	case "", "bash", "zsh":
		return renderPOSIXShellEnv(changes), nil
	case "powershell", "pwsh":
		return renderPowerShellEnv(changes), nil
	default:
		return "", fmt.Errorf("unsupported shell %q", shell)
	}
}

func renderPOSIXShellEnv(changes []ShellEnvChange) string {
	var b strings.Builder
	for _, change := range changes {
		if change.Unset {
			fmt.Fprintf(&b, "unset %s\n", change.Name)
			continue
		}
		fmt.Fprintf(&b, "export %s=%s\n", change.Name, quotePOSIX(change.Value))
	}
	return b.String()
}

func renderPowerShellEnv(changes []ShellEnvChange) string {
	var b strings.Builder
	for _, change := range changes {
		if change.Unset {
			fmt.Fprintf(&b, "Remove-Item Env:%s -ErrorAction SilentlyContinue\n", change.Name)
			continue
		}
		fmt.Fprintf(&b, "$env:%s = %s\n", change.Name, quotePowerShell(change.Value))
	}
	return b.String()
}

func quotePOSIX(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func quotePowerShell(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
