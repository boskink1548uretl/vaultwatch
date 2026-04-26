package monitor_test

import (
	"bytes"
	"context"
	"log"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
)

// stubTransitClient implements a minimal interface for transit key checks.
type stubTransitClient struct {
	keys []vault.TransitKey
	err  error
}

func (s *stubTransitClient) ListTransitKeys(_ context.Context) ([]vault.TransitKey, error) {
	return s.keys, s.err
}

func newTransitWatcher(keys []vault.TransitKey, err error) (*monitor.TransitWatcher, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	logger := log.New(buf, "", 0)
	client := &stubTransitClient{keys: keys, err: err}
	w := monitor.NewTransitWatcher(client, monitor.WithTransitLogger(logger))
	return w, buf
}

func TestTransitWatcher_NoKeys(t *testing.T) {
	w, buf := newTransitWatcher([]vault.TransitKey{}, nil)
	err := w.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty key list, got: %s", buf.String())
	}
}

func TestTransitWatcher_AllKeysHealthy(t *testing.T) {
	keys := []vault.TransitKey{
		{
			Name:            "my-key",
			Type:            "aes256-gcm96",
			DeletionAllowed: false,
			Exportable:      false,
			LatestVersion:   1,
		},
	}
	w, buf := newTransitWatcher(keys, nil)
	err := w.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no warnings for healthy key, got: %s", buf.String())
	}
}

func TestTransitWatcher_ExportableKeyLogsWarning(t *testing.T) {
	keys := []vault.TransitKey{
		{
			Name:            "risky-key",
			Type:            "aes256-gcm96",
			DeletionAllowed: false,
			Exportable:      true,
			LatestVersion:   2,
		},
	}
	w, buf := newTransitWatcher(keys, nil)
	err := w.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if output == "" {
		t.Error("expected warning for exportable key, got none")
	}
	if !containsSubstring(output, "risky-key") {
		t.Errorf("expected key name in output, got: %s", output)
	}
}

func TestTransitWatcher_DeletionAllowedLogsWarning(t *testing.T) {
	keys := []vault.TransitKey{
		{
			Name:            "deletable-key",
			Type:            "rsa-2048",
			DeletionAllowed: true,
			Exportable:      false,
			LatestVersion:   1,
		},
	}
	w, buf := newTransitWatcher(keys, nil)
	err := w.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if output == "" {
		t.Error("expected warning for deletion-allowed key, got none")
	}
	if !containsSubstring(output, "deletable-key") {
		t.Errorf("expected key name in output, got: %s", output)
	}
}

func TestTransitWatcher_ClientError(t *testing.T) {
	w, _ := newTransitWatcher(nil, context.DeadlineExceeded)
	err := w.Check(context.Background())
	if err == nil {
		t.Fatal("expected error from client, got nil")
	}
}

func TestTransitWatcher_StaleKeyLogsWarning(t *testing.T) {
	// A key with a very old rotation date should trigger a staleness warning.
	keys := []vault.TransitKey{
		{
			Name:            "old-key",
			Type:            "aes256-gcm96",
			DeletionAllowed: false,
			Exportable:      false,
			LatestVersion:   1,
			LastRotatedAt:   time.Now().Add(-365 * 24 * time.Hour),
		},
	}
	w, buf := newTransitWatcher(keys, nil)
	err := w.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if output == "" {
		t.Error("expected staleness warning for old key, got none")
	}
	if !containsSubstring(output, "old-key") {
		t.Errorf("expected key name in staleness warning, got: %s", output)
	}
}

// containsSubstring is a helper shared across monitor tests.
func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && stringContains(s, sub))
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
