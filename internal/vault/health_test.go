package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newHealthMockServer(initialized, sealed bool, version string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/sys/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"initialized": initialized,
				"sealed":      sealed,
				"version":     version,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestCheckHealth_Healthy(t *testing.T) {
	srv := newHealthMockServer(true, false, "1.15.0", http.StatusOK)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	status, err := client.CheckHealth(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Healthy {
		t.Error("expected healthy=true")
	}
	if status.Version != "1.15.0" {
		t.Errorf("expected version 1.15.0, got %s", status.Version)
	}
	if status.CheckedAt.IsZero() {
		t.Error("expected non-zero CheckedAt")
	}
}

func TestCheckHealth_Sealed(t *testing.T) {
	srv := newHealthMockServer(true, true, "1.15.0", http.StatusServiceUnavailable)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	status, err := client.CheckHealth(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Healthy {
		t.Error("expected healthy=false for sealed vault")
	}
	if !status.Sealed {
		t.Error("expected sealed=true")
	}
	str := status.String()
	if !strings.Contains(str, "sealed") {
		t.Errorf("expected 'sealed' in status string, got: %s", str)
	}
}

func TestCheckHealth_Uninitialized(t *testing.T) {
	srv := newHealthMockServer(false, false, "", http.StatusNotImplemented)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	status, err := client.CheckHealth(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Healthy {
		t.Error("expected healthy=false for uninitialized vault")
	}
	if status.Initialized {
		t.Error("expected initialized=false")
	}
}
