package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Snapshotter is satisfied by vault.Client.
type Snapshotter interface {
	TakeSnapshot(ctx context.Context, w io.Writer) (interface{ GetSize() int64 }, error)
}

// SnapshotManager periodically writes Vault Raft snapshots to disk.
type SnapshotManager struct {
	client   snapshotClient
	outDir   string
	interval time.Duration
	log      *log.Logger
}

type snapshotClient interface {
	TakeSnapshot(ctx context.Context, w io.Writer) (interface{}, error)
}

// NewSnapshotManager creates a SnapshotManager that stores snapshots in outDir.
func NewSnapshotManager(client snapshotClient, outDir string, interval time.Duration) *SnapshotManager {
	return &SnapshotManager{
		client:   client,
		outDir:   outDir,
		interval: interval,
		log:      log.New(os.Stderr, "[snapshot] ", log.LstdFlags),
	}
}

// Run blocks, taking snapshots at each interval tick until ctx is cancelled.
func (m *SnapshotManager) Run(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			if err := m.take(ctx, t); err != nil {
				m.log.Printf("snapshot failed: %v", err)
			}
		}
	}
}

func (m *SnapshotManager) take(ctx context.Context, t time.Time) error {
	name := fmt.Sprintf("vault-snapshot-%s.snap", t.UTC().Format("20060102-150405"))
	path := filepath.Join(m.outDir, name)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file %s: %w", path, err)
	}
	defer f.Close()
	_, err = m.client.TakeSnapshot(ctx, f)
	if err != nil {
		os.Remove(path)
		return fmt.Errorf("take snapshot: %w", err)
	}
	m.log.Printf("snapshot saved: %s", path)
	return nil
}
