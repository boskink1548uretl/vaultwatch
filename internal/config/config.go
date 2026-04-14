package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level VaultWatch configuration.
type Config struct {
	Vault   VaultConfig   `yaml:"vault"`
	Alerts  AlertsConfig  `yaml:"alerts"`
	Secrets []SecretEntry `yaml:"secrets"`
}

// VaultConfig contains connection settings for HashiCorp Vault.
type VaultConfig struct {
	Address   string `yaml:"address"`
	Token     string `yaml:"token"`
	Namespace string `yaml:"namespace"`
}

// AlertsConfig defines when and how to send alerts.
type AlertsConfig struct {
	// WarnBefore is the duration before expiry to start warning.
	WarnBefore    Duration `yaml:"warn_before"`
	CriticalBefore Duration `yaml:"critical_before"`
	SlackWebhook  string   `yaml:"slack_webhook"`
	Email         string   `yaml:"email"`
}

// SecretEntry describes a single secret path to monitor.
type SecretEntry struct {
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
}

// Duration is a wrapper around time.Duration that supports YAML unmarshalling.
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = parsed
	return nil
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Vault.Address == "" {
		return fmt.Errorf("vault.address is required")
	}
	if c.Vault.Token == "" {
		return fmt.Errorf("vault.token is required")
	}
	if c.Alerts.WarnBefore.Duration == 0 {
		c.Alerts.WarnBefore.Duration = 7 * 24 * time.Hour // default 7 days
	}
	if c.Alerts.CriticalBefore.Duration == 0 {
		c.Alerts.CriticalBefore.Duration = 24 * time.Hour // default 1 day
	}
	return nil
}
