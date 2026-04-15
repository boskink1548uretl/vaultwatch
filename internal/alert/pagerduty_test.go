package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

func newPDResult(level monitor.AlertLevel) monitor.CheckResult {
	return monitor.CheckResult{
		SecretPath:    "secret/data/db/password",
		Level:         level,
		TimeRemaining: 48 * time.Hour,
	}
}

func TestPagerDutyNotifier_NoAlert(t *testing.T) {
	notifier := NewPagerDutyNotifier("test-key")
	err := notifier.Notify(newPDResult(monitor.LevelOK))
	if err != nil {
		t.Fatalf("expected no error for LevelOK, got %v", err)
	}
}

func TestPagerDutyNotifier_Warning(t *testing.T) {
	var captured pdPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	notifier := NewPagerDutyNotifier("key-123")
	notifier.eventsURL = ts.URL

	err := notifier.Notify(newPDResult(monitor.LevelWarning))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.Payload.Severity != "warning" {
		t.Errorf("expected severity 'warning', got %q", captured.Payload.Severity)
	}
	if captured.RoutingKey != "key-123" {
		t.Errorf("expected routing key 'key-123', got %q", captured.RoutingKey)
	}
}

func TestPagerDutyNotifier_Critical(t *testing.T) {
	var captured pdPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&captured) //nolint:errcheck
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	notifier := NewPagerDutyNotifier("key-abc")
	notifier.eventsURL = ts.URL

	err := notifier.Notify(newPDResult(monitor.LevelCritical))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.Payload.Severity != "critical" {
		t.Errorf("expected severity 'critical', got %q", captured.Payload.Severity)
	}
}

func TestPagerDutyNotifier_WebhookError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	notifier := NewPagerDutyNotifier("bad-key")
	notifier.eventsURL = ts.URL

	err := notifier.Notify(newPDResult(monitor.LevelCritical))
	if err == nil {
		t.Fatal("expected error for non-202 response, got nil")
	}
}

func TestSeverityForLevel(t *testing.T) {
	tests := []struct {
		level    monitor.AlertLevel
		wantSev  string
	}{
		{monitor.LevelWarning, "warning"},
		{monitor.LevelCritical, "critical"},
		{monitor.LevelExpired, "critical"},
		{monitor.LevelOK, "info"},
	}
	for _, tt := range tests {
		got := severityForLevel(tt.level)
		if got != tt.wantSev {
			t.Errorf("severityForLevel(%v) = %q, want %q", tt.level, got, tt.wantSev)
		}
	}
}
