package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newEntityMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/identity/entity" && r.URL.Query().Get("list") == "true":
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"keys": []string{"alice", "bob"}},
			})
		case r.URL.Path == "/v1/identity/entity/name/alice":
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"id":       "abc-123",
					"name":     "alice",
					"policies": []string{"default"},
					"disabled": false,
				},
			})
		case r.URL.Path == "/v1/identity/entity/name/unknown":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
}

func TestListEntities_ReturnsList(t *testing.T) {
	srv := newEntityMockServer(t)
	defer srv.Close()
	c, _ := NewClient(srv.URL, "token")
	names, err := c.ListEntities(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(names))
	}
}

func TestGetEntity_Found(t *testing.T) {
	srv := newEntityMockServer(t)
	defer srv.Close()
	c, _ := NewClient(srv.URL, "token")
	e, err := c.GetEntity(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e == nil || e.Name != "alice" {
		t.Fatalf("expected entity alice, got %v", e)
	}
	if e.ID != "abc-123" {
		t.Errorf("expected id abc-123, got %s", e.ID)
	}
}

func TestGetEntity_NotFound(t *testing.T) {
	srv := newEntityMockServer(t)
	defer srv.Close()
	c, _ := NewClient(srv.URL, "token")
	e, err := c.GetEntity(context.Background(), "unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e != nil {
		t.Fatalf("expected nil entity, got %v", e)
	}
}
