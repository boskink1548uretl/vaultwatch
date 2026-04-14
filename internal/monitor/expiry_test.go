package monitor_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
)

var (
	warning  = 72 * time.Hour
	critical = 24 * time.Hour
)

func newChecker() *monitor.Checker {
	return monitor.NewChecker(warning, critical)
}

func TestEvaluate_NoAlert(t *testing.T) {
	c := newChecker()
	meta := &vault.SecretMetadata{ExpiresAt: time.Now().Add(96 * time.Hour)}
	alert := c.Evaluate("secret/data/myapp", meta)
	if alert != nil {
		t.Errorf("expected no alert, got %+v", alert)
	}
}

func TestEvaluate_WarningAlert(t *testing.T) {
	c := newChecker()
	meta := &vault.SecretMetadata{ExpiresAt: time.Now().Add(48 * time.Hour)}
	alert := c.Evaluate("secret/data/myapp", meta)
	if alert == nil {
		t.Fatal("expected a warning alert, got nil")
	}
	if alert.Level != monitor.AlertLevelWarning {
		t.Errorf("expected level %q, got %q", monitor.AlertLevelWarning, alert.Level)
	}
}

func TestEvaluate_CriticalAlert(t *testing.T) {
	c := newChecker()
	meta := &vault.SecretMetadata{ExpiresAt: time.Now().Add(12 * time.Hour)}
	alert := c.Evaluate("secret/data/myapp", meta)
	if alert == nil {
		t.Fatal("expected a critical alert, got nil")
	}
	if alert.Level != monitor.AlertLevelCritical {
		t.Errorf("expected level %q, got %q", monitor.AlertLevelCritical, alert.Level)
	}
}

func TestEvaluate_Expired(t *testing.T) {
	c := newChecker()
	meta := &vault.SecretMetadata{ExpiresAt: time.Now().Add(-1 * time.Hour)}
	alert := c.Evaluate("secret/data/myapp", meta)
	if alert == nil {
		t.Fatal("expected an expired alert, got nil")
	}
	if alert.Level != monitor.AlertLevelExpired {
		t.Errorf("expected level %q, got %q", monitor.AlertLevelExpired, alert.Level)
	}
}

func TestEvaluate_NilMetadata(t *testing.T) {
	c := newChecker()
	alert := c.Evaluate("secret/data/myapp", nil)
	if alert != nil {
		t.Errorf("expected no alert for nil metadata, got %+v", alert)
	}
}

func TestEvaluate_ZeroExpiry(t *testing.T) {
	c := newChecker()
	meta := &vault.SecretMetadata{ExpiresAt: time.Time{}}
	alert := c.Evaluate("secret/data/myapp", meta)
	if alert != nil {
		t.Errorf("expected no alert for zero expiry, got %+v", alert)
	}
}
