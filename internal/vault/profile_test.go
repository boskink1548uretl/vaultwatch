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

func newProfileMockServer(t *testing.T, status int, payload any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(payload)
	}))
}

func TestGetProfile_Success(t *testing.T) {
	payload := map[string]any{
		"version":      "1.15.0",
		"cluster_name": "vault-cluster-prod",
		"cluster_id":   "abc-123",
		"ha_enabled":   true,
		"sealed":       false,
		"initialized":  true,
	}
	srv := newProfileMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	require.NoError(t, err)

	profile, err := client.GetProfile(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "1.15.0", profile.Version)
	assert.Equal(t, "vault-cluster-prod", profile.ClusterName)
	assert.Equal(t, "abc-123", profile.ClusterID)
	assert.True(t, profile.HA)
	assert.False(t, profile.Sealed)
	assert.True(t, profile.Initialized)
}

func TestGetProfile_Sealed(t *testing.T) {
	payload := map[string]any{
		"version":     "1.14.0",
		"sealed":      true,
		"initialized": true,
		"ha_enabled":  false,
	}
	srv := newProfileMockServer(t, http.StatusServiceUnavailable, payload)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	require.NoError(t, err)

	profile, err := client.GetProfile(context.Background())
	require.NoError(t, err)
	assert.True(t, profile.Sealed)
	assert.Equal(t, "1.14.0", profile.Version)
}

func TestGetProfile_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	require.NoError(t, err)

	_, err = client.GetProfile(context.Background())
	assert.Error(t, err)
}
