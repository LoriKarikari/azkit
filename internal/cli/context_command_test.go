package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LoriKarikari/azkit/internal/interactive"
)

func TestRunner_ctxAddAndList(t *testing.T) {
	setupContextDirs(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	code := runner.Run(t.Context(), []string{"ctx", "add", "prod", "--tenant", "tenant-a"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Added context prod for tenant tenant-a") {
		t.Fatalf("missing add output: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = runner.Run(t.Context(), []string{"ctx", "-l"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	got := stdout.String()
	for _, want := range []string{"NAME", "TENANT", "STATUS", "prod", "tenant-a", "needs_login"} {
		if !strings.Contains(got, want) {
			t.Fatalf("ctx list missing %q:\n%s", want, got)
		}
	}
}

func TestRunner_ctxAddDiscoversTenantFromEnv(t *testing.T) {
	setupContextDirs(t)
	t.Setenv("AZURE_TENANT_ID", "tenant-from-env")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	code := runner.Run(t.Context(), []string{"ctx", "add", "dev"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "tenant-from-env") {
		t.Fatalf("missing discovered tenant in output: %s", stdout.String())
	}
}

func TestRunner_ctxListStatuses(t *testing.T) {
	_, stateRoot := setupContextDirs(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	if code := runner.Run(t.Context(), []string{"ctx", "add", "ready", "--tenant", "tenant-ready"}); code != 0 {
		t.Fatalf("add ready: exit %d: %s", code, stderr.String())
	}
	readyDir := filepath.Join(stateRoot, "azkit", "contexts", "ready")
	if err := os.WriteFile(filepath.Join(readyDir, "azureProfile.json"), []byte("{}"), 0600); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := runner.Run(t.Context(), []string{"ctx", "add", "missing", "--tenant", "tenant-missing"}); code != 0 {
		t.Fatalf("add missing: exit %d: %s", code, stderr.String())
	}
	missingDir := filepath.Join(stateRoot, "azkit", "contexts", "missing")
	if err := os.RemoveAll(missingDir); err != nil {
		t.Fatalf("remove missing dir: %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	code := runner.Run(t.Context(), []string{"ctx", "-l"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	got := stdout.String()
	for _, want := range []string{"ready", "tenant-ready", "ready", "missing", "tenant-missing", "missing_dir"} {
		if !strings.Contains(got, want) {
			t.Fatalf("ctx list missing %q:\n%s", want, got)
		}
	}
}

func TestRunner_ctxRemoveForceDeletesCatalogEntryAndCache(t *testing.T) {
	_, stateRoot := setupContextDirs(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	if code := runner.Run(t.Context(), []string{"ctx", "add", "prod", "--tenant", "tenant-a"}); code != 0 {
		t.Fatalf("add: exit %d: %s", code, stderr.String())
	}
	cacheDir := filepath.Join(stateRoot, "azkit", "contexts", "prod")
	if _, err := os.Stat(cacheDir); err != nil {
		t.Fatalf("cache dir missing before remove: %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	code := runner.Run(t.Context(), []string{"ctx", "rm", "prod", "--force"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Fatalf("want cache dir removed, stat err=%v", err)
	}

	stdout.Reset()
	stderr.Reset()
	code = runner.Run(t.Context(), []string{"ctx", "-l"})
	if code != 0 {
		t.Fatalf("list: exit %d: %s", code, stderr.String())
	}
	if stdout.String() != "No contexts.\n" {
		t.Fatalf("want empty catalog, got %q", stdout.String())
	}
}

func TestRunner_ctxRemoveActiveRequiresForce(t *testing.T) {
	setupContextDirs(t)
	t.Setenv("AZKIT_CONTEXT", "prod")
	interactive.IsTerminalFn = func() bool { return false }
	t.Cleanup(func() { interactive.IsTerminalFn = interactive.IsTerminal })
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	if code := runner.Run(t.Context(), []string{"ctx", "add", "prod", "--tenant", "tenant-a"}); code != 0 {
		t.Fatalf("add: exit %d: %s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code := runner.Run(t.Context(), []string{"ctx", "rm", "prod"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "is active") {
		t.Fatalf("want active-context error, got: %s", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = runner.Run(t.Context(), []string{"ctx", "rm", "prod", "--force"})
	if code != 0 {
		t.Fatalf("force remove: exit %d: %s", code, stderr.String())
	}
}

func TestRunner_ctxRemoveNeedsForceOutsideTerminal(t *testing.T) {
	setupContextDirs(t)
	interactive.IsTerminalFn = func() bool { return false }
	t.Cleanup(func() { interactive.IsTerminalFn = interactive.IsTerminal })
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	if code := runner.Run(t.Context(), []string{"ctx", "add", "prod", "--tenant", "tenant-a"}); code != 0 {
		t.Fatalf("add: exit %d: %s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code := runner.Run(t.Context(), []string{"ctx", "rm", "prod"})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "requires confirmation") {
		t.Fatalf("want confirmation error, got: %s", stderr.String())
	}
}

func TestRunner_ctxRejectsInvalidNames(t *testing.T) {
	setupContextDirs(t)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)

	for _, name := range []string{"1prod", "prod.example", "add"} {
		t.Run(name, func(t *testing.T) {
			stdout.Reset()
			stderr.Reset()
			code := runner.Run(t.Context(), []string{"ctx", "add", name, "--tenant", "tenant-a"})
			if code != 1 {
				t.Fatalf("want exit 1, got %d", code)
			}
			if !strings.Contains(stderr.String(), "Invalid context name") {
				t.Fatalf("want invalid name error, got: %s", stderr.String())
			}
		})
	}
}

func TestRunner_ctxIgnoresInvalidPimConfig(t *testing.T) {
	configRoot, _ := setupContextDirs(t)
	configPath := filepath.Join(configRoot, "azkit", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("invalid: [unclosed"), 0600); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runner := newRunner(&stdout, &stderr, nil)
	code := runner.Run(t.Context(), []string{"--config", configPath, "ctx", "-l"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, stderr.String())
	}
}

func setupContextDirs(t *testing.T) (string, string) {
	t.Helper()
	configRoot := t.TempDir()
	stateRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	t.Setenv("XDG_STATE_HOME", stateRoot)
	t.Setenv("HOME", t.TempDir())
	return configRoot, stateRoot
}
