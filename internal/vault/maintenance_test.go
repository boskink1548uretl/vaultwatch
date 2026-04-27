package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMaintenanceMockServer(t *testing.T, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/maintenance" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestGetMaintenanceStatus_Enabled(t *testing.T) {
	payload := map[string]interface{}{"enabled": true, "message": "scheduled maintenance", "response_code": 503}
	srv := newMaintenanceMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	status, err := c.GetMaintenanceStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Enabled {
		t.Error("expected maintenance mode to be enabled")
	}
	if status.Message != "scheduled maintenance" {
		t.Errorf("unexpected message: %q", status.Message)
	}
}

func TestGetMaintenanceStatus_NotFound(t *testing.T) {
	srv := newMaintenanceMockServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	status, err := c.GetMaintenanceStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Enabled {
		t.Error("expected maintenance mode to be disabled when endpoint returns 404")
	}
}

func TestIsMaintenanceModeEnabled_True(t *testing.T) {
	payload := map[string]interface{}{"enabled": true}
	srv := newMaintenanceMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	enabled, err := c.IsMaintenanceModeEnabled(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !enabled {
		t.Error("expected true")
	}
}

func TestIsMaintenanceModeEnabled_False(t *testing.T) {
	payload := map[string]interface{}{"enabled": false}
	srv := newMaintenanceMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	enabled, err := c.IsMaintenanceModeEnabled(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enabled {
		t.Error("expected false")
	}
}
