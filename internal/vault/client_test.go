package vault_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/vault"
)

func newMockVaultServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/v1/sys/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sealed":      false,
			"initialized": true,
		})
	})

	// Secret endpoint
	mux.HandleFunc("/v1/secret/myapp/db", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"lease_id":       "secret/myapp/db/abc123",
			"lease_duration": 3600,
			"renewable":      true,
			"data": map[string]interface{}{
				"password": "s3cr3t",
			},
		})
	})

	return httptest.NewServer(mux)
}

func TestNewClient(t *testing.T) {
	srv := newMockVaultServer(t)
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client.Address != srv.URL {
		t.Errorf("expected address %q, got %q", srv.URL, client.Address)
	}
}

func TestIsHealthy(t *testing.T) {
	srv := newMockVaultServer(t)
	defer srv.Close()

	client, _ := vault.NewClient(srv.URL, "test-token")
	if err := client.IsHealthy(); err != nil {
		t.Errorf("expected healthy vault, got error: %v", err)
	}
}

func TestGetSecretMetadata(t *testing.T) {
	srv := newMockVaultServer(t)
	defer srv.Close()

	client, _ := vault.NewClient(srv.URL, "test-token")
	meta, err := client.GetSecretMetadata("secret/myapp/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Path != "secret/myapp/db" {
		t.Errorf("expected path %q, got %q", "secret/myapp/db", meta.Path)
	}
	if meta.TTL.Seconds() != 3600 {
		t.Errorf("expected TTL 3600s, got %v", meta.TTL)
	}
	if !meta.Renewable {
		t.Error("expected secret to be renewable")
	}
}

func TestGetSecretMetadata_NotFound(t *testing.T) {
	srv := newMockVaultServer(t)
	defer srv.Close()

	client, _ := vault.NewClient(srv.URL, "test-token")
	_, err := client.GetSecretMetadata("secret/does/not/exist")
	if err == nil {
		t.Error("expected error for missing secret, got nil")
	}
}
