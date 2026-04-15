package vault

import (
	"context"
	"fmt"
	"time"
)

// RenewResult holds the outcome of a secret renewal attempt.
type RenewResult struct {
	Path      string
	RenewedAt time.Time
	NewTTL    time.Duration
	Err       error
}

// RenewSecret attempts to renew the lease for a KV v2 secret at the given path.
// It writes a new version with the same data, effectively resetting the
// custom_metadata expiry timestamp when managed externally.
//
// For dynamic secrets with a lease ID, the Vault lease-renew endpoint is used.
func (c *Client) RenewSecret(ctx context.Context, path string) (*RenewResult, error) {
	meta, err := c.GetSecretMetadata(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("renew: fetch metadata for %q: %w", path, err)
	}

	if meta.LeaseID != "" {
		return c.renewLease(ctx, path, meta.LeaseID)
	}

	return c.renewKVVersion(ctx, path)
}

// renewLease calls the sys/leases/renew endpoint for dynamic secrets.
func (c *Client) renewLease(ctx context.Context, path, leaseID string) (*RenewResult, error) {
	body := map[string]interface{}{
		"lease_id": leaseID,
	}

	resp, err := c.logical.WriteWithContext(ctx, "sys/leases/renew", body)
	if err != nil {
		return nil, fmt.Errorf("renew lease for %q: %w", path, err)
	}
	if resp == nil {
		return nil, fmt.Errorf("renew lease for %q: empty response", path)
	}

	ttl := time.Duration(resp.LeaseDuration) * time.Second
	return &RenewResult{
		Path:      path,
		RenewedAt: time.Now().UTC(),
		NewTTL:    ttl,
	}, nil
}

// renewKVVersion re-writes the latest version of a KV v2 secret in place,
// which bumps the updated_time and resets any TTL tracking convention.
func (c *Client) renewKVVersion(ctx context.Context, path string) (*RenewResult, error) {
	secret, err := c.logical.ReadWithContext(ctx, "secret/data/"+path)
	if err != nil {
		return nil, fmt.Errorf("renew kv: read %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("renew kv: secret %q not found", path)
	}

	data, ok := secret.Data["data"]
	if !ok {
		return nil, fmt.Errorf("renew kv: missing data field for %q", path)
	}

	_, err = c.logical.WriteWithContext(ctx, "secret/data/"+path, map[string]interface{}{
		"data": data,
	})
	if err != nil {
		return nil, fmt.Errorf("renew kv: write %q: %w", path, err)
	}

	return &RenewResult{
		Path:      path,
		RenewedAt: time.Now().UTC(),
		NewTTL:    0, // KV v2 has no inherent TTL
	}, nil
}
