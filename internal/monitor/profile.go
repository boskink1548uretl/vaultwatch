package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// ProfileClient is the interface for fetching Vault profile info.
type ProfileClient interface {
	GetProfile(ctx context.Context) (*vault.VaultProfile, error)
}

// ProfileWatcher checks and reports the Vault instance profile.
type ProfileWatcher struct {
	client ProfileClient
	logger *log.Logger
}

// NewProfileWatcher creates a ProfileWatcher with the given client.
func NewProfileWatcher(client ProfileClient, opts ...func(*ProfileWatcher)) *ProfileWatcher {
	w := &ProfileWatcher{
		client: client,
		logger: log.New(os.Stderr, "[profile] ", log.LstdFlags),
	}
	for _, o := range opts {
		o(w)
	}
	return w
}

// WithProfileLogger sets a custom logger on the ProfileWatcher.
func WithProfileLogger(l *log.Logger) func(*ProfileWatcher) {
	return func(w *ProfileWatcher) { w.logger = l }
}

// Check fetches the Vault profile and writes a formatted summary to w.
func (pw *ProfileWatcher) Check(ctx context.Context, w io.Writer) error {
	profile, err := pw.client.GetProfile(ctx)
	if err != nil {
		return fmt.Errorf("profile watcher: %w", err)
	}

	if profile.Sealed {
		pw.logger.Printf("WARNING: Vault is sealed (version=%s cluster=%s)", profile.Version, profile.ClusterName)
	} else {
		pw.logger.Printf("Vault healthy: version=%s cluster=%s ha=%v", profile.Version, profile.ClusterName, profile.HA)
	}

	_, err = fmt.Fprintf(w,
		"Vault Profile\n  Version:      %s\n  Cluster:      %s (%s)\n  HA Enabled:   %v\n  Sealed:       %v\n  Initialized:  %v\n",
		profile.Version, profile.ClusterName, profile.ClusterID,
		profile.HA, profile.Sealed, profile.Initialized,
	)
	return err
}
