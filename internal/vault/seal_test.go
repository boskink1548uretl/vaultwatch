package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newSealMockServer(t *testing.T, status int, body interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/seal-status" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

func TestGetSealStatus_Unsealed(t *testing.T) {
	srv := newSealMockServer(t, http.StatusOK, SealStatus{
		Sealed: false, Initialized: true, ClusterName: "vault-cluster", Version: "1.15.0",
	})
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	status, err := c.GetSealStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Sealed {
		t.Error("expected unsealed")
	}
	if status.ClusterName != "vault-cluster" {
		t.Errorf("unexpected cluster name: %s", status.ClusterName)
	}
}

func TestGetSealStatus_Sealed(t *testing.T) {
	srv := newSealMockServer(t, http.StatusOK, SealStatus{Sealed: true, Initialized: true})
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	sealed, err := c.IsSealed(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sealed {
		t.Error("expected sealed to be true")
	}
}

func TestGetSealStatus_ServerError(t *testing.T) {
	srv := newSealMockServer(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	_, err := c.GetSealStatus(context.Background())
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}
