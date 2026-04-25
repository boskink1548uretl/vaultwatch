package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newAgentMockServer(t *testing.T, status int, payload interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestGetAgentInfo_Success(t *testing.T) {
	payload := AgentInfo{
		Initialized:  true,
		SealStatus:   "unsealed",
		Version:      "1.15.0",
		CacheEnabled: true,
	}
	srv := newAgentMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	info, err := c.GetAgentInfo(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Initialized {
		t.Error("expected initialized to be true")
	}
	if !info.CacheEnabled {
		t.Error("expected cache_enabled to be true")
	}
	if info.Version != "1.15.0" {
		t.Errorf("expected version 1.15.0, got %s", info.Version)
	}
}

func TestGetAgentInfo_NotFound(t *testing.T) {
	srv := newAgentMockServer(t, http.StatusNotFound, nil)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = c.GetAgentInfo(context.Background())
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestIsAgentCacheEnabled_True(t *testing.T) {
	payload := AgentInfo{CacheEnabled: true}
	srv := newAgentMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	enabled, err := c.IsAgentCacheEnabled(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !enabled {
		t.Error("expected cache to be enabled")
	}
}

func TestIsAgentCacheEnabled_False(t *testing.T) {
	payload := AgentInfo{CacheEnabled: false}
	srv := newAgentMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	enabled, err := c.IsAgentCacheEnabled(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enabled {
		t.Error("expected cache to be disabled")
	}
}
