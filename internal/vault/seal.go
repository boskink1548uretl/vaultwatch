package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SealStatus holds the current seal state of the Vault cluster.
type SealStatus struct {
	Sealed      bool   `json:"sealed"`
	Initialized bool   `json:"initialized"`
	ClusterName string `json:"cluster_name"`
	Version     string `json:"version"`
	Progress    int    `json:"progress"`
	Threshold   int    `json:"t"`
	Shares      int    `json:"n"`
}

// GetSealStatus returns the current seal status from Vault.
func (c *Client) GetSealStatus(ctx context.Context) (*SealStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+"/v1/sys/seal-status", nil)
	if err != nil {
		return nil, fmt.Errorf("building seal-status request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("seal-status request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from seal-status", resp.StatusCode)
	}

	var status SealStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding seal-status response: %w", err)
	}
	return &status, nil
}

// IsSealed returns true if Vault is currently sealed.
func (c *Client) IsSealed(ctx context.Context) (bool, error) {
	status, err := c.GetSealStatus(ctx)
	if err != nil {
		return false, err
	}
	return status.Sealed, nil
}
