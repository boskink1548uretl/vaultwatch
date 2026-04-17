package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

// RekeyClient is the interface for rekey status checks.
type RekeyClient interface {
	IsRekeyInProgress(ctx context.Context) (bool, error)
}

// RekeyWatcher monitors whether a Vault rekey operation is in progress.
type RekeyWatcher struct {
	client RekeyClient
	logger *log.Logger
}

// NewRekeyWatcher creates a RekeyWatcher with the given client.
func NewRekeyWatcher(client RekeyClient, out io.Writer) *RekeyWatcher {
	if out == nil {
		out = os.Stdout
	}
	return &RekeyWatcher{
		client: client,
		logger: log.New(out, "[rekey] ", 0),
	}
}

// Check queries Vault for an active rekey operation and logs a warning if found.
func (rw *RekeyWatcher) Check(ctx context.Context) error {
	inProgress, err := rw.client.IsRekeyInProgress(ctx)
	if err != nil {
		return fmt.Errorf("rekey check failed: %w", err)
	}
	if inProgress {
		rw.logger.Println("WARNING: a rekey operation is currently in progress")
	}
	return nil
}
