package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newRaftMockServer(t *testing.T, status int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestGetRaftConfiguration_ReturnsPeers(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"index": 42,
			"servers": []map[string]interface{}{
				{"node_id": "node1", "address": "127.0.0.1:8201", "leader": true, "voter": true},
				{"node_id": "node2", "address": "127.0.0.2:8201", "leader": false, "voter": true},
			},
		},
	}
	srv := newRaftMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	cfg, err := c.GetRaftConfiguration(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 2 {
		t.Errorf("expected 2 peers, got %d", len(cfg.Servers))
	}
	if !cfg.Servers[0].Leader {
		t.Error("expected first server to be leader")
	}
}

func TestGetRaftConfiguration_NotFound(t *testing.T) {
	srv := newRaftMockServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = c.GetRaftConfiguration(context.Background())
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestIsRaftLeader_True(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"index": 1,
			"servers": []map[string]interface{}{
				{"node_id": "node1", "address": "127.0.0.1:8201", "leader": true, "voter": true},
			},
		},
	}
	srv := newRaftMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "tok")
	leader, err := c.IsRaftLeader(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !leader {
		t.Error("expected leader=true")
	}
}

func TestIsRaftLeader_ServerError(t *testing.T) {
	srv := newRaftMockServer(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "tok")
	_, err := c.IsRaftLeader(context.Background())
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}
