package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newPassthroughMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/myapp/db":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]string{"password": "s3cr3t", "user": "admin"},
			})
		case r.Method == "LIST" && r.URL.Path == "/v1/secret/myapp":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": []string{"db", "api"}},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/secret/missing":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestGetPassthrough_Found(t *testing.T) {
	srv := newPassthroughMockServer(t)
	defer srv.Close()
	c, _ := NewClient(srv.URL, "token")

	entry, err := c.GetPassthrough(context.Background(), "secret/myapp/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Data["password"] != "s3cr3t" {
		t.Errorf("expected password=s3cr3t, got %q", entry.Data["password"])
	}
	if entry.Path != "secret/myapp/db" {
		t.Errorf("expected path set correctly, got %q", entry.Path)
	}
}

func TestGetPassthrough_NotFound(t *testing.T) {
	srv := newPassthroughMockServer(t)
	defer srv.Close()
	c, _ := NewClient(srv.URL, "token")

	_, err := c.GetPassthrough(context.Background(), "secret/missing")
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestListPassthrough_ReturnsList(t *testing.T) {
	srv := newPassthroughMockServer(t)
	defer srv.Close()
	c, _ := NewClient(srv.URL, "token")

	keys, err := c.ListPassthrough(context.Background(), "secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestListPassthrough_NotFound_ReturnsNil(t *testing.T) {
	srv := newPassthroughMockServer(t)
	defer srv.Close()
	c, _ := NewClient(srv.URL, "token")

	keys, err := c.ListPassthrough(context.Background(), "secret/nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if keys != nil {
		t.Errorf("expected nil keys for not-found path")
	}
}
