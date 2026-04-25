package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newUnsealMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		// Handle reset
		if reset, ok := body["reset"].(bool); ok && reset {
			json.NewEncoder(w).Encode(UnsealStatus{
				Sealed: true, T: 3, N: 5, Progress: 0, Version: "1.15.0",
			})
			return
		}

		key, _ := body["key"].(string)
		if key == "badkey" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(UnsealStatus{
			Sealed: true, T: 3, N: 5, Progress: 1, Version: "1.15.0",
		})
	}))
}

func TestSubmitUnsealKey_Success(t *testing.T) {
	srv := newUnsealMockServer(t)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	status, err := client.SubmitUnsealKey(context.Background(), "validkey")
	if err != nil {
		t.Fatalf("SubmitUnsealKey: %v", err)
	}
	if status.Progress != 1 {
		t.Errorf("expected progress 1, got %d", status.Progress)
	}
	if !status.Sealed {
		t.Error("expected sealed=true")
	}
}

func TestSubmitUnsealKey_EmptyKey(t *testing.T) {
	client, _ := NewClient("http://127.0.0.1", "tok")
	_, err := client.SubmitUnsealKey(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestSubmitUnsealKey_BadResponse(t *testing.T) {
	srv := newUnsealMockServer(t)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "tok")
	_, err := client.SubmitUnsealKey(context.Background(), "badkey")
	if err == nil {
		t.Fatal("expected error for bad key")
	}
}

func TestResetUnsealProgress_Success(t *testing.T) {
	srv := newUnsealMockServer(t)
	defer srv.Close()

	client, _ := NewClient(srv.URL, "tok")
	status, err := client.ResetUnsealProgress(context.Background())
	if err != nil {
		t.Fatalf("ResetUnsealProgress: %v", err)
	}
	if status.Progress != 0 {
		t.Errorf("expected progress 0 after reset, got %d", status.Progress)
	}
}
