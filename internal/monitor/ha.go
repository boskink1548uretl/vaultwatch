package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// HAClient describes the Vault HA methods required by HAWatcher.
type HAClient interface {
	GetHAStatus(ctx context.Context) (*vault.HAStatus, error)
}

// HAWatcher monitors the high-availability status of a Vault cluster.
type HAWatcher struct {
	client    HAClient
	logger    *log.Logger
	minNodes  int
}

// HAWatcherOption configures an HAWatcher.
type HAWatcherOption func(*HAWatcher)

// WithHALogger sets a custom logger on the HAWatcher.
func WithHALogger(w io.Writer) HAWatcherOption {
	return func(h *HAWatcher) {
		h.logger = log.New(w, "[ha] ", 0)
	}
}

// WithMinNodes sets the minimum expected number of HA nodes.
func WithMinNodes(n int) HAWatcherOption {
	return func(h *HAWatcher) {
		h.minNodes = n
	}
}

// NewHAWatcher creates an HAWatcher with sensible defaults.
func NewHAWatcher(client HAClient, opts ...HAWatcherOption) *HAWatcher {
	h := &HAWatcher{
		client:   client,
		logger:   log.New(os.Stdout, "[ha] ", 0),
		minNodes: 1,
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

// Check fetches HA status and logs warnings for degraded cluster conditions.
func (h *HAWatcher) Check(ctx context.Context) error {
	status, err := h.client.GetHAStatus(ctx)
	if err != nil {
		return fmt.Errorf("ha watcher: %w", err)
	}
	if status == nil {
		h.logger.Println("warning: HA status unavailable (non-HA or standby node)")
		return nil
	}

	h.logger.Printf("cluster=%s leader=%s is_leader=%v nodes=%d",
		status.ClusterName, status.LeaderAddr, status.IsLeader, len(status.Nodes))

	if len(status.Nodes) < h.minNodes {
		h.logger.Printf("warning: cluster has %d node(s), expected at least %d",
			len(status.Nodes), h.minNodes)
	}

	for _, node := range status.Nodes {
		if node.ActiveNode {
			h.logger.Printf("active node: hostname=%s api=%s", node.Hostname, node.APIAddr)
		}
	}
	return nil
}
