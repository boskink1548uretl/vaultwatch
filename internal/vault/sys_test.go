package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newSysMockServer(t *testing.T, status int, payload any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestGetSysInfo_Success(t *testing.T) {
	payload := map[string]any{
		"version":      "1.15.0",
		"cluster_name": "vault-cluster-prod",
		"cluster_id":   "abc-123",
		"ha_enabled":   true,
		"initialized":  true,
		"sealed":       false,
		"standby":      false,
	}
	srv := newSysMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	info, err := client.GetSysInfo(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "1.15.0" {
		t.Errorf("expected version 1.15.0, got %s", info.Version)
	}
	if info.ClusterName != "vault-cluster-prod" {
		t.Errorf("expected cluster name vault-cluster-prod, got %s", info.ClusterName)
	}
	if !info.HA {
		t.Error("expected HA enabled")
	}
	if info.Sealed {
		t.Error("expected unsealed")
	}
}

func TestGetSysInfo_NotFound(t *testing.T) {
	srv := newSysMockServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.GetSysInfo(context.Background())
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestGetSysInfo_Sealed(t *testing.T) {
	payload := map[string]any{
		"version":     "1.15.0",
		"initialized": true,
		"sealed":      true,
		"standby":     false,
	}
	srv := newSysMockServer(t, http.StatusServiceUnavailable, payload)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	info, err := client.GetSysInfo(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Sealed {
		t.Error("expected sealed to be true")
	}
}
