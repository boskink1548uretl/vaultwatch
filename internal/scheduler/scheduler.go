// Package scheduler provides periodic execution of vault secret checks.
package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/monitor"
	"github.com/yourusername/vaultwatch/internal/vault"
)

// Scheduler periodically checks vault secrets and dispatches alerts.
type Scheduler struct {
	client   *vault.Client
	checker  *monitor.Checker
	notifier *alert.StdoutNotifier
	interval time.Duration
	paths    []string
}

// New creates a new Scheduler with the given dependencies and polling interval.
func New(
	client *vault.Client,
	checker *monitor.Checker,
	notifier *alert.StdoutNotifier,
	interval time.Duration,
	paths []string,
) *Scheduler {
	return &Scheduler{
		client:   client,
		checker:  checker,
		notifier: notifier,
		interval: interval,
		paths:    paths,
	}
}

// Run starts the polling loop, running until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) error {
	log.Printf("scheduler: starting with interval=%s, paths=%v", s.interval, s.paths)
	s.tick(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.tick(ctx)
		case <-ctx.Done():
			log.Println("scheduler: shutting down")
			return ctx.Err()
		}
	}
}

// tick performs a single round of checks across all configured paths.
func (s *Scheduler) tick(ctx context.Context) {
	for _, path := range s.paths {
		meta, err := s.client.GetSecretMetadata(ctx, path)
		if err != nil {
			log.Printf("scheduler: error fetching metadata for %q: %v", path, err)
			continue
		}

		result := s.checker.Evaluate(path, meta.ExpiresAt)
		if err := s.notifier.Notify(result); err != nil {
			log.Printf("scheduler: error sending notification for %q: %v", path, err)
		}
	}
}
