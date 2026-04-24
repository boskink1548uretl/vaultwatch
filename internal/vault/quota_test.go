package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newQuotaMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/sys/quotas/rate-limit" && r.URL.RawQuery == "list=true":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"keys": []string{"global-limit", "kv-limit"},
				},
			})
		case r.URL.Path == "/v1/sys/quotas/rate-limit/global-limit":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"name":           "global-limit",
					"path":           "",
					"rate":           1000.0,
					"interval":       1,
					"block_interval": 0,
				},
			})
		case r.URL.Path == "/v1/sys/quotas/rate-limit/missing":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestListQuotas_ReturnsList(t *testing.T) {
	srv := newQuotaMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	names, err := client.ListQuotas(context.Background())
	if err != nil {
		t.Fatalf("ListQuotas: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 quotas, got %d", len(names))
	}
	if names[0] != "global-limit" {
		t.Errorf("expected global-limit, got %s", names[0])
	}
}

func TestGetQuota_Found(t *testing.T) {
	srv := newQuotaMockServer(t)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "test-token")
	q, err := client.GetQuota(context.Background(), "global-limit")
	if err != nil {
		t.Fatalf("GetQuota: %v", err)
	}
	if q == nil {
		t.Fatal("expected quota, got nil")
	}
	if q.Rate != 1000.0 {
		t.Errorf("expected rate 1000, got %f", q.Rate)
	}
	if q.Name != "global-limit" {
		t.Errorf("expected name global-limit, got %s", q.Name)
	}
}

func TestGetQuota_NotFound(t *testing.T) {
	srv := newQuotaMockServer(t)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "test-token")
	q, err := client.GetQuota(context.Background(), "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q != nil {
		t.Errorf("expected nil quota for missing name, got %+v", q)
	}
}
