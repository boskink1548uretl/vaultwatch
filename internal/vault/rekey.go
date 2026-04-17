package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// RekeyStatus holds the current rekey operation state.
type RekeyStatus struct {
	Started        bool     `json:"started"`
	T              int      `json:"t"`
	N              int      `json:"n"`
	Progress       int      `json:"progress"`
	Required       int      `json:"required"`
	PGPFingerprints []string `json:"pgp_fingerprints"`
	Backup         bool     `json:"backup"`
}

// GetRekeyStatus returns the current rekey status from Vault.
func (c *Client) GetRekeyStatus(ctx context.Context) (*RekeyStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/sys/rekey/init", nil)
	if err != nil {
		return nil, fmt.Errorf("rekey status request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rekey status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &RekeyStatus{Started: false}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rekey status: unexpected status %d", resp.StatusCode)
	}

	var result RekeyStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("rekey status decode: %w", err)
	}
	return &result, nil
}

// IsRekeyInProgress returns true when a rekey operation is underway.
func (c *Client) IsRekeyInProgress(ctx context.Context) (bool, error) {
	status, err := c.GetRekeyStatus(ctx)
	if err != nil {
		return false, err
	}
	return status.Started, nil
}
