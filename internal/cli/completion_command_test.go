package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LoriKarikari/azkit/internal/cli"
)

func TestCompletionUsesKongplete(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := cli.NewRunner(cli.Services{}, &stdout, &stderr)

	code := runner.Run(t.Context(), []string{"completion"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	got := stdout.String()
	if !strings.Contains(got, "complete -C ") {
		t.Fatalf("want kongplete completion script, got: %s", got)
	}
	if !strings.Contains(got, "azkit") {
		t.Fatalf("want script referencing azkit, got: %s", got)
	}
}

func TestCompletionRejectsShellArgument(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := cli.NewRunner(cli.Services{}, &stdout, &stderr)

	code := runner.Run(t.Context(), []string{"completion", "bash"})
	if code != 2 {
		t.Fatalf("want exit 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unexpected argument bash") {
		t.Fatalf("want unexpected argument error, got: %s", stderr.String())
	}
}

func TestKongpleteCompletesCommands(t *testing.T) {
	t.Setenv("COMP_LINE", "azkit ")
	t.Setenv("COMP_POINT", "6")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := cli.NewRunner(cli.Services{}, &stdout, &stderr)

	code := runner.Run(t.Context(), []string{})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	got := stdout.String()
	if !strings.Contains(got, "pim") || !strings.Contains(got, "completion") {
		t.Fatalf("want command completions, got: %s", got)
	}
	if strings.Contains(got, "activate") {
		t.Fatalf("root PIM command should not be completed, got: %s", got)
	}
}

func TestCompletionHelpDoesNotExposeUnsupportedUninstallFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := cli.NewRunner(cli.Services{}, &stdout, &stderr)

	code := runner.Run(t.Context(), []string{"completion", "--help"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if strings.Contains(stdout.String(), "--uninstall") {
		t.Fatalf("unsupported uninstall flag should be hidden, got: %s", stdout.String())
	}
}

func TestCompletionIgnoresInvalidConfig(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := cli.NewRunner(cli.Services{}, &stdout, &stderr)
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte("invalid: [unclosed"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	code := runner.Run(t.Context(), []string{"--config", configPath, "completion"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "complete -C ") {
		t.Fatalf("want completion script, got: %s", stdout.String())
	}
}
