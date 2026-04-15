package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newMountMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/sys/mounts" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"secret/": map[string]interface{}{
						"type":        "kv",
						"description": "KV secrets",
						"options":     map[string]interface{}{"version": "2"},
					},
					"sys/": map[string]interface{}{
						"type":        "system",
						"description": "System backend",
					},
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func newMountClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("creating vault client: %v", err)
	}
	return &Client{logical: vc.Logical()}
}

func TestListMounts_ReturnsMounts(t *testing.T) {
	srv := newMountMockServer(t)
	defer srv.Close()
	c := newMountClient(t, srv)

	mounts, err := c.ListMounts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mounts) == 0 {
		t.Fatal("expected at least one mount")
	}
}

func TestGetMount_Found(t *testing.T) {
	srv := newMountMockServer(t)
	defer srv.Close()
	c := newMountClient(t, srv)

	info, err := c.GetMount(context.Background(), "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Type != "kv" {
		t.Errorf("expected type kv, got %q", info.Type)
	}
	if info.Options["version"] != "2" {
		t.Errorf("expected option version=2, got %q", info.Options["version"])
	}
}

func TestGetMount_NotFound(t *testing.T) {
	srv := newMountMockServer(t)
	defer srv.Close()
	c := newMountClient(t, srv)

	_, err := c.GetMount(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing mount, got nil")
	}
}
