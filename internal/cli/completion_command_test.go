package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/LoriKarikari/pimctl/internal/cli"
)

func TestCompletionBash(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := cli.NewRunner(cli.Services{}, &stdout, &stderr)

	code := runner.Run(t.Context(), []string{"completion", "bash"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	got := stdout.String()
	if got == "" {
		t.Fatal("want non-empty bash script, got empty")
	}
	if !strings.Contains(got, "pimctl") {
		t.Fatal("want script referencing pimctl")
	}
	if !strings.Contains(got, "complete -F _pimctl_completions pimctl") {
		t.Fatal("want bash complete registration")
	}
}

func TestCompletionZsh(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := cli.NewRunner(cli.Services{}, &stdout, &stderr)

	code := runner.Run(t.Context(), []string{"completion", "zsh"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	got := stdout.String()
	if !strings.Contains(got, "pimctl") {
		t.Fatal("want script referencing pimctl")
	}
	if !strings.Contains(got, "#compdef pimctl") {
		t.Fatal("want zsh compdef")
	}
}

func TestCompletionFish(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := cli.NewRunner(cli.Services{}, &stdout, &stderr)

	code := runner.Run(t.Context(), []string{"completion", "fish"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	got := stdout.String()
	if !strings.Contains(got, "pimctl") {
		t.Fatal("want script referencing pimctl")
	}
	if !strings.Contains(got, "complete -c pimctl") {
		t.Fatal("want fish complete command")
	}
}

func TestCompletionPowerShell(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := cli.NewRunner(cli.Services{}, &stdout, &stderr)

	code := runner.Run(t.Context(), []string{"completion", "powershell"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	got := stdout.String()
	if !strings.Contains(got, "pimctl") {
		t.Fatal("want script referencing pimctl")
	}
	if !strings.Contains(got, "Register-ArgumentCompleter") {
		t.Fatal("want PowerShell argument completer")
	}
}

func TestCompletionUnknownShell(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := cli.NewRunner(cli.Services{}, &stdout, &stderr)

	code := runner.Run(t.Context(), []string{"completion", "nushell"})
	if code == 0 {
		t.Fatal("want error for unknown shell")
	}
	if !strings.Contains(stderr.String(), "unknown shell") {
		t.Fatalf("want unknown shell error, got: %s", stderr.String())
	}
}
