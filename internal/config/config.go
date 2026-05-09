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
	NoColor         bool          `koanf:"no_color"`
}

func Load(cfgPath string) (*Config, error) {
	k := koanf.New(".")

	if cfgPath == "" {
		cfgPath = defaultConfigPath()
	}
	if cfgPath != "" {
		if err := k.Load(file.Provider(cfgPath), yaml.Parser()); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("config file %s: %w", cfgPath, err)
			}
		}
	}

	if err := k.Load(env.ProviderWithValue(envPrefix, ".", envMapper), nil); err != nil {
		return nil, fmt.Errorf("env vars: %w", err)
	}

	c := &Config{
		DefaultDuration: durationValue(k, "default_duration", 2*time.Hour),
		TenantID:        k.String("tenant_id"),
		SubscriptionID:  k.String("subscription_id"),
		NoColor:         boolValue(k, "no_color"),
	}
	return c, nil
}

func durationValue(k *koanf.Koanf, key string, fallback time.Duration) time.Duration {
	v := k.Get(key)
	switch val := v.(type) {
	case time.Duration:
		return val
	case string:
		d, err := time.ParseDuration(val)
		if err == nil {
			return d
		}
	}
	return fallback
}

func boolValue(k *koanf.Koanf, key string) bool {
	if v := k.Get(key); v != nil {
		switch val := v.(type) {
		case bool:
			return val
		case string:
			b, err := strconv.ParseBool(val)
			if err == nil {
				return b
			}
		}
	}
	return false
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
