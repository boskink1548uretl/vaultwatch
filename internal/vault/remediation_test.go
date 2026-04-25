package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newRemediationMockServer(t *testing.T, statusCode int, responseBody interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/remediations" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if responseBody != nil {
			_ = json.NewEncoder(w).Encode(responseBody)
		}
	}))
}

func TestApplyRemediation_Success(t *testing.T) {
	srv := newRemediationMockServer(t, http.StatusOK, nil)
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	action := RemediationAction{
		Path:        "secret/myapp",
		Action:      "rotate",
		Description: "rotate expired secret",
	}

	result, err := client.ApplyRemediation(context.Background(), action)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success=true, got false")
	}
	if !result.Action.Applied {
		t.Errorf("expected action.Applied=true")
	}
}

func TestApplyRemediation_EmptyPath(t *testing.T) {
	client, _ := NewClient("http://localhost", "token")
	_, err := client.ApplyRemediation(context.Background(), RemediationAction{Action: "rotate"})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestApplyRemediation_EmptyAction(t *testing.T) {
	client, _ := NewClient("http://localhost", "token")
	_, err := client.ApplyRemediation(context.Background(), RemediationAction{Path: "secret/myapp"})
	if err == nil {
		t.Fatal("expected error for empty action")
	}
}

func TestApplyRemediation_ServerError(t *testing.T) {
	srv := newRemediationMockServer(t, http.StatusInternalServerError, map[string]interface{}{
		"errors": []string{"internal server error"},
	})
	defer srv.Close()

	client, _ := NewClient(srv.URL, "test-token")
	_, err := client.ApplyRemediation(context.Background(), RemediationAction{
		Path:   "secret/myapp",
		Action: "rotate",
	})
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}
