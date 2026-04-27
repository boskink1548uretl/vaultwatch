package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newHAMockServer(t *testing.T, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/ha-status" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestGetHAStatus_Leader(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"cluster_id":     "abc-123",
			"cluster_name":   "vault-cluster",
			"leader_address": "https://vault-0:8200",
			"is_self":        true,
			"nodes": []map[string]interface{}{
				{"hostname": "vault-0", "api_address": "https://vault-0:8200", "active_node": true},
			},
		},
	}
	srv := newHAMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	status, err := client.GetHAStatus(context.Background())
	if err != nil {
		t.Fatalf("GetHAStatus: %v", err)
	}
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if !status.IsLeader {
		t.Errorf("expected IsLeader=true")
	}
	if status.ClusterName != "vault-cluster" {
		t.Errorf("expected cluster name 'vault-cluster', got %q", status.ClusterName)
	}
	if len(status.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(status.Nodes))
	}
}

func TestGetHAStatus_NotFound(t *testing.T) {
	srv := newHAMockServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	status, err := client.GetHAStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != nil {
		t.Errorf("expected nil status for 404")
	}
}

func TestIsHALeader_False(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"is_self": false,
		},
	}
	srv := newHAMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	leader, err := client.IsHALeader(context.Background())
	if err != nil {
		t.Fatalf("IsHALeader: %v", err)
	}
	if leader {
		t.Errorf("expected false for non-leader node")
	}
}
