package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newWrappingMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// WrapSecret endpoint — returns a synthetic wrap_info block.
	mux.HandleFunc("/v1/secret/data/myapp/db", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Wrap-TTL") == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"wrap_info": map[string]interface{}{
				"token":           "wrapping-token-abc",
				"accessor":        "acc-xyz",
				"ttl":             300,
				"creation_time":   "2024-01-01T00:00:00Z",
				"creation_path":   "secret/data/myapp/db",
				"wrapped_accessor": "wrapped-acc",
			},
		})
	})

	// LookupWrappingToken endpoint.
	mux.HandleFunc("/v1/sys/wrapping/lookup", func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if payload["token"] == "unknown-token" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"token":         payload["token"],
				"ttl":           300,
				"creation_path": "secret/data/myapp/db",
			},
		})
	})

	return httptest.NewServer(mux)
}

func TestWrapSecret_Success(t *testing.T) {
	srv := newWrappingMockServer(t)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	wrapped, err := c.WrapSecret(context.Background(), "secret/data/myapp/db", "300s")
	if err != nil {
		t.Fatalf("WrapSecret: %v", err)
	}
	if wrapped.Token != "wrapping-token-abc" {
		t.Errorf("expected token 'wrapping-token-abc', got %q", wrapped.Token)
	}
	if wrapped.TTL != 300 {
		t.Errorf("expected TTL 300, got %d", wrapped.TTL)
	}
}

func TestWrapSecret_EmptyPath(t *testing.T) {
	c, _ := NewClient("http://127.0.0.1", "tok")
	_, err := c.WrapSecret(context.Background(), "", "300s")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestWrapSecret_NotFound(t *testing.T) {
	srv := newWrappingMockServer(t)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "tok")
	_, err := c.WrapSecret(context.Background(), "secret/data/does/not/exist", "60s")
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestLookupWrappingToken_Success(t *testing.T) {
	srv := newWrappingMockServer(t)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "tok")
	info, err := c.LookupWrappingToken(context.Background(), "some-valid-token")
	if err != nil {
		t.Fatalf("LookupWrappingToken: %v", err)
	}
	if info.CreationPath != "secret/data/myapp/db" {
		t.Errorf("unexpected creation_path: %q", info.CreationPath)
	}
}

func TestLookupWrappingToken_NotFound(t *testing.T) {
	srv := newWrappingMockServer(t)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "tok")
	_, err := c.LookupWrappingToken(context.Background(), "unknown-token")
	if err == nil {
		t.Fatal("expected error for unknown wrapping token")
	}
}

func TestLookupWrappingToken_EmptyToken(t *testing.T) {
	c, _ := NewClient("http://127.0.0.1", "tok")
	_, err := c.LookupWrappingToken(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}
