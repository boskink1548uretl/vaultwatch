package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newACLMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/auth/token/lookup-self":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"policies": []string{"default", "admin"},
				},
			})
		case "/v1/sys/policy/admin":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name":  "admin",
				"rules": `path "secret/*" { capabilities = ["read"] }`,
			})
		case "/v1/sys/policy/missing":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestGetACLTokenPolicies_Success(t *testing.T) {
	srv := newACLMockServer(t)
	defer srv.Close()
	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	policies, err := c.GetACLTokenPolicies(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(policies) != 2 {
		t.Errorf("expected 2 policies, got %d", len(policies))
	}
	if policies[0] != "default" || policies[1] != "admin" {
		t.Errorf("unexpected policies: %v", policies)
	}
}

func TestGetACLPolicy_Found(t *testing.T) {
	srv := newACLMockServer(t)
	defer srv.Close()
	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	pol, err := c.GetACLPolicy(t.Context(), "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pol.Name != "admin" {
		t.Errorf("expected name admin, got %q", pol.Name)
	}
}

func TestGetACLPolicy_NotFound(t *testing.T) {
	srv := newACLMockServer(t)
	defer srv.Close()
	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.GetACLPolicy(t.Context(), "missing")
	if err == nil {
		t.Fatal("expected error for missing policy")
	}
}

func TestGetACLPolicy_EmptyName(t *testing.T) {
	srv := newACLMockServer(t)
	defer srv.Close()
	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = c.GetACLPolicy(t.Context(), "")
	if err == nil {
		t.Fatal("expected error for empty policy name")
	}
}
