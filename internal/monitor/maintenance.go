package monitor

import (
	"context"
	"io"
	"log"
	"os"
)

// maintenanceClient is the interface required by MaintenanceWatcher.
type maintenanceClient interface {
	IsMaintenanceModeEnabled(ctx context.Context) (bool, error)
}

// MaintenanceWatcher checks whether Vault is currently in maintenance mode
// and logs a warning when it is.
type MaintenanceWatcher struct {
	client maintenanceClient
	logger *log.Logger
}

// NewMaintenanceWatcher returns a MaintenanceWatcher using the provided client.
// If logger is nil a default stderr logger is used.
func NewMaintenanceWatcher(client maintenanceClient, logger *log.Logger) *MaintenanceWatcher {
	if logger == nil {
		logger = log.New(os.Stderr, "[maintenance] ", log.LstdFlags)
	}
	return &MaintenanceWatcher{client: client, logger: logger}
}

// Check queries Vault for maintenance mode status and logs accordingly.
// It returns true when Vault is in maintenance mode so callers can act on it.
func (w *MaintenanceWatcher) Check(ctx context.Context) (bool, error) {
	enabled, err := w.client.IsMaintenanceModeEnabled(ctx)
	if err != nil {
		w.logger.Printf("ERROR checking maintenance mode: %v", err)
		return false, err
	}
	if enabled {
		w.logger.Println("WARNING vault is currently in maintenance mode — secret operations may be unavailable")
	} else {
		w.logger.Println("INFO vault maintenance mode is not active")
	}
	return enabled, nil
}

// newMaintenanceLogger is a helper used in tests to create a logger writing to w.
func newMaintenanceLogger(w io.Writer) *log.Logger {
	return log.New(w, "", 0)
}
