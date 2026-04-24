package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// SysWatcher checks Vault system-level health and logs notable conditions.
type SysWatcher struct {
	client *vault.Client
	logger *log.Logger
}

// NewSysWatcher creates a SysWatcher using the provided Vault client.
// If logger is nil, output goes to stderr.
func NewSysWatcher(client *vault.Client, logger *log.Logger) *SysWatcher {
	if logger == nil {
		logger = log.New(os.Stderr, "[sys] ", log.LstdFlags)
	}
	return &SysWatcher{client: client, logger: logger}
}

// Check fetches SysInfo and emits warnings for concerning states.
// It returns the SysInfo so callers can include it in reports.
func (w *SysWatcher) Check(ctx context.Context) (*vault.SysInfo, error) {
	info, err := w.client.GetSysInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("sys watcher: %w", err)
	}

	if info.Sealed {
		w.logger.Printf("CRITICAL vault is sealed (cluster=%s)", info.ClusterName)
	}
	if !info.Initialized {
		w.logger.Printf("CRITICAL vault is not initialized")
	}
	if info.Standby {
		w.logger.Printf("WARNING vault node is in standby mode (cluster=%s)", info.ClusterName)
	}
	if info.ReplicationDR.Mode == "secondary" {
		w.logger.Printf("INFO vault is a DR secondary (cluster=%s)", info.ClusterName)
	}

	return info, nil
}

// WriteReport writes a human-readable summary of SysInfo to w.
func WriteReport(w io.Writer, info *vault.SysInfo) {
	fmt.Fprintf(w, "Vault System Info\n")
	fmt.Fprintf(w, "  Version     : %s\n", info.Version)
	fmt.Fprintf(w, "  Cluster     : %s (%s)\n", info.ClusterName, info.ClusterID)
	fmt.Fprintf(w, "  Initialized : %v\n", info.Initialized)
	fmt.Fprintf(w, "  Sealed      : %v\n", info.Sealed)
	fmt.Fprintf(w, "  HA Enabled  : %v\n", info.HA)
	fmt.Fprintf(w, "  Standby     : %v\n", info.Standby)
}
