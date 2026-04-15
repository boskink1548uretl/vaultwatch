package audit_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/audit"
	"github.com/yourusername/vaultwatch/internal/monitor"
)

func makeResult(path string, level monitor.AlertLevel, ttl time.Duration) monitor.CheckResult {
	return monitor.CheckResult{
		SecretPath: path,
		Level:      level,
		TTL:        ttl,
	}
}

func newBufferedLogger(t *testing.T) (*audit.Logger, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	l, err := audit.NewLogger("")
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	// Override writer via exported option — use path="" and swap via test helper.
	_ = l
	l2 := audit.NewLoggerWithWriter(&buf)
	return l2, &buf
}

func TestRecord_Warning(t *testing.T) {
	l, buf := newBufferedLogger(t)
	r := makeResult("secret/db/password", monitor.LevelWarning, 72*time.Hour)
	if err := l.Record(r); err != nil {
		t.Fatalf("Record: %v", err)
	}
	var entry audit.Entry
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if entry.SecretPath != "secret/db/password" {
		t.Errorf("path = %q, want secret/db/password", entry.SecretPath)
	}
	if entry.Level != "warning" {
		t.Errorf("level = %q, want warning", entry.Level)
	}
	if entry.DaysLeft < 2.9 || entry.DaysLeft > 3.1 {
		t.Errorf("days_left = %v, want ~3.0", entry.DaysLeft)
	}
}

func TestRecord_Expired(t *testing.T) {
	l, buf := newBufferedLogger(t)
	r := makeResult("secret/api/key", monitor.LevelExpired, -1*time.Hour)
	if err := l.Record(r); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if !strings.Contains(buf.String(), "expired") {
		t.Errorf("expected 'expired' in message, got: %s", buf.String())
	}
}

func TestRecord_NoAlert(t *testing.T) {
	l, buf := newBufferedLogger(t)
	r := makeResult("secret/safe", monitor.LevelNone, 30*24*time.Hour)
	if err := l.Record(r); err != nil {
		t.Fatalf("Record: %v", err)
	}
	var entry audit.Entry
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if entry.Level != "none" {
		t.Errorf("level = %q, want none", entry.Level)
	}
}
