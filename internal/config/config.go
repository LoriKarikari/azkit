package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const envPrefix = "PIMCTL_"

type Config struct {
	DefaultDuration time.Duration `koanf:"default_duration"`
	TenantID        string        `koanf:"tenant_id"`
	SubscriptionID  string        `koanf:"subscription_id"`
	IsColorDisabled bool          `koanf:"no_color"`
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

	defaultDuration, err := durationValue(k, "default_duration", 2*time.Hour)
	if err != nil {
		return nil, err
	}
	isColorDisabled, err := boolValue(k, "no_color")
	if err != nil {
		return nil, err
	}

	c := &Config{
		DefaultDuration: defaultDuration,
		TenantID:        k.String("tenant_id"),
		SubscriptionID:  k.String("subscription_id"),
		IsColorDisabled: isColorDisabled,
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

func boolValue(k *koanf.Koanf, key string) (bool, error) {
	v := k.Get(key)
	switch val := v.(type) {
	case nil:
		return false, nil
	case bool:
		return val, nil
	case string:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return false, fmt.Errorf("config %s: %w", key, err)
		}
		return b, nil
	}
	return false, fmt.Errorf("config %s: expected bool, got %T", key, v)
}

func envMapper(key string, value string) (string, any) {
	mapKey := strings.ToLower(strings.TrimPrefix(key, envPrefix))
	if d, err := time.ParseDuration(value); err == nil {
		return mapKey, d
	}
	if b, err := strconv.ParseBool(value); err == nil {
		return mapKey, b
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
