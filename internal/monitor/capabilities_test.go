package monitor

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

type mockCapabilityClient struct {
	result map[string][]string
	err    error
}

func (m *mockCapabilityClient) GetCapabilities(_ context.Context, paths []string) (map[string][]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func newCapWatcher(client CapabilityClient, paths, required []string) (*CapabilityWatcher, *bytes.Buffer) {
	var buf bytes.Buffer
	w := NewCapabilityWatcher(client, paths, required, &buf)
	return w, &buf
}

func TestCapabilityWatcher_NoPaths(t *testing.T) {
	client := &mockCapabilityClient{result: map[string][]string{}}
	watcher, _ := newCapWatcher(client, nil, []string{"read"})
	issues, err := watcher.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d", len(issues))
	}
}

func TestCapabilityWatcher_AllPresent(t *testing.T) {
	client := &mockCapabilityClient{
		result: map[string][]string{
			"secret/data/app": {"read", "list"},
		},
	}
	watcher, _ := newCapWatcher(client, []string{"secret/data/app"}, []string{"read", "list"})
	issues, err := watcher.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d", len(issues))
	}
}

func TestCapabilityWatcher_MissingCapability(t *testing.T) {
	client := &mockCapabilityClient{
		result: map[string][]string{
			"secret/data/app": {"read"},
		},
	}
	watcher, buf := newCapWatcher(client, []string{"secret/data/app"}, []string{"read", "delete"})
	issues, err := watcher.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Path != "secret/data/app" {
		t.Errorf("unexpected path: %s", issues[0].Path)
	}
	if len(issues[0].Missing) != 1 || issues[0].Missing[0] != "delete" {
		t.Errorf("expected missing 'delete', got %v", issues[0].Missing)
	}
	if buf.Len() == 0 {
		t.Error("expected log output for missing capability")
	}
}

func TestCapabilityWatcher_ClientError(t *testing.T) {
	client := &mockCapabilityClient{err: errors.New("vault unavailable")}
	watcher, _ := newCapWatcher(client, []string{"secret/data/app"}, []string{"read"})
	_, err := watcher.Check(context.Background())
	if err == nil {
		t.Fatal("expected error from client")
	}
}

func TestCapabilityWatcher_RootBypassesMissing(t *testing.T) {
	client := &mockCapabilityClient{
		result: map[string][]string{
			"secret/data/app": {"root"},
		},
	}
	watcher, _ := newCapWatcher(client, []string{"secret/data/app"}, []string{"read", "delete", "update"})
	issues, err := watcher.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("root token should bypass all capability checks, got %d issues", len(issues))
	}
}
