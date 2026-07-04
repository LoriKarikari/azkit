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
	envPrefix = "AZKIT_"

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

	defaultDuration, err := durationValue(k, "pim.default_duration", DefaultActivationDuration)
	if err != nil {
		return nil, err
	}
	c := &Config{
		DefaultDuration: defaultDuration,
		SubscriptionID:  k.String("pim.subscription_id"),
		TenantID:        k.String("pim.tenant_id"),
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
	switch mapKey {
	case "pim_default_duration":
		if d, err := time.ParseDuration(value); err == nil {
			return "pim.default_duration", d
		}
		return "pim.default_duration", value
	case "pim_subscription_id":
		return "pim.subscription_id", value
	case "pim_tenant_id":
		return "pim.tenant_id", value
	default:
		return strings.ReplaceAll(mapKey, "_", "."), value
	}
}

func defaultConfigPath() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "azkit", "config.yaml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "azkit", "config.yaml")
}
