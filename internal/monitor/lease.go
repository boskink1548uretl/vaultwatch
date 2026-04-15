package monitor

import (
	"context"
	"fmt"
	"log"
	"time"
)

// LeaseRenewer defines the interface for renewing Vault leases.
type LeaseRenewer interface {
	RenewLease(ctx context.Context, leaseID string, increment time.Duration) (interface{ GetDuration() time.Duration }, error)
}

// LeaseEntry tracks a dynamic secret lease that should be monitored and renewed.
type LeaseEntry struct {
	LeaseID       string
	Path          string
	RenewIncrement time.Duration
	ExpireTime    time.Time
	RenewThreshold time.Duration // renew when this much time remains
}

// LeaseManager monitors and auto-renews tracked leases.
type LeaseManager struct {
	entries []LeaseEntry
	renewer interface {
		RenewLease(ctx context.Context, leaseID string, increment time.Duration) (expireTime time.Time, err error)
	}
}

// RawLeaseRenewer is the concrete renewer interface used by LeaseManager.
type RawLeaseRenewer interface {
	RenewLease(ctx context.Context, leaseID string, increment time.Duration) (expireTime time.Time, err error)
}

// NewLeaseManager creates a LeaseManager with the given renewer.
func NewLeaseManager(renewer RawLeaseRenewer) *LeaseManager {
	return &LeaseManager{renewer: renewer}
}

// Track adds a lease to be monitored.
func (lm *LeaseManager) Track(entry LeaseEntry) {
	if entry.RenewThreshold == 0 {
		entry.RenewThreshold = 5 * time.Minute
	}
	if entry.RenewIncrement == 0 {
		entry.RenewIncrement = time.Hour
	}
	lm.entries = append(lm.entries, entry)
}

// CheckAndRenew iterates tracked leases and renews any that are within their threshold.
func (lm *LeaseManager) CheckAndRenew(ctx context.Context) []error {
	var errs []error
	for i, e := range lm.entries {
		remaining := time.Until(e.ExpireTime)
		if remaining > e.RenewThreshold {
			continue
		}
		log.Printf("[lease] renewing %q (expires in %v)", e.LeaseID, remaining.Round(time.Second))
		newExpiry, err := lm.renewer.RenewLease(ctx, e.LeaseID, e.RenewIncrement)
		if err != nil {
			errs = append(errs, fmt.Errorf("renew %q: %w", e.LeaseID, err))
			continue
		}
		lm.entries[i].ExpireTime = newExpiry
		log.Printf("[lease] renewed %q, new expiry %v", e.LeaseID, newExpiry.Format(time.RFC3339))
	}
	return errs
}
