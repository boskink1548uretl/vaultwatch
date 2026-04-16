package monitor

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/yourusername/vaultwatch/internal/vault"
)

func newAuthWatcherClient(t *testing.T, payload any) *vault.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(payload)
	}))
	t.Cleanup(srv.Close)
	c, err := vault.NewClient(srv.URL, "tok")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestAuthWatcher_NoRiskyMethods(t *testing.T) {
	payload := map[string]any{
		"token/": map[string]any{"type": "token"},
	}
	c := newAuthWatcherClient(t, payload)
	w := NewAuthWatcher(c, log.New(os.Stderr, "", 0))

	findings, err := w.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestAuthWatcher_DetectsRiskyMethod(t *testing.T) {
	payload := map[string]any{
		"userpass/": map[string]any{"type": "userpass"},
		"token/":    map[string]any{"type": "token"},
	}
	c := newAuthWatcherClient(t, payload)
	w := NewAuthWatcher(c, log.New(os.Stderr, "", 0))

	findings, err := w.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Type != "userpass" {
		t.Errorf("expected userpass finding, got %s", findings[0].Type)
	}
}

func TestAuthWatcher_DefaultThreshold(t *testing.T) {
	c, _ := vault.NewClient("http://localhost", "tok")
	w := NewAuthWatcher(c, log.New(os.Stderr, "", 0))
	if len(w.RiskyTypes) == 0 {
		t.Error("expected default risky types to be set")
	}
}
