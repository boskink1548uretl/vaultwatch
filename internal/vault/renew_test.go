package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenewSecret_KVPath(t *testing.T) {
	readCalled := false
	writeCalled := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/app/db":
			// No lease ID — KV v2 path
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"created_time": "2024-01-01T00:00:00Z",
					"custom_metadata": map[string]interface{}{
						"expires_at": "2024-06-01T00:00:00Z",
					},
				},
			})
		case "/v1/secret/data/app/db":
			if r.Method == http.MethodGet {
				readCalled = true
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"data": map[string]interface{}{"username": "admin"},
					},
				})
			} else {
				writeCalled = true
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	require.NoError(t, err)

	result, err := client.RenewSecret(context.Background(), "app/db")
	require.NoError(t, err)
	assert.Equal(t, "app/db", result.Path)
	assert.True(t, readCalled, "expected KV read to be called")
	assert.True(t, writeCalled, "expected KV write to be called")
	assert.Zero(t, result.NewTTL, "KV v2 renewal should have zero TTL")
}

func TestRenewSecret_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"errors": []string{"not found"}})
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	require.NoError(t, err)

	_, err = client.RenewSecret(context.Background(), "missing/secret")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "renew")
}

func TestRenewSecret_KVReadEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/empty/path":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"created_time": "2024-01-01T00:00:00Z"},
			})
		case "/v1/secret/data/empty/path":
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{"errors": []string{"not found"}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	require.NoError(t, err)

	_, err = client.RenewSecret(context.Background(), "empty/path")
	assert.Error(t, err)
}
