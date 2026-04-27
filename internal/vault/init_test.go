package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newInitMockServer(initialized bool, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/init" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if statusCode == http.StatusOK {
			_ = json.NewEncoder(w).Encode(InitStatus{Initialized: initialized})
		}
	}))
}

func TestGetInitStatus_Initialized(t *testing.T) {
	srv := newInitMockServer(true, http.StatusOK)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	status, err := client.GetInitStatus(context.Background())
	if err != nil {
		t.Fatalf("GetInitStatus: %v", err)
	}
	if !status.Initialized {
		t.Errorf("expected Initialized=true, got false")
	}
}

func TestGetInitStatus_NotInitialized(t *testing.T) {
	srv := newInitMockServer(false, http.StatusOK)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	status, err := client.GetInitStatus(context.Background())
	if err != nil {
		t.Fatalf("GetInitStatus: %v", err)
	}
	if status.Initialized {
		t.Errorf("expected Initialized=false, got true")
	}
}

func TestGetInitStatus_ServerError(t *testing.T) {
	srv := newInitMockServer(false, http.StatusInternalServerError)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.GetInitStatus(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestIsInitialized_True(t *testing.T) {
	srv := newInitMockServer(true, http.StatusOK)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ok, err := client.IsInitialized(context.Background())
	if err != nil {
		t.Fatalf("IsInitialized: %v", err)
	}
	if !ok {
		t.Errorf("expected true, got false")
	}
}

func TestIsInitialized_False(t *testing.T) {
	srv := newInitMockServer(false, http.StatusOK)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ok, err := client.IsInitialized(context.Background())
	if err != nil {
		t.Fatalf("IsInitialized: %v", err)
	}
	if ok {
		t.Errorf("expected false, got true")
	}
}
