package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// AuditDevice represents a Vault audit device configuration.
type AuditDevice struct {
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Path        string            `json:"path"`
	Options     map[string]string `json:"options"`
	Local       bool              `json:"local"`
}

// ListAuditDevices returns all enabled audit devices from Vault.
func (c *Client) ListAuditDevices(ctx context.Context) (map[string]*AuditDevice, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+"/v1/sys/audit", nil)
	if err != nil {
		return nil, fmt.Errorf("building audit list request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing audit devices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return map[string]*AuditDevice{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status listing audit devices: %d", resp.StatusCode)
	}

	var result map[string]*AuditDevice
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding audit devices response: %w", err)
	}
	return result, nil
}

// IsAuditEnabled returns true if at least one audit device is enabled.
func (c *Client) IsAuditEnabled(ctx context.Context) (bool, error) {
	devices, err := c.ListAuditDevices(ctx)
	if err != nil {
		return false, err
	}
	return len(devices) > 0, nil
}
