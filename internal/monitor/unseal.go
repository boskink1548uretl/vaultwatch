package monitor

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// UnsealClient defines the vault operations needed by UnsealWatcher.
type UnsealClient interface {
	GetSealStatus(ctx context.Context) (*vault.SealStatus, error)
}

// UnsealWatcher monitors whether Vault is sealed and logs warnings.
type UnsealWatcher struct {
	client UnsealClient
	logger *log.Logger
}

// NewUnsealWatcher creates a watcher that alerts when Vault becomes sealed.
func NewUnsealWatcher(client UnsealClient, logger *log.Logger) *UnsealWatcher {
	if logger == nil {
		logger = log.Default()
	}
	return &UnsealWatcher{client: client, logger: logger}
}

// Check queries Vault seal status and logs a warning if sealed.
// Returns an error only on client failure; a sealed vault is a warning, not an error.
func (w *UnsealWatcher) Check(ctx context.Context) error {
	status, err := w.client.GetSealStatus(ctx)
	if err != nil {
		return fmt.Errorf("unseal watcher: fetching seal status: %w", err)
	}

	if status.Sealed {
		w.logger.Printf("[CRITICAL] Vault is sealed — intervention required (type=%s)", status.Type)
		return nil
	}

	w.logger.Printf("[OK] Vault is unsealed (type=%s)", status.Type)
	return nil
}
