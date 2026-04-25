package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// TransitClient defines the vault operations required by TransitWatcher.
type TransitClient interface {
	ListTransitKeys(ctx context.Context, mount string) ([]string, error)
	GetTransitKey(ctx context.Context, mount, keyName string) (*vault.TransitKeyInfo, error)
}

// TransitWatcherOption configures a TransitWatcher.
type TransitWatcherOption func(*TransitWatcher)

// TransitWatcher monitors transit encryption keys for risky configurations.
type TransitWatcher struct {
	client TransitClient
	mount  string
	logger *log.Logger
}

// WithTransitLogger sets a custom logger on the TransitWatcher.
func WithTransitLogger(l *log.Logger) TransitWatcherOption {
	return func(w *TransitWatcher) {
		w.logger = l
	}
}

// NewTransitWatcher creates a TransitWatcher for the given mount path.
func NewTransitWatcher(client TransitClient, mount string, opts ...TransitWatcherOption) *TransitWatcher {
	w := &TransitWatcher{
		client: client,
		mount:  mount,
		logger: log.New(os.Stderr, "[transit] ", log.LstdFlags),
	}
	for _, o := range opts {
		o(w)
	}
	return w
}

// Check lists all transit keys and logs warnings for risky configurations
// such as exportable keys or deletion-allowed keys.
func (w *TransitWatcher) Check(ctx context.Context, out io.Writer) error {
	keys, err := w.client.ListTransitKeys(ctx, w.mount)
	if err != nil {
		return fmt.Errorf("transit watcher: list keys: %w", err)
	}
	if len(keys) == 0 {
		w.logger.Printf("no transit keys found under mount %q", w.mount)
		return nil
	}

	for _, name := range keys {
		info, err := w.client.GetTransitKey(ctx, w.mount, name)
		if err != nil {
			w.logger.Printf("WARN could not fetch key %q: %v", name, err)
			continue
		}
		if info == nil {
			continue
		}
		if info.DeletionAllowed {
			msg := fmt.Sprintf("WARN transit key %q has deletion_allowed=true (mount: %s)\n", name, w.mount)
			w.logger.Print(msg)
			fmt.Fprint(out, msg)
		}
		if info.Exportable {
			msg := fmt.Sprintf("WARN transit key %q is exportable (mount: %s)\n", name, w.mount)
			w.logger.Print(msg)
			fmt.Fprint(out, msg)
		}
	}
	return nil
}
