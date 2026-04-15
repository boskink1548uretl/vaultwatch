package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newAuditMockServer(t *testing.T, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestListAuditDevices_ReturnsList(t *testing.T) {
	payload := map[string]interface{}{
		"file/": map[string]interface{}{
			"type":        "file",
			"description": "file audit log",
			"path":        "file/",
			"options":     map[string]string{"file_path": "/var/log/vault_audit.log"},
			"local":       false,
		},
	}
	srv := newAuditMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	devices, err := c.ListAuditDevices(context.Background())
	if err != nil {
		t.Fatalf("ListAuditDevices: %v", err)
	}
	if len(devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(devices))
	}
	if d, ok := devices["file/"]; !ok || d.Type != "file" {
		t.Errorf("expected file device, got %+v", devices)
	}
}

func TestListAuditDevices_Empty(t *testing.T) {
	srv := newAuditMockServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	devices, err := c.ListAuditDevices(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("expected empty map, got %d entries", len(devices))
	}
}

func TestIsAuditEnabled_True(t *testing.T) {
	payload := map[string]interface{}{
		"syslog/": map[string]interface{}{"type": "syslog", "path": "syslog/"},
	}
	srv := newAuditMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	enabled, err := c.IsAuditEnabled(context.Background())
	if err != nil {
		t.Fatalf("IsAuditEnabled: %v", err)
	}
	if !enabled {
		t.Error("expected audit to be enabled")
	}
}

func TestIsAuditEnabled_False(t *testing.T) {
	srv := newAuditMockServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	enabled, err := c.IsAuditEnabled(context.Background())
	if err != nil {
		t.Fatalf("IsAuditEnabled: %v", err)
	}
	if enabled {
		t.Error("expected audit to be disabled")
	}
}
