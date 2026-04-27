package monitor

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
)

// fakePassthroughClient satisfies passthroughVaultClient.
type fakePassthroughClient struct {
	keys    []string
	listErr error
	data    map[string]map[string]string // path -> data
	getErr  error
}

func (f *fakePassthroughClient) ListPassthrough(_ context.Context, _ string) ([]string, error) {
	return f.keys, f.listErr
}

func (f *fakePassthroughClient) GetPassthroughRaw(_ context.Context, path string) (map[string]string, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if d, ok := f.data[path]; ok {
		return d, nil
	}
	return map[string]string{}, nil
}

func newPassthroughWatcher(client *fakePassthroughClient, required []string) (*PassthroughWatcher, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	l := log.New(buf, "", 0)
	w := NewPassthroughWatcher(client, "secret/myapp", required, WithPassthroughLogger(l))
	return w, buf
}

func TestPassthroughWatcher_EmptyList(t *testing.T) {
	client := &fakePassthroughClient{keys: []string{}}
	w, buf := newPassthroughWatcher(client, []string{"password"})
	if err := w.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no secrets found") {
		t.Errorf("expected 'no secrets found' log, got: %q", buf.String())
	}
}

func TestPassthroughWatcher_AllRequiredPresent(t *testing.T) {
	client := &fakePassthroughClient{
		keys: []string{"db"},
		data: map[string]map[string]string{
			"secret/myapp/db": {"password": "s3cr3t", "user": "admin"},
		},
	}
	w, buf := newPassthroughWatcher(client, []string{"password", "user"})
	if err := w.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(buf.String(), "WARNING") {
		t.Errorf("unexpected warning: %q", buf.String())
	}
}

func TestPassthroughWatcher_MissingRequiredKey(t *testing.T) {
	client := &fakePassthroughClient{
		keys: []string{"db"},
		data: map[string]map[string]string{
			"secret/myapp/db": {"user": "admin"},
		},
	}
	w, buf := newPassthroughWatcher(client, []string{"password"})
	if err := w.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "missing required key") {
		t.Errorf("expected missing-key warning, got: %q", buf.String())
	}
}

func TestPassthroughWatcher_ListError(t *testing.T) {
	client := &fakePassthroughClient{listErr: errors.New("vault unavailable")}
	w, _ := newPassthroughWatcher(client, []string{"password"})
	if err := w.Check(context.Background()); err == nil {
		t.Fatal("expected error from list failure")
	}
}

func TestPassthroughWatcher_WriteReport(t *testing.T) {
	client := &fakePassthroughClient{keys: []string{"db", "api"}}
	w, _ := newPassthroughWatcher(client, nil)
	var out bytes.Buffer
	if err := w.WriteReport(context.Background(), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "2 entries") {
		t.Errorf("expected entry count in report, got: %q", out.String())
	}
}
