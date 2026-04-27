package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// InitStatus represents the initialization state of a Vault cluster.
type InitStatus struct {
	Initialized bool `json:"initialized"`
}

// InitConfig holds parameters for initializing a Vault cluster.
type InitConfig struct {
	SecretShares    int `json:"secret_shares"`
	SecretThreshold int `json:"secret_threshold"`
}

// InitResult holds the keys and root token returned after initialization.
type InitResult struct {
	Keys       []string `json:"keys"`
	KeysBase64 []string `json:"keys_base64"`
	RootToken  string   `json:"root_token"`
}

// GetInitStatus returns whether the Vault instance has been initialized.
func (c *Client) GetInitStatus(ctx context.Context) (*InitStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+"/v1/sys/init", nil)
	if err != nil {
		return nil, fmt.Errorf("vault: build init status request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vault: get init status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vault: get init status: unexpected status %d", resp.StatusCode)
	}

	var status InitStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("vault: decode init status: %w", err)
	}

	return &status, nil
}

// IsInitialized is a convenience wrapper that returns true when Vault has been initialized.
func (c *Client) IsInitialized(ctx context.Context) (bool, error) {
	status, err := c.GetInitStatus(ctx)
	if err != nil {
		return false, err
	}
	return status.Initialized, nil
}
