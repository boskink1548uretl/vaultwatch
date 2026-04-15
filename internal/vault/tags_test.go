package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTagsMockServer(t *testing.T, path string, payload interface{}, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/"+path {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(payload)
	}))
}

func TestGetSecretTags_WithTags(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"custom_metadata": map[string]interface{}{
				"owner": "team-infra",
				"env": "production",
			},
		},
	}
	srv := newTagsMockServer(t, "secret/metadata/myapp/db", payload, http.StatusOK)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	tags, err := client.GetSecretTags(context.Background(), "secret/myapp/db")
	if err != nil {
		t.Fatalf("GetSecretTags: %v", err)
	}
	if tags["owner"] != "team-infra" {
		t.Errorf("expected owner=team-infra, got %q", tags["owner"])
	}
	if tags["env"] != "production" {
		t.Errorf("expected env=production, got %q", tags["env"])
	}
}

func TestGetSecretTags_NoCustomMetadata(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{},
	}
	srv := newTagsMockServer(t, "secret/metadata/myapp/token", payload, http.StatusOK)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	tags, err := client.GetSecretTags(context.Background(), "secret/myapp/token")
	if err != nil {
		t.Fatalf("GetSecretTags: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("expected empty tags, got %v", tags)
	}
}

func TestGetSecretTags_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	tags, err := client.GetSecretTags(context.Background(), "secret/missing/path")
	if err != nil {
		t.Fatalf("expected nil error for missing path, got %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("expected empty tags for missing secret, got %v", tags)
	}
}
