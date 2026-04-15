package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newNamespaceMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/sys/namespaces" && r.URL.RawQuery == "list=true":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"keys": []string{"team-a/", "team-b/"},
				},
			})
		case r.URL.Path == "/v1/sys/namespaces/team-a/":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"path":        "team-a/",
					"id":          "abc123",
					"custom_metadata": map[string]string{"owner": "alice"},
				},
			})
		case r.URL.Path == "/v1/sys/namespaces/missing/":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestListNamespaces_ReturnsList(t *testing.T) {
	srv := newNamespaceMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	keys, err := client.ListNamespaces(context.Background())
	if err != nil {
		t.Fatalf("ListNamespaces: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 namespaces, got %d", len(keys))
	}
	if keys[0] != "team-a/" {
		t.Errorf("expected first key team-a/, got %q", keys[0])
	}
}

func TestGetNamespace_Found(t *testing.T) {
	srv := newNamespaceMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ns, err := client.GetNamespace(context.Background(), "team-a/")
	if err != nil {
		t.Fatalf("GetNamespace: %v", err)
	}
	if ns.ID != "abc123" {
		t.Errorf("expected ID abc123, got %q", ns.ID)
	}
	if ns.CustomMeta["owner"] != "alice" {
		t.Errorf("expected owner=alice, got %q", ns.CustomMeta["owner"])
	}
}

func TestGetNamespace_NotFound(t *testing.T) {
	srv := newNamespaceMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.GetNamespace(context.Background(), "missing/")
	if err == nil {
		t.Fatal("expected error for missing namespace, got nil")
	}
}
