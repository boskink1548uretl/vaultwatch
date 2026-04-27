package monitor

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yourusername/vaultwatch/internal/vault"
)

type fakeHAClient struct {
	status *vault.HAStatus
	err    error
}

func (f *fakeHAClient) GetHAStatus(_ context.Context) (*vault.HAStatus, error) {
	return f.status, f.err
}

func newHAWatcher(client HAClient, buf *bytes.Buffer) *HAWatcher {
	return NewHAWatcher(client, WithHALogger(buf), WithMinNodes(2))
}

func TestHAWatcher_HealthyCluster(t *testing.T) {
	buf := &bytes.Buffer{}
	client := &fakeHAClient{
		status: &vault.HAStatus{
			ClusterName: "prod-cluster",
			LeaderAddr:  "https://vault-0:8200",
			IsLeader:    true,
			Nodes: []vault.HANode{
				{Hostname: "vault-0", ActiveNode: true},
				{Hostname: "vault-1", ActiveNode: false},
			},
		},
	}
	w := newHAWatcher(client, buf)
	if err := w.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "prod-cluster") {
		t.Errorf("expected cluster name in output, got: %s", buf.String())
	}
}

func TestHAWatcher_DegradedCluster(t *testing.T) {
	buf := &bytes.Buffer{}
	client := &fakeHAClient{
		status: &vault.HAStatus{
			ClusterName: "prod-cluster",
			Nodes:       []vault.HANode{{Hostname: "vault-0", ActiveNode: true}},
		},
	}
	w := newHAWatcher(client, buf)
	if err := w.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "warning") {
		t.Errorf("expected degraded warning in output, got: %s", buf.String())
	}
}

func TestHAWatcher_NilStatus(t *testing.T) {
	buf := &bytes.Buffer{}
	client := &fakeHAClient{status: nil}
	w := newHAWatcher(client, buf)
	if err := w.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "unavailable") {
		t.Errorf("expected unavailable message, got: %s", buf.String())
	}
}

func TestHAWatcher_ClientError(t *testing.T) {
	buf := &bytes.Buffer{}
	client := &fakeHAClient{err: errors.New("connection refused")}
	w := newHAWatcher(client, buf)
	if err := w.Check(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestHAWatcher_DefaultMinNodes(t *testing.T) {
	w := NewHAWatcher(&fakeHAClient{})
	if w.minNodes != 1 {
		t.Errorf("expected default minNodes=1, got %d", w.minNodes)
	}
}
