package vault

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newSnapshotMockServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sys/storage/raft/snapshot" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func TestTakeSnapshot_Success(t *testing.T) {
	payload := "snapshot-binary-data"
	srv := newSnapshotMockServer(t, http.StatusOK, payload)
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	var buf bytes.Buffer
	info, err := c.TakeSnapshot(context.Background(), &buf)
	if err != nil {
		t.Fatalf("TakeSnapshot: %v", err)
	}
	if buf.String() != payload {
		t.Errorf("body = %q, want %q", buf.String(), payload)
	}
	if info.Size != int64(len(payload)) {
		t.Errorf("size = %d, want %d", info.Size, len(payload))
	}
}

func TestTakeSnapshot_NonOKStatus(t *testing.T) {
	srv := newSnapshotMockServer(t, http.StatusForbidden, "")
	defer srv.Close()

	c, err := NewClient(srv.URL, "bad-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	var buf bytes.Buffer
	_, err = c.TakeSnapshot(context.Background(), &buf)
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestTakeSnapshot_ContextCancelled(t *testing.T) {
	srv := newSnapshotMockServer(t, http.StatusOK, "data")
	defer srv.Close()

	c, err := NewClient(srv.URL, "token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = c.TakeSnapshot(ctx, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
