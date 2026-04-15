package monitor

import (
	"context"
	"fmt"
	"log"
	"time"
)

// TokenClient is the interface for retrieving Vault token information.
type TokenClient interface {
	IsTokenExpiringSoon(ctx context.Context, threshold time.Duration) (bool, error)
}

// TokenWatcherConfig holds configuration for the token watcher.
type TokenWatcherConfig struct {
	// WarnThreshold is the TTL duration below which a warning is emitted.
	WarnThreshold time.Duration
	// Logger receives warning messages; defaults to log.Printf if nil.
	Logger func(format string, args ...interface{})
}

// TokenWatcher monitors the Vault token TTL and emits warnings.
type TokenWatcher struct {
	client TokenClient
	cfg    TokenWatcherConfig
}

// NewTokenWatcher creates a TokenWatcher with the given client and config.
func NewTokenWatcher(client TokenClient, cfg TokenWatcherConfig) *TokenWatcher {
	if cfg.WarnThreshold == 0 {
		cfg.WarnThreshold = 30 * time.Minute
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Printf
	}
	return &TokenWatcher{client: client, cfg: cfg}
}

// Check evaluates the current token TTL and logs a warning if it is expiring soon.
// Returns true if the token is expiring soon, false otherwise.
func (w *TokenWatcher) Check(ctx context.Context) (bool, error) {
	expiring, err := w.client.IsTokenExpiringSoon(ctx, w.cfg.WarnThreshold)
	if err != nil {
		return false, fmt.Errorf("token watcher check failed: %w", err)
	}
	if expiring {
		w.cfg.Logger("[vaultwatch] WARNING: Vault token is expiring within %s — consider renewing or re-authenticating", w.cfg.WarnThreshold)
	}
	return expiring, nil
}
