package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
)

func newSlackResult(level monitor.AlertLevel, ttl time.Duration) monitor.CheckResult {
	return monitor.CheckResult{
		SecretPath:      "secret/data/myapp/db",
		Level:           level,
		TimeUntilExpiry: ttl,
	}
}

func TestSlackNotifier_NoAlert(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()

	n := NewSlackNotifier(ts.URL)
	err := n.Notify(newSlackResult(monitor.LevelNone, 48*time.Hour))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if called {
		t.Error("expected webhook NOT to be called for LevelNone")
	}
}

func TestSlackNotifier_Warning(t *testing.T) {
	var captured slackPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := NewSlackNotifier(ts.URL)
	err := n.Notify(newSlackResult(monitor.LevelWarning, 72*time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(captured.Attachments) == 0 {
		t.Fatal("expected at least one attachment")
	}
	if captured.Attachments[0].Color != "warning" {
		t.Errorf("expected color 'warning', got %q", captured.Attachments[0].Color)
	}
}

func TestSlackNotifier_Critical(t *testing.T) {
	var captured slackPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := NewSlackNotifier(ts.URL)
	err := n.Notify(newSlackResult(monitor.LevelCritical, 12*time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.Attachments[0].Color != "danger" {
		t.Errorf("expected color 'danger', got %q", captured.Attachments[0].Color)
	}
}

func TestSlackNotifier_WebhookError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := NewSlackNotifier(ts.URL)
	err := n.Notify(newSlackResult(monitor.LevelCritical, 6*time.Hour))
	if err == nil {
		t.Error("expected error on non-200 response")
	}
}
