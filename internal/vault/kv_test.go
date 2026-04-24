package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newKVMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/mounts/secret":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"options": map[string]string{"version": "2"},
			})
		case "/v1/sys/mounts/kv1":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"options": map[string]string{"version": "1"},
			})
		case "/v1/sys/mounts/missing":
			w.WriteHeader(http.StatusNotFound)
		case "/v1/secret/metadata/myapp/db":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"current_version":      3,
					"oldest_version":       1,
					"created_time":         "2024-01-15T10:00:00.000000000Z",
					"updated_time":         "2024-06-01T08:30:00.000000000Z",
					"max_versions":         10,
					"delete_version_after": "0s",
				},
			})
		case "/v1/secret/metadata/notfound":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestGetKVVersion_V2(t *testing.T) {
	srv := newKVMockServer(t)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	v, err := c.GetKVVersion(context.Background(), "secret")
	if err != nil {
		t.Fatalf("GetKVVersion: %v", err)
	}
	if v != KVv2 {
		t.Errorf("expected KVv2, got %d", v)
	}
}

func TestGetKVVersion_V1(t *testing.T) {
	srv := newKVMockServer(t)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	v, err := c.GetKVVersion(context.Background(), "kv1")
	if err != nil {
		t.Fatalf("GetKVVersion: %v", err)
	}
	if v != KVv1 {
		t.Errorf("expected KVv1, got %d", v)
	}
}

func TestGetKVVersion_MissingMount(t *testing.T) {
	srv := newKVMockServer(t)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	_, err := c.GetKVVersion(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for missing mount")
	}
}

func TestGetKVMetadata_Found(t *testing.T) {
	srv := newKVMockServer(t)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	meta, err := c.GetKVMetadata(context.Background(), "secret", "myapp/db")
	if err != nil {
		t.Fatalf("GetKVMetadata: %v", err)
	}
	if meta.CurrentVersion != 3 {
		t.Errorf("expected CurrentVersion=3, got %d", meta.CurrentVersion)
	}
	if meta.MaxVersions != 10 {
		t.Errorf("expected MaxVersions=10, got %d", meta.MaxVersions)
	}
	if meta.Path != "secret/myapp/db" {
		t.Errorf("unexpected path: %s", meta.Path)
	}
	if meta.CreatedTime.IsZero() {
		t.Error("expected non-zero CreatedTime")
	}
}

func TestGetKVMetadata_NotFound(t *testing.T) {
	srv := newKVMockServer(t)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	_, err := c.GetKVMetadata(context.Background(), "secret", "notfound")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}
