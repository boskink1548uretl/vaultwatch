package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newPluginMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/plugins/catalog/secret":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"detailed": []Plugin{
						{Name: "kv", Type: "secret", Version: "v2", Builtin: true},
						{Name: "aws", Type: "secret", Version: "v1", Builtin: false, SHA256: "abc123"},
					},
				},
			})
		case "/v1/sys/plugins/catalog/secret/kv":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": Plugin{Name: "kv", Type: "secret", Version: "v2", Builtin: true},
			})
		case "/v1/sys/plugins/catalog/secret/missing":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestListPlugins_ReturnsList(t *testing.T) {
	srv := newPluginMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	plugins, err := client.ListPlugins(context.Background(), "secret")
	if err != nil {
		t.Fatalf("ListPlugins: %v", err)
	}
	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}
	if plugins[0].Name != "kv" {
		t.Errorf("expected kv, got %s", plugins[0].Name)
	}
}

func TestGetPlugin_Found(t *testing.T) {
	srv := newPluginMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	p, err := client.GetPlugin(context.Background(), "secret", "kv")
	if err != nil {
		t.Fatalf("GetPlugin: %v", err)
	}
	if p == nil {
		t.Fatal("expected plugin, got nil")
	}
	if p.Name != "kv" || !p.Builtin {
		t.Errorf("unexpected plugin data: %+v", p)
	}
}

func TestGetPlugin_NotFound(t *testing.T) {
	srv := newPluginMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	p, err := client.GetPlugin(context.Background(), "secret", "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != nil {
		t.Errorf("expected nil plugin, got %+v", p)
	}
}
