package alert_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
	"github.com/vaultwatch/internal/monitor"
)

func baseResult(level monitor.AlertLevel) monitor.CheckResult {
	return monitor.CheckResult{
		SecretPath: "secret/data/db",
		Level:      level,
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
	}
}

func TestFromCheckResult_NoAlert(t *testing.T) {
	r := baseResult(monitor.LevelNone)
	if got := alert.FromCheckResult(r); got != nil {
		t.Errorf("expected nil notification for LevelNone, got %+v", got)
	}
}

func TestFromCheckResult_Warning(t *testing.T) {
	r := baseResult(monitor.LevelWarning)
	n := alert.FromCheckResult(r)
	if n == nil {
		t.Fatal("expected notification, got nil")
	}
	if n.Severity != alert.SeverityWarning {
		t.Errorf("severity: want WARNING, got %s", n.Severity)
	}
	if !strings.Contains(n.Message, "secret/data/db") {
		t.Errorf("message missing secret path: %s", n.Message)
	}
}

func TestFromCheckResult_Critical(t *testing.T) {
	r := baseResult(monitor.LevelCritical)
	n := alert.FromCheckResult(r)
	if n == nil {
		t.Fatal("expected notification, got nil")
	}
	if n.Severity != alert.SeverityCritical {
		t.Errorf("severity: want CRITICAL, got %s", n.Severity)
	}
}

func TestFromCheckResult_Expired(t *testing.T) {
	r := baseResult(monitor.LevelExpired)
	n := alert.FromCheckResult(r)
	if n == nil {
		t.Fatal("expected notification, got nil")
	}
	if n.Severity != alert.SeverityExpired {
		t.Errorf("severity: want EXPIRED, got %s", n.Severity)
	}
}

func TestStdoutNotifier_Send(t *testing.T) {
	var buf bytes.Buffer
	sn := alert.NewStdoutNotifier(&buf)

	if sn.Name() != "stdout" {
		t.Errorf("name: want stdout, got %s", sn.Name())
	}

	n := alert.Notification{
		SecretPath: "secret/data/api",
		Severity:   alert.SeverityWarning,
		Message:    "[WARNING] Secret 'secret/data/api' expires soon",
	}
	if err := sn.Send(n); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "WARNING") {
		t.Errorf("output missing WARNING: %q", buf.String())
	}
}
