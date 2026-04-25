package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

// AgentClient is the interface for retrieving Vault Agent information.
type AgentClient interface {
	IsAgentCacheEnabled(ctx context.Context) (bool, error)
}

// AgentWatcher monitors the state of a Vault Agent process.
type AgentWatcher struct {
	client      AgentClient
	requireCache bool
	logger      *log.Logger
}

// AgentWatcherOption configures an AgentWatcher.
type AgentWatcherOption func(*AgentWatcher)

// WithRequireCache configures the watcher to warn when agent cache is disabled.
func WithRequireCache(require bool) AgentWatcherOption {
	return func(w *AgentWatcher) {
		w.requireCache = require
	}
}

// WithAgentLogger sets a custom logger on the AgentWatcher.
func WithAgentLogger(l *log.Logger) AgentWatcherOption {
	return func(w *AgentWatcher) {
		w.logger = l
	}
}

// NewAgentWatcher creates a new AgentWatcher with the given client and options.
func NewAgentWatcher(client AgentClient, opts ...AgentWatcherOption) *AgentWatcher {
	w := &AgentWatcher{
		client:      client,
		requireCache: false,
		logger:      log.New(os.Stderr, "[agent-watcher] ", log.LstdFlags),
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// Check inspects the Vault Agent state and logs warnings as appropriate.
// It returns a non-nil error only if the agent endpoint is unreachable.
func (w *AgentWatcher) Check(ctx context.Context) error {
	enabled, err := w.client.IsAgentCacheEnabled(ctx)
	if err != nil {
		w.logger.Printf("ERROR: could not reach Vault Agent: %v", err)
		return fmt.Errorf("agent check failed: %w", err)
	}

	if w.requireCache && !enabled {
		w.logger.Println("WARNING: Vault Agent cache is disabled but requireCache is set")
		return nil
	}

	if enabled {
		w.logger.Println("INFO: Vault Agent cache is enabled")
	} else {
		w.logger.Println("INFO: Vault Agent cache is disabled")
	}
	return nil
}

// WriteReport writes a human-readable agent status summary to the provided writer.
func (w *AgentWatcher) WriteReport(ctx context.Context, out io.Writer) error {
	enabled, err := w.client.IsAgentCacheEnabled(ctx)
	if err != nil {
		return fmt.Errorf("fetching agent info for report: %w", err)
	}
	cacheStatus := "disabled"
	if enabled {
		cacheStatus = "enabled"
	}
	_, werr := fmt.Fprintf(out, "Vault Agent Status\n  Cache: %s\n", cacheStatus)
	return werr
}
