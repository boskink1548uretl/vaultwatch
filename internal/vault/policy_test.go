package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newPolicyMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/sys/policies/acl/my-policy" && r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]string{
					"name":   "my-policy",
					"policy": `path "secret/*" { capabilities = ["read"] }`,
				},
			})
		case r.URL.Path == "/v1/sys/policies/acl/missing-policy":
			w.WriteHeader(http.StatusNotFound)
		case r.URL.Path == "/v1/sys/policies/acl" && r.URL.Query().Get("list") == "true":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"keys": []string{"default", "my-policy", "root"},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestGetPolicy_Found(t *testing.T) {
	srv := newPolicyMockServer(t)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	policy, err := c.GetPolicy(context.Background(), "my-policy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Name != "my-policy" {
		t.Errorf("expected name %q, got %q", "my-policy", policy.Name)
	}
	if policy.Rules == "" {
		t.Error("expected non-empty rules")
	}
}

func TestGetPolicy_NotFound(t *testing.T) {
	srv := newPolicyMockServer(t)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	_, err := c.GetPolicy(context.Background(), "missing-policy")
	if err == nil {
		t.Fatal("expected error for missing policy, got nil")
	}
}

func TestListPolicies_ReturnsList(t *testing.T) {
	srv := newPolicyMockServer(t)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	policies, err := c.ListPolicies(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(policies) != 3 {
		t.Errorf("expected 3 policies, got %d", len(policies))
	}
	found := false
	for _, p := range policies {
		if p == "my-policy" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected my-policy in list")
	}
}
