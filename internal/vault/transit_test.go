package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTransitMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/transit/keys":
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"keys": []string{"my-key", "other-key"}},
			})
		case "/v1/transit/keys/my-key":
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"name":                   "my-key",
					"type":                   "aes256-gcm96",
					"deletion_allowed":        false,
					"exportable":              true,
					"min_decryption_version":  1,
					"latest_version":          3,
				},
			})
		case "/v1/transit/keys/missing":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestListTransitKeys_ReturnsList(t *testing.T) {
	srv := newTransitMockServer(t)
	defer srv.Close()

	c := &Client{address: srv.URL, token: "test-token", http: srv.Client()}
	keys, err := c.ListTransitKeys(context.Background(), "transit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != "my-key" {
		t.Errorf("expected first key 'my-key', got %q", keys[0])
	}
}

func TestGetTransitKey_Found(t *testing.T) {
	srv := newTransitMockServer(t)
	defer srv.Close()

	c := &Client{address: srv.URL, token: "test-token", http: srv.Client()}
	key, err := c.GetTransitKey(context.Background(), "transit", "my-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Fatal("expected key info, got nil")
	}
	if key.Name != "my-key" {
		t.Errorf("expected name 'my-key', got %q", key.Name)
	}
	if key.Type != "aes256-gcm96" {
		t.Errorf("expected type 'aes256-gcm96', got %q", key.Type)
	}
	if key.LatestVersion != 3 {
		t.Errorf("expected latest version 3, got %d", key.LatestVersion)
	}
	if !key.Exportable {
		t.Error("expected key to be exportable")
	}
}

func TestGetTransitKey_NotFound(t *testing.T) {
	srv := newTransitMockServer(t)
	defer srv.Close()

	c := &Client{address: srv.URL, token: "test-token", http: srv.Client()}
	key, err := c.GetTransitKey(context.Background(), "transit", "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != nil {
		t.Errorf("expected nil for missing key, got %+v", key)
	}
}
