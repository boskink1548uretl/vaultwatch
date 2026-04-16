package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newAuthMockServer(t *testing.T, status int, payload any) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(payload)
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return srv, c
}

func TestListAuthMethods_ReturnsList(t *testing.T) {
	payload := map[string]any{
		"token/": map[string]any{"type": "token", "description": "token auth", "local": false},
		"approle/": map[string]any{"type": "approle", "description": "approle auth", "local": false},
	}
	_, c := newAuthMockServer(t, http.StatusOK, payload)

	methods, err := c.ListAuthMethods(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(methods) != 2 {
		t.Errorf("expected 2 methods, got %d", len(methods))
	}
}

func TestGetAuthMethod_Found(t *testing.T) {
	payload := map[string]any{
		"approle/": map[string]any{"type": "approle", "description": "approle"},
	}
	_, c := newAuthMockServer(t, http.StatusOK, payload)

	m, err := c.GetAuthMethod(context.Background(), "approle")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Type != "approle" {
		t.Errorf("expected type approle, got %s", m.Type)
	}
}

func TestGetAuthMethod_NotFound(t *testing.T) {
	payload := map[string]any{}
	_, c := newAuthMockServer(t, http.StatusOK, payload)

	_, err := c.GetAuthMethod(context.Background(), "ldap")
	if err == nil {
		t.Fatal("expected error for missing auth method")
	}
}

func TestListAuthMethods_ServerError(t *testing.T) {
	_, c := newAuthMockServer(t, http.StatusInternalServerError, nil)

	_, err := c.ListAuthMethods(context.Background())
	if err == nil {
		t.Fatal("expected error on server error")
	}
}
