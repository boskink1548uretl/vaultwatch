package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTelemetryMockServer(t *testing.T, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/metrics" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestGetTelemetry_Success(t *testing.T) {
	snap := TelemetrySnapshot{
		Gauges: []TelemetryMetric{
			{Name: "vault.runtime.num_goroutines", Mean: 42},
		},
		Counters: []TelemetryMetric{
			{Name: "vault.token.create", Count: 7},
		},
	}
	srv := newTelemetryMockServer(t, http.StatusOK, snap)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	result, err := c.GetTelemetry(context.Background())
	if err != nil {
		t.Fatalf("GetTelemetry: %v", err)
	}
	if len(result.Gauges) != 1 {
		t.Errorf("expected 1 gauge, got %d", len(result.Gauges))
	}
	if result.Gauges[0].Name != "vault.runtime.num_goroutines" {
		t.Errorf("unexpected gauge name: %s", result.Gauges[0].Name)
	}
	if len(result.Counters) != 1 || result.Counters[0].Count != 7 {
		t.Errorf("unexpected counter data")
	}
}

func TestGetTelemetry_NotFound(t *testing.T) {
	srv := newTelemetryMockServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = c.GetTelemetry(context.Background())
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestGetTelemetry_ServerError(t *testing.T) {
	srv := newTelemetryMockServer(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = c.GetTelemetry(context.Background())
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}
