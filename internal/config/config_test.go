package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "vaultwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	content := `
vault:
  address: "https://vault.example.com"
  token: "s.abc123"
alerts:
  warn_before: "168h"
  critical_before: "24h"
  slack_webhook: "https://hooks.slack.com/xxx"
secrets:
  - path: "secret/data/myapp/db"
    description: "App DB credentials"
`
	path := writeTempConfig(t, content)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Vault.Address != "https://vault.example.com" {
		t.Errorf("expected vault address, got %q", cfg.Vault.Address)
	}
	if cfg.Alerts.WarnBefore.Duration != 168*time.Hour {
		t.Errorf("expected 168h warn_before, got %v", cfg.Alerts.WarnBefore.Duration)
	}
	if len(cfg.Secrets) != 1 {
		t.Errorf("expected 1 secret, got %d", len(cfg.Secrets))
	}
}

func TestLoad_MissingAddress(t *testing.T) {
	content := `
vault:
  token: "s.abc123"
`
	path := writeTempConfig(t, content)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for missing vault.address")
	}
}

func TestLoad_DefaultDurations(t *testing.T) {
	content := `
vault:
  address: "https://vault.example.com"
  token: "s.abc123"
`
	path := writeTempConfig(t, content)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Alerts.WarnBefore.Duration != 7*24*time.Hour {
		t.Errorf("expected default 7-day warn_before, got %v", cfg.Alerts.WarnBefore.Duration)
	}
	if cfg.Alerts.CriticalBefore.Duration != 24*time.Hour {
		t.Errorf("expected default 24h critical_before, got %v", cfg.Alerts.CriticalBefore.Duration)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
