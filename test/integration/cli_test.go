//go:build integration

package integration_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPimctlBinary(t *testing.T) {
	bin := buildPimctl(t)

	tests := []struct {
		name         string
		args         []string
		wantCode     int
		wantStdout   string
		wantStderr   string
		wantNoStdout bool
	}{
		{
			name:       "version",
			args:       []string{"version"},
			wantCode:   0,
			wantStdout: "pimctl dev",
		},
		{
			name:         "usage error exits two",
			args:         []string{"not-a-command"},
			wantCode:     2,
			wantStderr:   "unexpected argument not-a-command",
			wantNoStdout: true,
		},
		{
			name:         "activate validates before azure auth",
			args:         []string{"activate", "--scope", "/subscriptions/sub-a", "--reason", "break glass"},
			wantCode:     1,
			wantStderr:   "Activation role is required.",
			wantNoStdout: true,
		},
		{
			name:         "deactivate requires assignment id outside a terminal",
			args:         []string{"deactivate"},
			wantCode:     1,
			wantStderr:   "Assignment ID is required.",
			wantNoStdout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, code := runPimctl(t, bin, tt.args...)
			if code != tt.wantCode {
				t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, tt.wantCode, stdout, stderr)
			}
			if tt.wantStdout != "" && !strings.Contains(stdout, tt.wantStdout) {
				t.Fatalf("stdout missing %q\nstdout:\n%s", tt.wantStdout, stdout)
			}
			if tt.wantStderr != "" && !strings.Contains(stderr, tt.wantStderr) {
				t.Fatalf("stderr missing %q\nstderr:\n%s", tt.wantStderr, stderr)
			}
			if tt.wantNoStdout && strings.TrimSpace(stdout) != "" {
				t.Fatalf("want empty stdout, got %q", stdout)
			}
		})
	}
}

func buildPimctl(t *testing.T) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "pimctl")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/pimctl")
	cmd.Dir = repoRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build pimctl: %v\n%s", err, output)
	}
	return bin
}

func runPimctl(t *testing.T, bin string, args ...string) (string, string, int) {
	t.Helper()

	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(),
		"NO_COLOR=1",
		"XDG_CONFIG_HOME="+t.TempDir(),
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		return stdout.String(), stderr.String(), 0
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return stdout.String(), stderr.String(), exitErr.ExitCode()
	}
	t.Fatalf("run pimctl %v: %v", args, err)
	return "", "", 0
}

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root")
		}
		dir = parent
	}
}
