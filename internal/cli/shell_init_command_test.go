package cli_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/cli"
	"github.com/LoriKarikari/azkit/internal/domain"
)

func TestRunner_shellInit(t *testing.T) {
	tests := []struct {
		name  string
		shell string
		want  []string
	}{
		{
			name:  "bash",
			shell: "bash",
			want: []string{
				"azkit() {",
				"AZKIT_SHELL=bash command azkit --shell-env",
				"command azkit \"$@\"",
				"ctx)",
				"sub)",
			},
		},
		{
			name:  "zsh",
			shell: "zsh",
			want: []string{
				"azkit() {",
				"AZKIT_SHELL=zsh command azkit --shell-env",
				"command azkit \"$@\"",
			},
		},
		{
			name:  "powershell",
			shell: "powershell",
			want: []string{
				"function global:azkit",
				`$env:AZKIT_SHELL = "powershell"`,
				"--shell-env @args",
				"Test-AzkitShellSwitch",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			runner := newRunner(&stdout, &stderr, nil)

			code := runner.Run(t.Context(), []string{"shell-init", tt.shell})
			if code != 0 {
				t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
			}
			if stderr.String() != "" {
				t.Fatalf("want empty stderr, got %q", stderr.String())
			}
			got := stdout.String()
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Fatalf("shell-init output missing %q:\n%s", want, got)
				}
			}
			if strings.Contains(got, "--json") {
				t.Fatalf("shell env path should not use --json:\n%s", got)
			}
		})
	}
}

func TestRunner_shellInitIgnoresInvalidConfig(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)
	configPath := writeRunnerConfig(t, "invalid: [unclosed")

	code := runner.Run(t.Context(), []string{"--config", configPath, "shell-init", "bash"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "azkit() {") {
		t.Fatalf("missing shell init output:\n%s", stdout.String())
	}
}

func TestRunner_shellInitRejectsUnsupportedShell(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	code := runner.Run(t.Context(), []string{"shell-init", "fish"})
	if code != 2 {
		t.Fatalf("want exit 2, got %d", code)
	}
	if stdout.String() != "" {
		t.Fatalf("want empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "must be one of") {
		t.Fatalf("want enum parse error, got %s", stderr.String())
	}
}

func TestStreamsRequireShellIntegration(t *testing.T) {
	if err := (&cli.Streams{ShellEnv: true}).RequireShellIntegration("azkit ctx prod"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := (&cli.Streams{}).RequireShellIntegration("azkit ctx prod")
	var appErr *app.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("want app error, got %T", err)
	}
	if appErr.Code != domain.CodeShellIntegrationRequired {
		t.Fatalf("want shell integration code, got %s", appErr.Code)
	}
}

func TestStreamsRenderShellEnv(t *testing.T) {
	changes := []cli.ShellEnvChange{
		{Name: "AZURE_TENANT_ID", Value: "tenant'one"},
		{Name: "AZURE_SUBSCRIPTION_ID", Unset: true},
	}

	posix, err := (&cli.Streams{Shell: "bash"}).RenderShellEnv(changes)
	if err != nil {
		t.Fatalf("unexpected POSIX render error: %v", err)
	}
	if !strings.Contains(posix, `export AZURE_TENANT_ID='tenant'"'"'one'`) {
		t.Fatalf("unexpected POSIX output:\n%s", posix)
	}
	if !strings.Contains(posix, "unset AZURE_SUBSCRIPTION_ID") {
		t.Fatalf("missing unset in POSIX output:\n%s", posix)
	}

	powershell, err := (&cli.Streams{Shell: "powershell"}).RenderShellEnv(changes)
	if err != nil {
		t.Fatalf("unexpected PowerShell render error: %v", err)
	}
	if !strings.Contains(powershell, `$env:AZURE_TENANT_ID = 'tenant''one'`) {
		t.Fatalf("unexpected PowerShell output:\n%s", powershell)
	}
	if !strings.Contains(powershell, "Remove-Item Env:AZURE_SUBSCRIPTION_ID") {
		t.Fatalf("missing unset in PowerShell output:\n%s", powershell)
	}
}
