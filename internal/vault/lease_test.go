package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newLeaseMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/renew":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"lease_id":       "database/creds/my-role/abc123",
				"renewable":      true,
				"lease_duration": 3600,
			})
		case "/v1/sys/revoke":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestRenewLease_Success(t *testing.T) {
	srv := newLeaseMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	info, err := client.RenewLease(context.Background(), "database/creds/my-role/abc123", time.Hour)
	if err != nil {
		t.Fatalf("RenewLease: %v", err)
	}

	if info.LeaseID != "database/creds/my-role/abc123" {
		t.Errorf("expected lease ID %q, got %q", "database/creds/my-role/abc123", info.LeaseID)
	}
	if !info.Renewable {
		t.Error("expected lease to be renewable")
	}
	if info.Duration != time.Hour {
		t.Errorf("expected duration 1h, got %v", info.Duration)
	}
	if info.ExpireTime.Before(time.Now()) {
		t.Error("expected ExpireTime to be in the future")
	}
}

func TestRenewLease_EmptyID(t *testing.T) {
	srv := newLeaseMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.RenewLease(context.Background(), "", time.Hour)
	if err == nil {
		t.Fatal("expected error for empty lease ID")
	}
}

func TestRevokeLease_EmptyID(t *testing.T) {
	srv := newLeaseMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	err = client.RevokeLease(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty lease ID")
	}
}

func TestRevokeLease_Success(t *testing.T) {
	srv := newLeaseMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	err = client.RevokeLease(context.Background(), "database/creds/my-role/abc123")
	if err != nil {
		t.Fatalf("RevokeLease: %v", err)
	}
}
