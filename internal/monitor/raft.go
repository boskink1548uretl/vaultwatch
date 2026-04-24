package monitor

import (
	"context"
	"fmt"
	"log"
	"time"
)

// RaftClient defines the Vault operations required by RaftWatcher.
type RaftClient interface {
	GetRaftConfiguration(ctx context.Context) (interface{ PeerCount() int; HasLeader() bool }, error)
}

// raftConfigAccessor wraps the concrete vault type to satisfy RaftClient.
// In practice callers pass *vault.Client directly via the concrete adapter below.

// RaftWatcher monitors the Raft cluster for unhealthy conditions such as
// missing leaders or low voter counts.
type RaftWatcher struct {
	client      raftConfigProvider
	minVoters   int
	logger      *log.Logger
	lastChecked time.Time
}

type raftConfigProvider interface {
	GetRaftConfiguration(ctx context.Context) (raftCfg, error)
}

type raftCfg interface {
	LeaderCount() int
	VoterCount() int
}

// NewRaftWatcher creates a RaftWatcher with a minimum expected voter count.
func NewRaftWatcher(client raftConfigProvider, minVoters int, logger *log.Logger) *RaftWatcher {
	if minVoters <= 0 {
		minVoters = 3
	}
	return &RaftWatcher{
		client:    client,
		minVoters: minVoters,
		logger:    logger,
	}
}

// Check evaluates the current Raft cluster health and logs warnings.
func (w *RaftWatcher) Check(ctx context.Context) error {
	cfg, err := w.client.GetRaftConfiguration(ctx)
	if err != nil {
		return fmt.Errorf("raft watcher: fetching configuration: %w", err)
	}
	w.lastChecked = time.Now()

	if cfg.LeaderCount() == 0 {
		w.logger.Println("[CRITICAL] raft cluster has no leader")
	}

	voters := cfg.VoterCount()
	if voters < w.minVoters {
		w.logger.Printf("[WARNING] raft voter count %d is below minimum %d", voters, w.minVoters)
	}

	return nil
}
