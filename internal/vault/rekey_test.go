package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newRekeyMockServer(t *testing.T, status int, body interface{}) (*httptest.Server, *Client) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
	t.Cleanup(server.Close)
	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return server, client
}

func TestGetRekeyStatus_NotStarted(t *testing.T) {
	_, client := newRekeyMockServer(t, http.StatusNotFound, nil)
	status, err := client.GetRekeyStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Started {
		t.Error("expected Started=false for 404 response")
	}
}

func TestGetRekeyStatus_InProgress(t *testing.T) {
	body := RekeyStatus{Started: true, T: 2, N: 5, Progress: 1, Required: 2}
	_, client := newRekeyMockServer(t, http.StatusOK, body)
	status, err := client.GetRekeyStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Started {
		t.Error("expected Started=true")
	}
	if status.T != 2 || status.N != 5 {
		t.Errorf("unexpected threshold values: T=%d N=%d", status.T, status.N)
	}
}

func TestGetRekeyStatus_ServerError(t *testing.T) {
	_, client := newRekeyMockServer(t, http.StatusInternalServerError, nil)
	_, err := client.GetRekeyStatus(context.Background())
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestIsRekeyInProgress_True(t *testing.T) {
	body := RekeyStatus{Started: true}
	_, client := newRekeyMockServer(t, http.StatusOK, body)
	inProgress, err := client.IsRekeyInProgress(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !inProgress {
		t.Error("expected rekey in progress")
	}
}

func TestIsRekeyInProgress_False(t *testing.T) {
	_, client := newRekeyMockServer(t, http.StatusNotFound, nil)
	inProgress, err := client.IsRekeyInProgress(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inProgress {
		t.Error("expected rekey not in progress")
	}
}
