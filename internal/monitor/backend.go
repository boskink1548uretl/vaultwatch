package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/your-org/vaultwatch/internal/vault"
)

// BackendWatcher inspects enabled secrets backends and warns about
// unexpected or potentially risky configurations.
type BackendWatcher struct {
	client    *vault.Client
	logger    *log.Logger
	risky     []string // backend types considered risky if local=false
}

// NewBackendWatcher creates a BackendWatcher with sensible defaults.
func NewBackendWatcher(client *vault.Client, opts ...func(*BackendWatcher)) *BackendWatcher {
	w := &BackendWatcher{
		client: client,
		logger: log.New(os.Stdout, "[backend] ", log.LstdFlags),
		risky:  []string{"aws", "gcp", "azure"},
	}
	for _, o := range opts {
		o(w)
	}
	return w
}

// WithBackendLogger overrides the default logger.
func WithBackendLogger(w io.Writer) func(*BackendWatcher) {
	return func(bw *BackendWatcher) {
		bw.logger = log.New(w, "[backend] ", 0)
	}
}

// Check fetches all backends and logs warnings for risky configurations.
func (bw *BackendWatcher) Check(ctx context.Context) error {
	backends, err := bw.client.ListBackends(ctx)
	if err != nil {
		return fmt.Errorf("backend check: %w", err)
	}

	for _, b := range backends {
		if b.SealWrap {
			bw.logger.Printf("INFO  backend %q has seal_wrap enabled", b.Path)
		}
		for _, rt := range bw.risky {
			if b.Type == rt && !b.Local {
				bw.logger.Printf("WARN  backend %q type=%s is non-local dynamic credentials backend", b.Path, b.Type)
			}
		}
	}
	return nil
}
