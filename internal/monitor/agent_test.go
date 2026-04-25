package monitor

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
)

type mockAgentClient struct {
	cacheEnabled bool
	err          error
}

func (m *mockAgentClient) IsAgentCacheEnabled(_ context.Context) (bool, error) {
	return m.cacheEnabled, m.err
}

func newAgentWatcher(client AgentClient, buf *bytes.Buffer, opts ...AgentWatcherOption) *AgentWatcher {
	logger := log.New(buf, "", 0)
	opts = append(opts, WithAgentLogger(logger))
	return NewAgentWatcher(client, opts...)
}

func TestAgentWatcher_CacheEnabled(t *testing.T) {
	var buf bytes.Buffer
	client := &mockAgentClient{cacheEnabled: true}
	w := newAgentWatcher(client, &buf)

	if err := w.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "enabled") {
		t.Errorf("expected log to mention 'enabled', got: %s", buf.String())
	}
}

func TestAgentWatcher_CacheDisabled_NoRequire(t *testing.T) {
	var buf bytes.Buffer
	client := &mockAgentClient{cacheEnabled: false}
	w := newAgentWatcher(client, &buf)

	if err := w.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "disabled") {
		t.Errorf("expected log to mention 'disabled', got: %s", buf.String())
	}
}

func TestAgentWatcher_CacheDisabled_WithRequire(t *testing.T) {
	var buf bytes.Buffer
	client := &mockAgentClient{cacheEnabled: false}
	w := newAgentWatcher(client, &buf, WithRequireCache(true))

	if err := w.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "WARNING") {
		t.Errorf("expected WARNING in log, got: %s", buf.String())
	}
}

func TestAgentWatcher_ClientError(t *testing.T) {
	var buf bytes.Buffer
	client := &mockAgentClient{err: errors.New("connection refused")}
	w := newAgentWatcher(client, &buf)

	if err := w.Check(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAgentWatcher_WriteReport(t *testing.T) {
	var logBuf, out bytes.Buffer
	client := &mockAgentClient{cacheEnabled: true}
	w := newAgentWatcher(client, &logBuf)

	if err := w.WriteReport(context.Background(), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Cache: enabled") {
		t.Errorf("expected report to contain 'Cache: enabled', got: %s", out.String())
	}
}
