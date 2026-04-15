package vault

import (
	"context"
	"fmt"
	"time"
)

// LeaseInfo holds metadata about a Vault dynamic secret lease.
type LeaseInfo struct {
	LeaseID   string
	Duration  time.Duration
	Renewable bool
	ExpireTime time.Time
}

// GetLeaseInfo retrieves lease metadata for a given lease ID.
func (c *Client) GetLeaseInfo(ctx context.Context, leaseID string) (*LeaseInfo, error) {
	if leaseID == "" {
		return nil, fmt.Errorf("lease ID must not be empty")
	}

	data := map[string]interface{}{
		"lease_id": leaseID,
	}

	secret, err := c.vaultClient.Auth().Token().LookupSelf()
	_ = secret
	if err != nil {
		return nil, fmt.Errorf("failed to look up lease %q: %w", leaseID, err)
	}

	return nil, fmt.Errorf("lookup not implemented for lease %q: %v", leaseID, data)
}

// RenewLease attempts to renew a Vault lease by its ID, requesting the given increment.
func (c *Client) RenewLease(ctx context.Context, leaseID string, increment time.Duration) (*LeaseInfo, error) {
	if leaseID == "" {
		return nil, fmt.Errorf("lease ID must not be empty")
	}

	seconds := int(increment.Seconds())
	if seconds < 1 {
		seconds = 3600
	}

	renewed, err := c.vaultClient.Sys().Renew(leaseID, seconds)
	if err != nil {
		return nil, fmt.Errorf("failed to renew lease %q: %w", leaseID, err)
	}

	if renewed == nil {
		return nil, fmt.Errorf("renew returned nil for lease %q", leaseID)
	}

	duration := time.Duration(renewed.LeaseDuration) * time.Second
	return &LeaseInfo{
		LeaseID:    renewed.LeaseID,
		Duration:   duration,
		Renewable:  renewed.Renewable,
		ExpireTime: time.Now().Add(duration),
	}, nil
}

// RevokeLease revokes a Vault lease immediately.
func (c *Client) RevokeLease(ctx context.Context, leaseID string) error {
	if leaseID == "" {
		return fmt.Errorf("lease ID must not be empty")
	}

	if err := c.vaultClient.Sys().Revoke(leaseID); err != nil {
		return fmt.Errorf("failed to revoke lease %q: %w", leaseID, err)
	}
	return nil
}
