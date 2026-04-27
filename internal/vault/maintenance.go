package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// MaintenanceStatus represents the current maintenance mode state of Vault.
type MaintenanceStatus struct {
	Enabled        bool   `json:"enabled"`
	Message        string `json:"message"`
	ResponseCode   int    `json:"response_code"`
}

// GetMaintenanceStatus returns the current maintenance mode status from Vault.
func (c *Client) GetMaintenanceStatus(ctx context.Context) (*MaintenanceStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/sys/maintenance", nil)
	if err != nil {
		return nil, fmt.Errorf("building maintenance status request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("maintenance status request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Maintenance mode not supported or not enabled — treat as disabled.
		return &MaintenanceStatus{Enabled: false}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status from maintenance endpoint: %d", resp.StatusCode)
	}

	var status MaintenanceStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding maintenance status response: %w", err)
	}
	return &status, nil
}

// IsMaintenanceModeEnabled is a convenience wrapper that returns true when
// Vault is currently in maintenance mode.
func (c *Client) IsMaintenanceModeEnabled(ctx context.Context) (bool, error) {
	status, err := c.GetMaintenanceStatus(ctx)
	if err != nil {
		return false, err
	}
	return status.Enabled, nil
}
