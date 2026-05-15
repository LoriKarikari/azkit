package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LoriKarikari/pimctl/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	c, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.DefaultDuration != 2*time.Hour {
		t.Fatalf("want 2h default, got %v", c.DefaultDuration)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	contents := "default_duration: 30m\nsubscription_id: sub-file\n"
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
	if c.SubscriptionID != "sub-file" {
		t.Fatalf("want subscription_id sub-file, got %s", c.SubscriptionID)
	}
}

func TestLoadEnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	contents := "default_duration: 30m\nsubscription_id: sub-file\ntenant_id: tenant-file\n"
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("PIMCTL_DEFAULT_DURATION", "1h")
	t.Setenv("PIMCTL_SUBSCRIPTION_ID", "sub-env")
	t.Setenv("PIMCTL_TENANT_ID", "1h")

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
	if c.TenantID != "1h" {
		t.Fatalf("want tenant id to remain a string, got %q", c.TenantID)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
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

func TestLoadExplicitMissingFileFails(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("want missing explicit file error, got nil")
	}
}

func TestLoadEnvOnlyDuration(t *testing.T) {
	t.Setenv("PIMCTL_DEFAULT_DURATION", "90m")

	c, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.DefaultDuration != 90*time.Minute {
		t.Fatalf("want 90m, got %v", c.DefaultDuration)
	}
}

func TestLoadInvalidDurationFails(t *testing.T) {
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

func TestActivationDefaults(t *testing.T) {
	c := &config.Config{
		DefaultDuration: 45 * time.Minute,
		SubscriptionID:  "sub-default",
	}

	if got := c.ActivationDuration(10 * time.Minute); got != 10*time.Minute {
		t.Fatalf("want explicit duration, got %v", got)
	}
	if got := c.ActivationDuration(0); got != 45*time.Minute {
		t.Fatalf("want configured duration, got %v", got)
	}
	if got := (*config.Config)(nil).ActivationDuration(0); got != config.DefaultActivationDuration {
		t.Fatalf("want package default duration, got %v", got)
	}

	if got := c.ActivationSubscription("sub-explicit", ""); got != "sub-explicit" {
		t.Fatalf("want explicit subscription, got %q", got)
	}
	if got := c.ActivationSubscription("", "/subscriptions/sub-scope"); got != "" {
		t.Fatalf("want no default when scope is explicit, got %q", got)
	}
	if got := c.ActivationSubscription("", ""); got != "sub-default" {
		t.Fatalf("want configured subscription, got %q", got)
	}
}
