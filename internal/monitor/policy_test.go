package monitor

import (
	"context"
	"errors"
	"testing"
)

type mockPolicyRules struct {
	name  string
	rules string
}

func (m *mockPolicyRules) GetName() string  { return m.name }
func (m *mockPolicyRules) GetRules() string { return m.rules }

type mockPolicyFetcher struct {
	policies map[string]string
	listErr  error
	getErr   error
}

func (m *mockPolicyFetcher) ListPolicies(_ context.Context) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	keys := make([]string, 0, len(m.policies))
	for k := range m.policies {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *mockPolicyFetcher) GetPolicy(_ context.Context, name string) (PolicyRules, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	rules, ok := m.policies[name]
	if !ok {
		return nil, errors.New("not found")
	}
	return &mockPolicyRules{name: name, rules: rules}, nil
}

func TestPolicyAudit_NoRiskyCapabilities(t *testing.T) {
	fetcher := &mockPolicyFetcher{
		policies: map[string]string{
			"safe-policy": `path "secret/*" { capabilities = ["read", "list"] }`,
		},
	}
	pc := NewPolicyChecker(fetcher)
	findings, err := pc.Audit(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestPolicyAudit_DetectsSudoCapability(t *testing.T) {
	fetcher := &mockPolicyFetcher{
		policies: map[string]string{
			"admin-policy": `path "sys/*" { capabilities = ["sudo", "read"] }`,
		},
	}
	pc := NewPolicyChecker(fetcher)
	findings, err := pc.Audit(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Capability != "sudo" {
		t.Errorf("expected capability %q, got %q", "sudo", findings[0].Capability)
	}
}

func TestPolicyAudit_SkipsRootPolicy(t *testing.T) {
	fetcher := &mockPolicyFetcher{
		policies: map[string]string{
			"root": `path "*" { capabilities = ["sudo", "delete"] }`,
		},
	}
	pc := NewPolicyChecker(fetcher)
	findings, err := pc.Audit(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("root policy should be skipped, got %d findings", len(findings))
	}
}

func TestPolicyAudit_ListError(t *testing.T) {
	fetcher := &mockPolicyFetcher{listErr: errors.New("vault unavailable")}
	pc := NewPolicyChecker(fetcher)
	_, err := pc.Audit(context.Background())
	if err == nil {
		t.Fatal("expected error from list failure, got nil")
	}
}
