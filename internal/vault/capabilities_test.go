package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newCapabilitiesMockServer(t *testing.T, response map[string]interface{}, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/capabilities-self" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(status)
		if response != nil {
			_ = json.NewEncoder(w).Encode(response)
		}
	}))
}

func TestGetCapabilities_Success(t *testing.T) {
	response := map[string]interface{}{
		"secret/data/myapp": []interface{}{"read", "list"},
	}
	server := newCapabilitiesMockServer(t, response, http.StatusOK)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	caps, err := client.GetCapabilities(context.Background(), []string{"secret/data/myapp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(caps["secret/data/myapp"]) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(caps["secret/data/myapp"]))
	}
}

func TestGetCapabilities_EmptyPaths(t *testing.T) {
	client, err := NewClient("http://127.0.0.1", "token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = client.GetCapabilities(context.Background(), []string{})
	if err == nil {
		t.Fatal("expected error for empty paths")
	}
}

func TestGetCapabilities_Forbidden(t *testing.T) {
	server := newCapabilitiesMockServer(t, nil, http.StatusForbidden)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	_, err = client.GetCapabilities(context.Background(), []string{"secret/data/myapp"})
	if err == nil {
		t.Fatal("expected error for forbidden response")
	}
}

func TestHasCapability_True(t *testing.T) {
	response := map[string]interface{}{
		"secret/data/myapp": []interface{}{"read", "update"},
	}
	server := newCapabilitiesMockServer(t, response, http.StatusOK)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ok, err := client.HasCapability(context.Background(), "secret/data/myapp", "read")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected capability 'read' to be present")
	}
}

func TestHasCapability_False(t *testing.T) {
	response := map[string]interface{}{
		"secret/data/myapp": []interface{}{"read"},
	}
	server := newCapabilitiesMockServer(t, response, http.StatusOK)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ok, err := client.HasCapability(context.Background(), "secret/data/myapp", "delete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected capability 'delete' to be absent")
	}
}
