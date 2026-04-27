package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newConnectionMockServer(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
}

func TestCheckConnection_Reachable(t *testing.T) {
	srv := newConnectionMockServer(http.StatusOK)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	info, err := client.CheckConnection(context.Background())
	if err != nil {
		t.Fatalf("CheckConnection: %v", err)
	}
	if !info.Reachable {
		t.Error("expected reachable=true")
	}
	if info.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", info.StatusCode)
	}
	if info.Latency <= 0 {
		t.Error("expected positive latency")
	}
}

func TestCheckConnection_Unreachable(t *testing.T) {
	client, err := NewClient("http://127.0.0.1:19999", "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	info, err := client.CheckConnection(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Reachable {
		t.Error("expected reachable=false for unreachable host")
	}
}

func TestCheckConnection_TLSDetection(t *testing.T) {
	srv := newConnectionMockServer(http.StatusOK)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	info, err := client.CheckConnection(context.Background())
	if err != nil {
		t.Fatalf("CheckConnection: %v", err)
	}
	// httptest.NewServer uses http, not https
	if info.TLSEnabled {
		t.Error("expected TLSEnabled=false for plain http server")
	}
}

func TestIsTLS(t *testing.T) {
	if !isTLS("https://vault.example.com") {
		t.Error("expected true for https")
	}
	if isTLS("http://vault.example.com") {
		t.Error("expected false for http")
	}
}
