// Package monitor provides watchers and checkers for Vault state.
package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

// PassthroughClient defines the vault operations needed by PassthroughWatcher.
type PassthroughClient interface {
	ListPassthrough(ctx context.Context, path string) ([]string, error)
	GetPassthrough(ctx context.Context, path string) (interface{ GetData() map[string]string }, error)
}

// PassthroughWatcher monitors generic (passthrough) secret paths for
// expected keys and logs warnings when required keys are absent.
type PassthroughWatcher struct {
	client       passthroughVaultClient
	basePath     string
	requiredKeys []string
	logger       *log.Logger
}

// passthroughVaultClient is the minimal interface used internally.
type passthroughVaultClient interface {
	ListPassthrough(ctx context.Context, path string) ([]string, error)
	GetPassthroughRaw(ctx context.Context, path string) (map[string]string, error)
}

// PassthroughOption configures a PassthroughWatcher.
type PassthroughOption func(*PassthroughWatcher)

// WithPassthroughLogger sets a custom logger.
func WithPassthroughLogger(l *log.Logger) PassthroughOption {
	return func(w *PassthroughWatcher) { w.logger = l }
}

// NewPassthroughWatcher creates a watcher for the given base path and required keys.
func NewPassthroughWatcher(client passthroughVaultClient, basePath string, requiredKeys []string, opts ...PassthroughOption) *PassthroughWatcher {
	w := &PassthroughWatcher{
		client:       client,
		basePath:     basePath,
		requiredKeys: requiredKeys,
		logger:       log.New(os.Stderr, "[passthrough] ", log.LstdFlags),
	}
	for _, o := range opts {
		o(w)
	}
	return w
}

// Check lists secrets under basePath and verifies required keys exist in each entry.
func (w *PassthroughWatcher) Check(ctx context.Context) error {
	keys, err := w.client.ListPassthrough(ctx, w.basePath)
	if err != nil {
		return fmt.Errorf("passthrough watcher: list %q: %w", w.basePath, err)
	}
	if len(keys) == 0 {
		w.logger.Printf("no secrets found under %q", w.basePath)
		return nil
	}

	for _, key := range keys {
		path := w.basePath + "/" + key
		data, err := w.client.GetPassthroughRaw(ctx, path)
		if err != nil {
			w.logger.Printf("WARNING: could not read %q: %v", path, err)
			continue
		}
		for _, rk := range w.requiredKeys {
			if _, ok := data[rk]; !ok {
				w.logger.Printf("WARNING: secret %q missing required key %q", path, rk)
			}
		}
	}
	return nil
}

// WriteReport writes a summary of passthrough paths to w.
func (w *PassthroughWatcher) WriteReport(ctx context.Context, out io.Writer) error {
	keys, err := w.client.ListPassthrough(ctx, w.basePath)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Passthrough secrets under %q: %d entries\n", w.basePath, len(keys))
	for _, k := range keys {
		fmt.Fprintf(out, "  - %s/%s\n", w.basePath, k)
	}
	return nil
}
