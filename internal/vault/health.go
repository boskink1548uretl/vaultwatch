package vault

import (
	"context"
	"fmt"
	"time"
)

// HealthStatus represents the result of a Vault health check.
type HealthStatus struct {
	Healthy     bool
	Initialized bool
	Sealed      bool
	Version     string
	CheckedAt   time.Time
}

// String returns a human-readable summary of the health status.
func (h HealthStatus) String() string {
	if !h.Healthy {
		if h.Sealed {
			return fmt.Sprintf("vault is sealed (version %s, checked at %s)", h.Version, h.CheckedAt.Format(time.RFC3339))
		}
		if !h.Initialized {
			return fmt.Sprintf("vault is not initialized (checked at %s)", h.CheckedAt.Format(time.RFC3339))
		}
		return fmt.Sprintf("vault is unhealthy (checked at %s)", h.CheckedAt.Format(time.RFC3339))
	}
	return fmt.Sprintf("vault is healthy (version %s, checked at %s)", h.Version, h.CheckedAt.Format(time.RFC3339))
}

// CheckHealth queries the Vault health endpoint and returns a HealthStatus.
func (c *Client) CheckHealth(ctx context.Context) (HealthStatus, error) {
	health, err := c.vault.Sys().HealthWithContext(ctx)
	if err != nil {
		return HealthStatus{
			Healthy:   false,
			CheckedAt: time.Now(),
		}, fmt.Errorf("vault health check failed: %w", err)
	}

	return HealthStatus{
		Healthy:     health.Initialized && !health.Sealed,
		Initialized: health.Initialized,
		Sealed:      health.Sealed,
		Version:     health.Version,
		CheckedAt:   time.Now(),
	}, nil
}
