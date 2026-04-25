package monitor

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/vaultwatch/internal/vault"
)

// QuotaClient is the subset of vault.Client used by QuotaWatcher.
type QuotaClient interface {
	ListQuotas(ctx context.Context) ([]string, error)
	GetQuota(ctx context.Context, name string) (*vault.RateQuota, error)
}

// QuotaWatcher monitors Vault rate-limit quotas and logs rules with no
// block interval configured (which may allow burst abuse).
type QuotaWatcher struct {
	client           QuotaClient
	logger           *log.Logger
	warnIfNoBlock    bool
}

// NewQuotaWatcher creates a QuotaWatcher. When warnIfNoBlock is true, any
// quota rule with block_interval == 0 is logged as a warning.
func NewQuotaWatcher(client QuotaClient, logger *log.Logger, warnIfNoBlock bool) *QuotaWatcher {
	return &QuotaWatcher{
		client:        client,
		logger:        logger,
		warnIfNoBlock: warnIfNoBlock,
	}
}

// Check lists all quota rules and inspects each one, emitting log warnings
// for rules that may expose the cluster to abuse.
func (w *QuotaWatcher) Check(ctx context.Context) error {
	names, err := w.client.ListQuotas(ctx)
	if err != nil {
		return fmt.Errorf("quota watcher list: %w", err)
	}

	if len(names) == 0 {
		w.logger.Println("[quota] warning: no rate-limit quotas configured")
		return nil
	}

	var fetchErrors int
	for _, name := range names {
		q, err := w.client.GetQuota(ctx, name)
		if err != nil {
			w.logger.Printf("[quota] error fetching %s: %v", name, err)
			fetchErrors++
			continue
		}
		if q == nil {
			continue
		}
		if w.warnIfNoBlock && q.BlockInterval == 0 {
			w.logger.Printf("[quota] warning: rule %q (path=%q rate=%.0f) has no block_interval — burst requests not blocked",
				q.Name, q.Path, q.Rate)
		}
	}

	if fetchErrors > 0 {
		return fmt.Errorf("quota watcher: failed to fetch %d of %d quota rule(s)", fetchErrors, len(names))
	}
	return nil
}
