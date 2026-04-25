package monitor

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/yourusername/vaultwatch/internal/vault"
)

type fakeACLClient struct {
	policies []string
	policyFn func(name string) (*vault.ACLTokenPolicy, error)
}

func (f *fakeACLClient) GetACLTokenPolicies(_ context.Context) ([]string, error) {
	return f.policies, nil
}

func (f *fakeACLClient) GetACLPolicy(_ context.Context, name string) (*vault.ACLTokenPolicy, error) {
	if f.policyFn != nil {
		return f.policyFn(name)
	}
	return &vault.ACLTokenPolicy{Name: name, Rules: []vault.ACLRule{}}, nil
}

func newACLWatcher(client ACLClient, required []string) (*ACLWatcher, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	w := NewACLWatcher(client, required, buf)
	return w, buf
}

func TestACLWatcher_AllRequiredPresent(t *testing.T) {
	client := &fakeACLClient{policies: []string{"default", "ops"}}
	w, buf := newACLWatcher(client, []string{"default", "ops"})
	if err := w.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bytes.Contains(buf.Bytes(), []byte("WARNING")) {
		t.Errorf("expected no warnings, got: %s", buf.String())
	}
}

func TestACLWatcher_MissingRequiredPolicy(t *testing.T) {
	client := &fakeACLClient{policies: []string{"default"}}
	w, buf := newACLWatcher(client, []string{"default", "ops"})
	if err := w.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("ops")) {
		t.Errorf("expected warning about missing policy ops, got: %s", buf.String())
	}
}

func TestACLWatcher_RootPolicyWarning(t *testing.T) {
	client := &fakeACLClient{policies: []string{"root"}}
	w, buf := newACLWatcher(client, nil)
	if err := w.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("CRITICAL")) {
		t.Errorf("expected CRITICAL warning for root policy, got: %s", buf.String())
	}
}

func TestACLWatcher_PolicyFetchError(t *testing.T) {
	client := &fakeACLClient{
		policies: []string{"broken"},
		policyFn: func(_ string) (*vault.ACLTokenPolicy, error) {
			return nil, errors.New("fetch failed")
		},
	}
	w, buf := newACLWatcher(client, nil)
	if err := w.Check(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("WARNING")) {
		t.Errorf("expected warning on policy fetch error, got: %s", buf.String())
	}
}
