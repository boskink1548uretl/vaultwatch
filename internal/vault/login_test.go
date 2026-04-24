package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newLoginMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/auth/approle/login":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["role_id"] == "" || body["secret_id"] == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"auth": map[string]interface{}{
					"client_token":   "s.testtoken",
					"accessor":       "acc123",
					"policies":       []string{"default"},
					"lease_duration": 3600,
					"renewable":      true,
				},
			})
		case "/v1/auth/token/lookup-self":
			token := r.Header.Get("X-Vault-Token")
			if token == "bad-token" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"accessor":  "acc456",
					"policies":  []string{"default", "admin"},
					"ttl":       1800,
					"renewable": false,
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestLoginWithAppRole_Success(t *testing.T) {
	srv := newLoginMockServer(t)
	defer srv.Close()
	c, err := NewClient(srv.URL, "ignored")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	res, err := c.LoginWithAppRole(context.Background(), "my-role-id", "my-secret-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ClientToken != "s.testtoken" {
		t.Errorf("expected client_token s.testtoken, got %q", res.ClientToken)
	}
	if !res.Renewable {
		t.Error("expected renewable=true")
	}
}

func TestLoginWithAppRole_EmptyRoleID(t *testing.T) {
	srv := newLoginMockServer(t)
	defer srv.Close()
	c, _ := NewClient(srv.URL, "tok")
	_, err := c.LoginWithAppRole(context.Background(), "", "secret")
	if err == nil {
		t.Fatal("expected error for empty roleID")
	}
}

func TestLoginWithToken_Success(t *testing.T) {
	srv := newLoginMockServer(t)
	defer srv.Close()
	c, _ := NewClient(srv.URL, "valid-token")
	res, err := c.LoginWithToken(context.Background(), "valid-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Accessor != "acc456" {
		t.Errorf("expected accessor acc456, got %q", res.Accessor)
	}
	if len(res.Policies) != 2 {
		t.Errorf("expected 2 policies, got %d", len(res.Policies))
	}
}

func TestLoginWithToken_InvalidToken(t *testing.T) {
	srv := newLoginMockServer(t)
	defer srv.Close()
	c, _ := NewClient(srv.URL, "bad-token")
	_, err := c.LoginWithToken(context.Background(), "bad-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestLoginWithToken_EmptyToken(t *testing.T) {
	srv := newLoginMockServer(t)
	defer srv.Close()
	c, _ := NewClient(srv.URL, "")
	_, err := c.LoginWithToken(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}
