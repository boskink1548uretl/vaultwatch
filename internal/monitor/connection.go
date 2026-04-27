package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/your-org/vaultwatch/internal/vault"
)

// ConnectionWatcher checks Vault connectivity and logs diagnostics.
type ConnectionWatcher struct {
	client    *vault.Client
	logger    *log.Logger
	threshold time.Duration
}

// NewConnectionWatcher creates a ConnectionWatcher with optional logger.
// threshold is the maximum acceptable latency before a warning is logged.
func NewConnectionWatcher(client *vault.Client, threshold time.Duration, w io.Writer) *ConnectionWatcher {
	if w == nil {
		w = os.Stdout
	}
	if threshold <= 0 {
		threshold = 500 * time.Millisecond
	}
	return &ConnectionWatcher{
		client:    client,
		logger:    log.New(w, "[connection] ", 0),
		threshold: threshold,
	}
}

// Check probes Vault and logs the connection state.
// Returns an error if Vault is unreachable.
func (cw *ConnectionWatcher) Check(ctx context.Context) error {
	info, err := cw.client.CheckConnection(ctx)
	if err != nil {
		return fmt.Errorf("connection check failed: %w", err)
	}

	if !info.Reachable {
		cw.logger.Printf("UNREACHABLE addr=%s tls=%v", info.Address, info.TLSEnabled)
		return fmt.Errorf("vault unreachable at %s", info.Address)
	}

	if info.Latency > cw.threshold {
		cw.logger.Printf("HIGH_LATENCY addr=%s latency=%s threshold=%s status=%d",
			info.Address, info.Latency.Round(time.Millisecond),
			cw.threshold.Round(time.Millisecond), info.StatusCode)
		return nil
	}

	cw.logger.Printf("OK addr=%s latency=%s tls=%v status=%d",
		info.Address, info.Latency.Round(time.Millisecond),
		info.TLSEnabled, info.StatusCode)
	return nil
}
