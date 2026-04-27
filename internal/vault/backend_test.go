package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newBackendMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/mounts" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		payload := map[string]interface{}{
			"secret/": map[string]interface{}{
				"type":        "kv",
				"description": "key/value store",
				"options":     map[string]string{"version": "2"},
				"local":       false,
				"seal_wrap":   false,
			},
			"pki/": map[string]interface{}{
				"type":        "pki",
				"description": "PKI secrets",
				"options":     map[string]string{},
				"local":       false,
				"seal_wrap":   true,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
}

func TestListBackends_ReturnsList(t *testing.T) {
	srv := newBackendMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	backends, err := client.ListBackends(context.Background())
	if err != nil {
		t.Fatalf("ListBackends: %v", err)
	}
	if len(backends) != 2 {
		t.Errorf("expected 2 backends, got %d", len(backends))
	}
}

func TestGetBackend_Found(t *testing.T) {
	srv := newBackendMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	b, err := client.GetBackend(context.Background(), "secret/")
	if err != nil {
		t.Fatalf("GetBackend: %v", err)
	}
	if b.Type != "kv" {
		t.Errorf("expected type kv, got %s", b.Type)
	}
}

func TestGetBackend_NotFound(t *testing.T) {
	srv := newBackendMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.GetBackend(context.Background(), "nonexistent/")
	if err == nil {
		t.Error("expected error for missing backend, got nil")
	}
}
