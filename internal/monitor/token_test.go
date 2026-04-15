package monitor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

type mockTokenClient struct {
	expiring bool
	err      error
}

func (m *mockTokenClient) IsTokenExpiringSoon(_ context.Context, _ time.Duration) (bool, error) {
	return m.expiring, m.err
}

func TestTokenWatcher_NotExpiring(t *testing.T) {
	client := &mockTokenClient{expiring: false}
	watcher := NewTokenWatcher(client, TokenWatcherConfig{})

	expiring, err := watcher.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expiring {
		t.Error("expected token not to be expiring")
	}
}

func TestTokenWatcher_Expiring_LogsWarning(t *testing.T) {
	client := &mockTokenClient{expiring: true}

	var logged string
	cfg := TokenWatcherConfig{
		WarnThreshold: 15 * time.Minute,
		Logger: func(format string, args ...interface{}) {
			logged = fmt.Sprintf(format, args...)
		},
	}
	watcher := NewTokenWatcher(client, cfg)

	expiring, err := watcher.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !expiring {
		t.Error("expected token to be expiring")
	}
	if !strings.Contains(logged, "WARNING") {
		t.Errorf("expected WARNING in log output, got: %s", logged)
	}
	if !strings.Contains(logged, "15m") {
		t.Errorf("expected threshold in log output, got: %s", logged)
	}
}

func TestTokenWatcher_ClientError(t *testing.T) {
	client := &mockTokenClient{err: errors.New("vault unreachable")}
	watcher := NewTokenWatcher(client, TokenWatcherConfig{})

	_, err := watcher.Check(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "token watcher check failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTokenWatcher_DefaultThreshold(t *testing.T) {
	client := &mockTokenClient{expiring: false}
	watcher := NewTokenWatcher(client, TokenWatcherConfig{})

	if watcher.cfg.WarnThreshold != 30*time.Minute {
		t.Errorf("expected default threshold 30m, got %v", watcher.cfg.WarnThreshold)
	}
}
