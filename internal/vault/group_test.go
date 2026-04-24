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

func newGroupMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/identity/group/id":
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"key_info": map[string]any{
						"abc123": map[string]any{
							"name":     "ops-team",
							"type":     "internal",
							"policies": []string{"default", "ops"},
						},
					},
				},
			})
		case "/v1/identity/group/name/ops-team":
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":       "abc123",
					"name":     "ops-team",
					"type":     "internal",
					"policies": []string{"default", "ops"},
				},
			})
		case "/v1/identity/group/name/missing":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestListGroups_ReturnsList(t *testing.T) {
	srv := newGroupMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	require.NoError(t, err)

	groups, err := client.ListGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, "ops-team", groups[0].Name)
	assert.Equal(t, "internal", groups[0].Type)
	assert.Contains(t, groups[0].Policies, "ops")
}

func TestGetGroup_Found(t *testing.T) {
	srv := newGroupMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	require.NoError(t, err)

	group, err := client.GetGroup(context.Background(), "ops-team")
	require.NoError(t, err)
	require.NotNil(t, group)
	assert.Equal(t, "ops-team", group.Name)
	assert.Equal(t, "abc123", group.ID)
}

func TestGetGroup_NotFound(t *testing.T) {
	srv := newGroupMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	require.NoError(t, err)

	group, err := client.GetGroup(context.Background(), "missing")
	require.NoError(t, err)
	assert.Nil(t, group)
}
