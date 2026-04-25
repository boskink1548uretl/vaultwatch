package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newStepDownMockServer(t *testing.T, stepDownStatus int, isSelf bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/step-down":
			w.WriteHeader(stepDownStatus)
		case "/v1/sys/leader":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"is_self": isSelf})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestStepDownLeader_Success(t *testing.T) {
	srv := newStepDownMockServer(t, http.StatusNoContent, true)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	res, err := c.StepDownLeader(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Errorf("expected Success=true, got false (msg: %s)", res.Message)
	}
}

func TestStepDownLeader_Forbidden(t *testing.T) {
	srv := newStepDownMockServer(t, http.StatusForbidden, false)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "bad-token")
	res, err := c.StepDownLeader(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Success {
		t.Error("expected Success=false for forbidden response")
	}
	if res.Message != "permission denied" {
		t.Errorf("unexpected message: %s", res.Message)
	}
}

func TestStepDownLeader_NotActive(t *testing.T) {
	srv := newStepDownMockServer(t, http.StatusBadRequest, false)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	res, err := c.StepDownLeader(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Success {
		t.Error("expected Success=false for non-leader node")
	}
}

func TestIsLeader_True(t *testing.T) {
	srv := newStepDownMockServer(t, http.StatusNoContent, true)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	leader, err := c.IsLeader(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !leader {
		t.Error("expected IsLeader=true")
	}
}

func TestIsLeader_False(t *testing.T) {
	srv := newStepDownMockServer(t, http.StatusNoContent, false)
	defer srv.Close()

	c, _ := NewClient(srv.URL, "test-token")
	leader, err := c.IsLeader(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if leader {
		t.Error("expected IsLeader=false for standby node")
	}
}
