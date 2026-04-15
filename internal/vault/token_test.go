package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTokenMockServer(t *testing.T, ttl float64, renewable bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/auth/token/lookup-self" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body := map[string]interface{}{
			"data": map[string]interface{}{
				"accessor":     "abc123",
				"display_name": "token-test",
				"policies":     []interface{}{"default", "read-secrets"},
				"ttl":          ttl,
				"renewable":    renewable,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body)
	}))
}

func TestGetTokenInfo_Success(t *testing.T) {
	srv := newTokenMockServer(t, 3600, true)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	info, err := client.GetTokenInfo(context.Background())
	if err != nil {
		t.Fatalf("GetTokenInfo error: %v", err)
	}

	if info.Accessor != "abc123" {
		t.Errorf("expected accessor abc123, got %s", info.Accessor)
	}
	if info.DisplayName != "token-test" {
		t.Errorf("expected display_name token-test, got %s", info.DisplayName)
	}
	if len(info.Policies) != 2 {
		t.Errorf("expected 2 policies, got %d", len(info.Policies))
	}
	if info.TTL != 3600*time.Second {
		t.Errorf("expected TTL 3600s, got %v", info.TTL)
	}
	if !info.Renewable {
		t.Error("expected renewable to be true")
	}
}

func TestIsTokenExpiringSoon_True(t *testing.T) {
	srv := newTokenMockServer(t, 60, true)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	expiring, err := client.IsTokenExpiringSoon(context.Background(), 5*time.Minute)
	if err != nil {
		t.Fatalf("IsTokenExpiringSoon error: %v", err)
	}
	if !expiring {
		t.Error("expected token to be expiring soon")
	}
}

func TestIsTokenExpiringSoon_False(t *testing.T) {
	srv := newTokenMockServer(t, 86400, true)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	expiring, err := client.IsTokenExpiringSoon(context.Background(), 5*time.Minute)
	if err != nil {
		t.Fatalf("IsTokenExpiringSoon error: %v", err)
	}
	if expiring {
		t.Error("expected token NOT to be expiring soon")
	}
}
