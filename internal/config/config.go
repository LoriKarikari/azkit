package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const (
	envPrefix = "PIMCTL_"

	// DefaultActivationDuration is the fallback Activation duration when neither flags nor Configuration set one.
	DefaultActivationDuration = 2 * time.Hour
)

type Config struct {
	DefaultDuration time.Duration `koanf:"default_duration"`
	SubscriptionID  string        `koanf:"subscription_id"`
	TenantID        string        `koanf:"tenant_id"`
}

// ActivationDuration applies the configured Activation duration unless the caller supplied one.
func (c *Config) ActivationDuration(input time.Duration) time.Duration {
	if input > 0 {
		return input
	}
	if c != nil && c.DefaultDuration > 0 {
		return c.DefaultDuration
	}
	return DefaultActivationDuration
}

// ActivationSubscription applies the configured Subscription selector unless the caller supplied a selector or Resource scope.
func (c *Config) ActivationSubscription(input string, scopeID string) string {
	if input != "" || scopeID != "" || c == nil {
		return input
	}
	return c.SubscriptionID
}

func Load(configPath string) (*Config, error) {
	k := koanf.New(".")
	hasExplicitPath := configPath != ""

	if configPath == "" {
		configPath = defaultConfigPath()
	}
	if configPath != "" {
		if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			if hasExplicitPath || !os.IsNotExist(err) {
				return nil, fmt.Errorf("config file %q: %w", configPath, err)
			}
		}
	}

	if err := k.Load(env.ProviderWithValue(envPrefix, ".", envMapper), nil); err != nil {
		return nil, fmt.Errorf("env vars: %w", err)
	}

	defaultDuration, err := durationValue(k, "default_duration", DefaultActivationDuration)
	if err != nil {
		return nil, err
	}
	c := &Config{
		DefaultDuration: defaultDuration,
		SubscriptionID:  k.String("subscription_id"),
		TenantID:        k.String("tenant_id"),
	}
	return c, nil
}

func durationValue(k *koanf.Koanf, key string, fallback time.Duration) (time.Duration, error) {
	v := k.Get(key)
	switch val := v.(type) {
	case nil:
		return fallback, nil
	case time.Duration:
		return val, nil
	case string:
		d, err := time.ParseDuration(val)
		if err != nil {
			return 0, fmt.Errorf("config %s: %w", key, err)
		}
		return d, nil
	}
	return 0, fmt.Errorf("config %s: expected duration, got %T", key, v)
}

func envMapper(key string, value string) (string, any) {
	mapKey := strings.ToLower(strings.TrimPrefix(key, envPrefix))
	if mapKey == "default_duration" {
		if d, err := time.ParseDuration(value); err == nil {
			return mapKey, d
		}
	}
	return mapKey, value
}

func defaultConfigPath() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		p := filepath.Join(dir, "pimctl", "config.yaml")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	p := filepath.Join(home, ".config", "pimctl", "config.yaml")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}
