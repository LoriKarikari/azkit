package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LoriKarikari/pimctl/internal/config"
)

func TestLoad_defaults(t *testing.T) {
	c, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.DefaultDuration != 2*time.Hour {
		t.Fatalf("want 2h default, got %v", c.DefaultDuration)
	}
	if c.NoColor {
		t.Fatal("want no_color false by default")
	}
	if c.TenantID != "" {
		t.Fatal("want empty tenant_id by default")
	}
}

func TestLoad_fromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	contents := "default_duration: 30m\nno_color: true\ntenant_id: abc-123\n"
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	c, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.DefaultDuration != 30*time.Minute {
		t.Fatalf("want 30m, got %v", c.DefaultDuration)
	}
	if !c.NoColor {
		t.Fatal("want no_color true")
	}
	if c.TenantID != "abc-123" {
		t.Fatalf("want tenant_id abc-123, got %s", c.TenantID)
	}
}

func TestLoad_envOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	contents := "default_duration: 30m\nsubscription_id: sub-file\n"
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("PIMCTL_DEFAULT_DURATION", "1h")
	t.Setenv("PIMCTL_SUBSCRIPTION_ID", "sub-env")

	c, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.DefaultDuration != 1*time.Hour {
		t.Fatalf("want 1h env override, got %v", c.DefaultDuration)
	}
	if c.SubscriptionID != "sub-env" {
		t.Fatalf("want sub-env from env, got %s", c.SubscriptionID)
	}
}

func TestLoad_invalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("invalid: [unclosed"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("want error, got nil")
	}
}

func TestLoad_explicitMissingFileFails(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("want missing explicit file error, got nil")
	}
}

func TestLoad_envOnlyDuration(t *testing.T) {
	t.Setenv("PIMCTL_DEFAULT_DURATION", "90m")

	c, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.DefaultDuration != 90*time.Minute {
		t.Fatalf("want 90m, got %v", c.DefaultDuration)
	}
}

func TestLoad_invalidDurationFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("default_duration: nope\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("want invalid duration error, got nil")
	}
}

func TestLoad_invalidBoolFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("no_color: sometimes\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("want invalid bool error, got nil")
	}
}
